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
	fullAccessBitcoinRPCBlockList = map[string]bool{
		"unloadwallet": true,
	}
	limitedAccessBitcoinRPCBlockList = map[string]bool{
		"unloadwallet":       true,
		"sendrawtransaction": true,
	}
	minimalAccessBitcoinRPCBlockList = map[string]bool{
		"unloadwallet":       true,
		"sendrawtransaction": true,
		"getbalances":        true,
	}

	// TODO: we might need the allow list in the future
	// once we know in more detail what the members will be allowed to access in the Group Pod
	rpcBlockList = map[AccessMode]map[string]bool{
		AccessModeFull:    fullAccessBitcoinRPCBlockList,
		AccessModeLimited: limitedAccessBitcoinRPCBlockList,
		AccessModeMinimal: minimalAccessBitcoinRPCBlockList,
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
	_, blocked := rpcBlockList[mode][rpcCommand]
	return !blocked
}
