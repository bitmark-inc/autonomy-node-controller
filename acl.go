package main

type AccessMode int

const (
	AccessModeNotApplicant = AccessMode(-1)
	AccessModeFull         = AccessMode(0)
	AccessModeLimited      = AccessMode(1)
	AccessModeMinimal      = AccessMode(2)
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
	switch command {
	case "bind", "bind_ack", "bitcoind":
		return mode > AccessModeNotApplicant && mode <= AccessModeMinimal
	default:
		return mode == AccessModeFull
	}
}

// TODO: this method could be integrated in to `HasRPCAccess`
func HasBitcoinRPCAccess(rpcCommand string, mode AccessMode) bool {
	if mode == AccessModeNotApplicant {
		return false
	}
	_, blocked := rpcBlockList[mode][rpcCommand]
	return !blocked
}
