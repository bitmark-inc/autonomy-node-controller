package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/bitmark-inc/autonomy-pod-controller/key"
)

const BindingFile = "TEST_OWNER_BOUND"

type ControllerTestSuite struct {
	suite.Suite
	Identity *PodIdentity
}

func (suite *ControllerTestSuite) TestBindWithOwnerDID() {
	ownerDID := "did:key:zQ3shvD5cZSLggSCiu4jmF3jRY6GMUb7zvwChfhYQGJfQudJE"

	c := Controller{
		bindingFile: BindingFile,
		ownerDID:    ownerDID,
		Identity:    suite.Identity,
	}

	b := c.bind(ownerDID)
	resp := map[string]json.RawMessage{}
	suite.NoError(json.Unmarshal(b, &resp))
	suite.NotNil(resp["data"])

	result := map[string]string{}
	suite.NoError(json.Unmarshal(resp["data"], &result))
	suite.Equal(result["identity"], suite.Identity.DID)
	suite.True(key.VerifySignature(result["identity"], result["nonce"]+result["timestamp"], result["signature"]))
}

func (suite *ControllerTestSuite) TestBindWithNonOwnerDID() {
	ownerDID := "did:key:zQ3shvD5cZSLggSCiu4jmF3jRY6GMUb7zvwChfhYQGJfQudJE"
	nonOwnerDID := "did:key:zQ3shrG4MGtHFTq4BMaPtWRysMuTXVB5H2G4upbQzvk9PyANM"
	c := Controller{
		bindingFile: BindingFile,
		ownerDID:    ownerDID,
		Identity:    suite.Identity,
	}

	b := c.bind(nonOwnerDID)
	resp := map[string]json.RawMessage{}
	suite.NoError(json.Unmarshal(b, &resp))
	suite.Nil(resp["data"])
	suite.NotNil(resp["error"])

	var err string
	suite.NoError(json.Unmarshal(resp["error"], &err))
	suite.Equal(err, "illegal owner")
}

func (suite *ControllerTestSuite) TestBindWithBoundPod() {
	ownerDID := "did:key:zQ3shvD5cZSLggSCiu4jmF3jRY6GMUb7zvwChfhYQGJfQudJE"

	c := Controller{
		bindingFile: BindingFile,
		ownerDID:    ownerDID,
		Identity:    suite.Identity,
	}
	suite.NoError(c.BindAccount())
	defer os.Remove(BindingFile)

	b := c.bind(ownerDID)
	resp := map[string]json.RawMessage{}
	suite.NoError(json.Unmarshal(b, &resp))
	suite.Nil(resp["data"])
	suite.NotNil(resp["error"])

	var err string
	suite.NoError(json.Unmarshal(resp["error"], &err))
	suite.Equal(err, "node has bound")
}

func (suite *ControllerTestSuite) TestBindAckWithValidNonceAndSignature() {
	ownerDID := "did:key:zQ3shvD5cZSLggSCiu4jmF3jRY6GMUb7zvwChfhYQGJfQudJE"
	nonce := "1eba606e"

	c := Controller{
		bindingFile:  BindingFile,
		bindingNonce: nonce,
		ownerDID:     ownerDID,
		Identity:     suite.Identity,
	}

	b := c.bindACK(ownerDID, BindACKParams{
		Timestamp: "1618456405107",
		Signature: "3045022100d500b7ebbadeed51aaff844a0e7d741eb5bbf4c14b8d8476d87fae4ae02ab08b0220787dcaeae59327d1ff17db5b25386bf3250425a702a212e4fbd470b890d45ea6",
	})

	resp := map[string]json.RawMessage{}
	suite.NoError(json.Unmarshal(b, &resp))
	suite.NotNil(resp["data"])

	result := map[string]string{}
	suite.NoError(json.Unmarshal(resp["data"], &result))
	suite.Equal(result["status"], "ok")
}

func (suite *ControllerTestSuite) TestBindAckWithDifferentNonce() {
	ownerDID := "did:key:zQ3shvD5cZSLggSCiu4jmF3jRY6GMUb7zvwChfhYQGJfQudJE"
	nonce := "1eba606a"

	c := Controller{
		bindingFile:  BindingFile,
		bindingNonce: nonce,
		ownerDID:     ownerDID,
		Identity:     suite.Identity,
	}

	b := c.bindACK(ownerDID, BindACKParams{
		Timestamp: "1618456405107",
		// the following signature is signed with 1eba606e + timestmap (not the tested nonce 1eba606a)
		Signature: "3045022100d500b7ebbadeed51aaff844a0e7d741eb5bbf4c14b8d8476d87fae4ae02ab08b0220787dcaeae59327d1ff17db5b25386bf3250425a702a212e4fbd470b890d45ea8",
	})

	resp := map[string]json.RawMessage{}
	suite.NoError(json.Unmarshal(b, &resp))
	suite.Nil(resp["data"])
	suite.NotNil(resp["error"])

	var err string
	suite.NoError(json.Unmarshal(resp["error"], &err))
	suite.Equal(err, "invalid binding ack signature")
}

func (suite *ControllerTestSuite) TestBindAckWithInvalidSignature() {
	ownerDID := "did:key:zQ3shvD5cZSLggSCiu4jmF3jRY6GMUb7zvwChfhYQGJfQudJE"
	nonce := "1eba606e"

	c := Controller{
		bindingFile:  BindingFile,
		bindingNonce: nonce,
		ownerDID:     ownerDID,
		Identity:     suite.Identity,
	}

	b := c.bindACK(ownerDID, BindACKParams{
		Timestamp: "1618456405107",
		Signature: "3045022100d500b7ebbadeed51aaff844a0e7d741eb5bbf4c14b8d8476d87fae4ae02ab08b0220787dcaeae59327d1ff17db5b25386bf3250425a702a212e4fbd470b890d45ea8",
	})

	resp := map[string]json.RawMessage{}
	suite.NoError(json.Unmarshal(b, &resp))
	suite.Nil(resp["data"])
	suite.NotNil(resp["error"])

	var err string
	suite.NoError(json.Unmarshal(resp["error"], &err))
	suite.Equal(err, "invalid binding ack signature")
}

func (suite *ControllerTestSuite) TestHasRPCAccess() {
	ownerDID := "did:key:zQ3shvD5cZSLggSCiu4jmF3jRY6GMUb7zvwChfhYQGJfQudJE"
	memberDID := "did:key:zQ3shrG4MGtHFTq4BMaPtWRysMuTXVB5H2G4upbQzvk9PyANM"

	c := Controller{
		bindingFile: BindingFile,
		ownerDID:    ownerDID,
		Identity:    suite.Identity,
	}

	access := map[string]map[string]bool{
		ownerDID: {
			"createwallet": true,
			"setmember":    true,
			"removemember": true,
			"bitcoind":     true,
		},
		memberDID: {
			"bitcoind": true,
		},
	}
	for did, access := range access {
		for rpc, allowed := range access {
			suite.Equal(allowed, c.HasRPCAccess(did, rpc))
		}
	}
}

func (suite *ControllerTestSuite) TestHasBitcoinRPCAccess() {
	ownerDID := "did:key:zQ3shvD5cZSLggSCiu4jmF3jRY6GMUb7zvwChfhYQGJfQudJE"
	memberWithLimitedAccessDID := "did:key:zQ3shrG4MGtHFTq4BMaPtWRysMuTXVB5H2G4upbQzvk9PyANM"
	memberWithMinimalAccessDID := "did:key:zQ3shokFTS3brHcDQrn82RUDfCZESWL1ZdCEJwekUDPQiYBme"
	strangerDID := "did:key:zQ3shvtd8SgRF7UBHzJsnH1Qu2MKgEMBRujEfr4wp5a171Vv4"

	mockCtl := gomock.NewController(suite.T())
	defer mockCtl.Finish()
	mockedStore := NewMockStore(mockCtl)
	mockedStore.EXPECT().AccessMode(memberWithLimitedAccessDID).AnyTimes().Return(AccessModeLimited)
	mockedStore.EXPECT().AccessMode(memberWithMinimalAccessDID).AnyTimes().Return(AccessModeMinimal)
	mockedStore.EXPECT().AccessMode(strangerDID).AnyTimes().Return(AccessModeNotApplicant)

	c := Controller{
		bindingFile: BindingFile,
		ownerDID:    ownerDID,
		Identity:    suite.Identity,
		store:       mockedStore,
	}

	access := map[string]map[string]bool{
		ownerDID: {
			"sendrawtransaction": true,
			"getbalances":        true,
			"getblockchaininfo":  true,
		},
		memberWithLimitedAccessDID: {
			"getbalances":       true,
			"getblockchaininfo": true,
		},
		memberWithMinimalAccessDID: {
			"getblockchaininfo": true,
		},
		strangerDID: {},
	}
	for did, access := range access {
		for rpc, allowed := range access {
			suite.Equal(allowed, c.HasBitcoinRPCAccess(did, rpc))
		}
	}
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
