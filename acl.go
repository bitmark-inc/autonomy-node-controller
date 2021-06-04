package main

type AccessMode int

const (
	AccessModeNotApplicant = AccessMode(-1)
	AccessModeFull         = AccessMode(0)
	AccessModeLimited      = AccessMode(1)
	AccessModeMinimal      = AccessMode(2)
)

var (
	fullAccessRPCAllowList = map[string]bool{
		"bind":          true,
		"bind_ack":      true,
		"bitcoind":      true,
		"create_wallet": true,
		"finish_psbt":   true,
		"set_member":    true,
		"remove_member": true,
	}
	limitedAccessRPCAllowList = map[string]bool{
		"bind":     true,
		"bind_ack": true,
		"bitcoind": true,
	}
	minimalAccessRPCAllowList = map[string]bool{
		"bind":     true,
		"bind_ack": true,
		"bitcoind": true,
	}

	rpcAllowList = map[AccessMode]map[string]bool{
		AccessModeFull:    fullAccessRPCAllowList,
		AccessModeLimited: limitedAccessRPCAllowList,
		AccessModeMinimal: minimalAccessRPCAllowList,
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

func HasRPCAccess(command string, mode AccessMode) bool {
	_, ok := rpcAllowList[mode][command]
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
