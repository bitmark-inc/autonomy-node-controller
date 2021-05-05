package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ACLTestSuite struct {
	suite.Suite
}

func (suite *ACLTestSuite) TestHasRPCAccess() {
	access := map[AccessMode]map[string]bool{
		AccessModeFull: {
			"create_wallet": true,
			"set_member":    true,
			"remove_member": true,
			"bitcoind":      true,
		},
		AccessModeLimited: {
			"bitcoind": true,
		},
		AccessModeMinimal: {
			"bitcoind": true,
		},
		AccessModeNotApplicant: {},
	}
	for mode, access := range access {
		for rpc, allowed := range access {
			suite.Equal(allowed, HasRPCAccess(rpc, mode))
		}
	}
}

func (suite *ACLTestSuite) TestHasBitcoinRPCAccess() {
	access := map[AccessMode]map[string]bool{
		AccessModeFull: {
			"sendrawtransaction": true,
			"getbalances":        true,
			"getblockchaininfo":  true,
		},
		AccessModeLimited: {
			"sendrawtransaction": false,
			"getbalances":        true,
			"getblockchaininfo":  true,
		},
		AccessModeMinimal: {
			"sendrawtransaction": false,
			"getbalances":        false,
			"getblockchaininfo":  true,
		},
		AccessModeNotApplicant: {
			"sendrawtransaction": false,
			"getbalances":        false,
			"getblockchaininfo":  false,
		},
	}
	for mode, access := range access {
		for rpc, allowed := range access {
			suite.Equal(allowed, HasBitcoinRPCAccess(rpc, mode))
		}
	}
}

func TestACLTestSuite(t *testing.T) {
	suite.Run(t, &ACLTestSuite{})
}
