// SPDX-License-Identifier: ISC
// Copyright (c) 2019-2021 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/bitmark-inc/autonomy-pod-controller/key"
)

const BindingFile = "TEST_OWNER_BOUND"

type ControllerTestSuite struct {
	suite.Suite
	Identity *PodIdentity
}

func (suite *ControllerTestSuite) TestBind() {
	did := "did:key:zQ3shvD5cZSLggSCiu4jmF3jRY6GMUb7zvwChfhYQGJfQudJE"

	mockCtl := gomock.NewController(suite.T())
	defer mockCtl.Finish()
	mockedStore := NewMockStore(mockCtl)
	mockedStore.EXPECT().SetBinding(did, gomock.Any()).Times(1).Return(nil)

	c := Controller{
		Identity: suite.Identity,
		store:    mockedStore,
	}

	r, err := c.bind(did)
	suite.NoError(err)
	suite.Equal(r["identity"], suite.Identity.DID)
	suite.True(key.VerifySignature(r["identity"], r["nonce"]+r["timestamp"], r["signature"]))
}

func (suite *ControllerTestSuite) TestBindAckWithValidNonceAndSignature() {
	didWithValidNonceAndSignature := "did:key:zQ3shvD5cZSLggSCiu4jmF3jRY6GMUb7zvwChfhYQGJfQudJE"
	didWithWrongNonce := "did:key:zQ3shrG4MGtHFTq4BMaPtWRysMuTXVB5H2G4upbQzvk9PyANM"
	didWithInvalidSignature := "did:key:zQ3shvtd8SgRF7UBHzJsnH1Qu2MKgEMBRujEfr4wp5a171Vv4"

	mockCtl := gomock.NewController(suite.T())
	defer mockCtl.Finish()
	mockedStore := NewMockStore(mockCtl)
	mockedStore.EXPECT().BindingNonce(didWithValidNonceAndSignature).AnyTimes().Return("1eba606e")
	mockedStore.EXPECT().BindingNonce(didWithWrongNonce).AnyTimes().Return("1eba606a")
	mockedStore.EXPECT().BindingNonce(didWithInvalidSignature).AnyTimes().Return("1eba606c")
	mockedStore.EXPECT().CompleteBinding(didWithValidNonceAndSignature).Times(1).Return(nil)

	c := Controller{
		Identity: suite.Identity,
		store:    mockedStore,
	}

	testCases := []struct {
		did       string
		timestamp string
		signature string
		err       error
	}{
		{
			didWithValidNonceAndSignature,
			"1618456405107",
			"3045022100d500b7ebbadeed51aaff844a0e7d741eb5bbf4c14b8d8476d87fae4ae02ab08b0220787dcaeae59327d1ff17db5b25386bf3250425a702a212e4fbd470b890d45ea6",
			nil,
		},
		{
			didWithWrongNonce, // the signature is signed with 1eba606e + timestmap (not the tested nonce 1eba606a)
			"1618456405107",
			"3045022100d500b7ebbadeed51aaff844a0e7d741eb5bbf4c14b8d8476d87fae4ae02ab08b0220787dcaeae59327d1ff17db5b25386bf3250425a702a212e4fbd470b890d45ea8",
			errors.New("invalid binding ack signature"),
		},
		{
			didWithInvalidSignature,
			"1618456405107",
			"",
			errors.New("invalid binding ack signature"),
		},
	}

	for _, t := range testCases {
		resp, err := c.bindACK(t.did, BindACKParams{
			Timestamp: "1618456405107",
			Signature: "3045022100d500b7ebbadeed51aaff844a0e7d741eb5bbf4c14b8d8476d87fae4ae02ab08b0220787dcaeae59327d1ff17db5b25386bf3250425a702a212e4fbd470b890d45ea6",
		})

		if t.err == nil {
			suite.NoError(err)
			suite.Equal(resp["status"], "ok")
		} else {
			suite.Error(err, "invalid binding ack signature")
		}
	}
}

func (suite *ControllerTestSuite) TestAccessMode() {
	ownerDID := "did:key:zQ3shvD5cZSLggSCiu4jmF3jRY6GMUb7zvwChfhYQGJfQudJE"
	memberDID := "did:key:zQ3shrG4MGtHFTq4BMaPtWRysMuTXVB5H2G4upbQzvk9PyANM"
	memberWithInvalidAccessModeDID := "did:key:zQ3shvtd8SgRF7UBHzJsnH1Qu2MKgEMBRujEfr4wp5a171Vv4"

	mockCtl := gomock.NewController(suite.T())
	defer mockCtl.Finish()
	mockedStore := NewMockStore(mockCtl)
	mockedStore.EXPECT().MemberAccessMode(memberDID).AnyTimes().Return(AccessModeLimited)
	mockedStore.EXPECT().MemberAccessMode(memberWithInvalidAccessModeDID).AnyTimes().Return(AccessMode(10))

	c := Controller{ownerDID: ownerDID, store: mockedStore}

	suite.Equal(AccessModeFull, c.accessMode(ownerDID))
	suite.Equal(AccessModeLimited, c.accessMode(memberDID))
	suite.Equal(AccessModeNotApplicant, c.accessMode(memberWithInvalidAccessModeDID))
}

func (suite *ControllerTestSuite) TestHasCorrectBindingState() {
	didWithBinding := "did:key:zQ3shvD5cZSLggSCiu4jmF3jRY6GMUb7zvwChfhYQGJfQudJE"
	didWithoutBinding := "did:key:zQ3shrG4MGtHFTq4BMaPtWRysMuTXVB5H2G4upbQzvk9PyANM"

	mockCtl := gomock.NewController(suite.T())
	defer mockCtl.Finish()
	mockedStore := NewMockStore(mockCtl)
	mockedStore.EXPECT().HasBinding(didWithBinding).AnyTimes().Return(true)
	mockedStore.EXPECT().HasBinding(didWithoutBinding).AnyTimes().Return(false)

	c := Controller{store: mockedStore}

	suite.False(c.hasCorrectBindingState(didWithBinding, "bind"))
	suite.False(c.hasCorrectBindingState(didWithBinding, "bind_ack"))
	suite.True(c.hasCorrectBindingState(didWithoutBinding, "bind"))
	suite.True(c.hasCorrectBindingState(didWithoutBinding, "bind_ack"))

	suite.True(c.hasCorrectBindingState(didWithBinding, "create_wallet"))
	suite.False(c.hasCorrectBindingState(didWithoutBinding, "create_wallet"))
}

func (suite *ControllerTestSuite) TestCreateWallet() {
	incompleteDescriptor := "wsh(sortedmulti(2,[119dbcab/48h/1h/0h/2h]tpubDFYr9xD4WtT3yDBdX2qT2j2v6ZruqccwPKFwLguuJL99bWBrk6D2Lv1aPpRbFnw1sQUU9DM7ScMAkPRJqR1iXKhWMBNMAJ45QCTuvSZbzzv/0/*,[e650dc93/48h/1h/0h/2h]tpubDEijNAeHVNmm6wHwspPv4fV8mRkoMimeVCk47dExpN9e17jFti12BdjzL8MX17GvKEekRzknNuDoLy1Q8fujYfsWfCvjwYmjjENUpzwDy6B/0/*,[<fingerprint>/48h/1h/0h/2h]<xpub>/0/*))"
	tempMasterKey := "tprv8ZgxMBicQKsPdbrPFhUK3QmhRsWiRYrCPuftQDnpGS1pDMhvUhVrLQvH8yxLTS7P5DCRwn55f9xHh8x97extRmrN5BGqYxuAtGbeKYn78MH"
	tempKeyFilePath := "/tmp/test-master-key"
	x, _ := os.Create(tempKeyFilePath)
	x.WriteString(tempMasterKey)
	x.Close()
	masterFingerprint := "da8a7377"
	gordianPrivateKey := "tprv8iat9qorVppXiivghjfoNwgGfL4ejNZNWyURC78sBnM9Xur23w743PUJ2CN95gkC7THPrYKAhPCLQ1Qsf5Brq3HrqWCARCCWD6UZmFe1HA4"
	gordianPublicKey := "tpubDFGvJFr6eCWCcBxUbPLPnMLPEMaathkH6H5CUdBAc49YNQ6ngKveDt6ACL9FG7yVPCGPoejizvdYLumw4YmsyUMZfbD3xRyb3DAXd5y9NFr"

	importedDescriptorReplacer := strings.NewReplacer(
		"<fingerprint>", masterFingerprint,
		"<xpub>", gordianPrivateKey,
		"<tpub>", gordianPrivateKey,
	)

	externalDescriptorWithoutChecksum := importedDescriptorReplacer.Replace(incompleteDescriptor)
	internalDescriptorWithoutChecksum := strings.ReplaceAll(externalDescriptorWithoutChecksum, "/0/*", "/1/*")

	accountMapDescriptorReplacer := strings.NewReplacer(
		"<fingerprint>", masterFingerprint,
		"<xpub>", gordianPublicKey,
		"<tpub>", gordianPrivateKey,
	)

	gordianWalletDescriptor := accountMapDescriptorReplacer.Replace(incompleteDescriptor)

	mockCtl := gomock.NewController(suite.T())
	defer func() {
		os.Remove(tempKeyFilePath)
		mockCtl.Finish()
	}()

	// wallet exist but not import Descriptor
	mockedRPCClient := NewMockRPCClient(mockCtl)
	mockedRPCClient.EXPECT().GetWalletInfo().Times(1).Return(&btcjson.GetWalletInfoResult{KeyPoolSize: 0, KeyPoolSizeHDInternal: new(int)}, nil)
	mockedRPCClient.EXPECT().GetBlockChainInfo().Times(1).Return(&btcjson.GetBlockChainInfoResult{Chain: "test"}, nil)
	mockedRPCClient.EXPECT().GetDescriptorInfo(externalDescriptorWithoutChecksum).Times(1).Return(&btcjson.GetDescriptorInfoResult{Checksum: "a"}, nil)
	mockedRPCClient.EXPECT().GetDescriptorInfo(internalDescriptorWithoutChecksum).Times(1).Return(&btcjson.GetDescriptorInfoResult{Checksum: "b"}, nil)
	mockedRPCClient.EXPECT().RawRequest("importdescriptors", gomock.Any()).Times(1).Return(nil, nil)

	// wallet not exist
	mockedRPCClient2 := NewMockRPCClient(mockCtl)
	mockedRPCClient2.EXPECT().GetWalletInfo().Times(1).Return(nil, &btcjson.RPCError{Code: btcjson.ErrRPCWalletNotFound})
	mockedRPCClient2.EXPECT().GetBlockChainInfo().Times(1).Return(&btcjson.GetBlockChainInfoResult{Chain: "test"}, nil)
	mockedRPCClient2.EXPECT().RawRequest("createwallet", gomock.Any()).Times(1).Return(nil, nil)
	mockedRPCClient2.EXPECT().GetDescriptorInfo(externalDescriptorWithoutChecksum).Times(1).Return(&btcjson.GetDescriptorInfoResult{Checksum: "a"}, nil)
	mockedRPCClient2.EXPECT().GetDescriptorInfo(internalDescriptorWithoutChecksum).Times(1).Return(&btcjson.GetDescriptorInfoResult{Checksum: "b"}, nil)
	mockedRPCClient2.EXPECT().RawRequest("importdescriptors", gomock.Any()).Times(1).Return(nil, nil)

	c := Controller{}

	result, err := c.createWallet(mockedRPCClient, tempKeyFilePath, incompleteDescriptor)
	suite.NoError(err)
	suite.Equal(map[string]string{"descriptor": gordianWalletDescriptor}, result)
	result, err = c.createWallet(mockedRPCClient2, tempKeyFilePath, incompleteDescriptor)
	suite.NoError(err)
	suite.Equal(map[string]string{"descriptor": gordianWalletDescriptor}, result)
}

func (suite *ControllerTestSuite) TestFinishPSBT() {
	type finalizePSBTResult struct {
		PSBT     string `json:"psbt"`
		Hex      string `json:"hex"`
		Complete bool   `json:"complete"`
	}
	psbtBytes, _ := json.Marshal("ProcessPsbt")
	PSBTResult := finalizePSBTResult{
		PSBT:     "psbt",
		Hex:      "hex",
		Complete: true,
	}
	PSBTResultBytes, _ := json.Marshal(PSBTResult)
	hexBytes, _ := json.Marshal(PSBTResult.Hex)
	txID, _ := json.Marshal("txID")

	mockCtl := gomock.NewController(suite.T())
	defer func() {
		mockCtl.Finish()
	}()

	mockedRPCClient := NewMockRPCClient(mockCtl)
	mockedRPCClient.EXPECT().WalletProcessPsbt("PSBT", btcjson.Bool(true), rpcclient.SigHashAll, btcjson.Bool(true)).Times(1).Return(&btcjson.WalletProcessPsbtResult{Complete: true, Psbt: "ProcessPsbt"}, nil)
	mockedRPCClient.EXPECT().RawRequest("finalizepsbt", []json.RawMessage{psbtBytes}).Times(1).Return(PSBTResultBytes, nil)
	mockedRPCClient.EXPECT().RawRequest("sendrawtransaction", []json.RawMessage{hexBytes}).Times(1).Return(txID, nil)

	c := Controller{}

	result, err := c.finishPSBT(mockedRPCClient, "PSBT")
	suite.NoError(err)
	suite.Equal(map[string]string{"txid": "txID"}, result)
}

func TestControllerTestSuite(t *testing.T) {
	i, err := NewPodIdentity()
	if err != nil {
		t.Fatal("unable to create test key file")
	}
	suite.Run(t, &ControllerTestSuite{
		Identity: i,
	})
}
