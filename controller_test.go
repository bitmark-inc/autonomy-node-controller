// SPDX-License-Identifier: ISC
// Copyright (c) 2019-2021 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"errors"
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

func TestControllerTestSuite(t *testing.T) {
	i, err := NewPodIdentity()
	if err != nil {
		t.Fatal("unable to create test key file")
	}
	suite.Run(t, &ControllerTestSuite{
		Identity: i,
	})
}
