// SPDX-License-Identifier: ISC
// Copyright (c) 2019-2021 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

type AccessMode int

const (
	AccessModeNotApplicant = AccessMode(-1)
	AccessModeFull         = AccessMode(0)
	AccessModeLimited      = AccessMode(1)
	AccessModeMinimal      = AccessMode(2)
)

var (
	fullAccessCommandAllowList = map[string]bool{
		"bind":                true,
		"bind_ack":            true,
		"bitcoind":            true,
		"create_wallet":       true,
		"finish_psbt":         true,
		"set_member":          true,
		"remove_member":       true,
		"start_bitcoind":      true,
		"stop_bitcoind":       true,
		"get_bitcoind_status": true,
	}
	limitedAccessCommandAllowList = map[string]bool{
		"bind":                true,
		"bind_ack":            true,
		"bitcoind":            true,
		"get_bitcoind_status": true,
	}
	minimalAccessCommandAllowList = map[string]bool{
		"bind":                true,
		"bind_ack":            true,
		"bitcoind":            true,
		"get_bitcoind_status": true,
	}

	commandAllowList = map[AccessMode]map[string]bool{
		AccessModeFull:    fullAccessCommandAllowList,
		AccessModeLimited: limitedAccessCommandAllowList,
		AccessModeMinimal: minimalAccessCommandAllowList,
	}
)

var (
	fullAccessBitcoinRPCAllowList = map[string]bool{
		"getbalances":            true,
		"getblockchaininfo":      true,
		"getmininginfo":          true,
		"getnettotals":           true,
		"getnetworkinfo":         true,
		"getnewaddress":          true,
		"getreceivedbyaddress":   true,
		"gettransaction":         true,
		"getwalletinfo":          true,
		"listtransactions":       true,
		"walletcreatefundedpsbt": true,
	}
	limitedAccessBitcoinRPCAllowList = map[string]bool{}
	minimalAccessBitcoinRPCAllowList = map[string]bool{}

	bitcoinRPCAllowList = map[AccessMode]map[string]bool{
		AccessModeFull:    fullAccessBitcoinRPCAllowList,
		AccessModeLimited: limitedAccessBitcoinRPCAllowList,
		AccessModeMinimal: minimalAccessBitcoinRPCAllowList,
	}
)

func HasCommandAccess(command string, mode AccessMode) bool {
	_, ok := commandAllowList[mode][command]
	return ok
}

// TODO: this method could be integrated in to `HasRPCAccess`
func HasBitcoinRPCAccess(rpcCommand string, mode AccessMode) bool {
	if mode == AccessModeNotApplicant {
		return false
	}
	_, allowed := bitcoinRPCAllowList[mode][rpcCommand]
	return allowed
}
