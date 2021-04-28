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

func (c *Controller) HasRPCAccess(did, command string) bool {
	switch command {
	case "bitcoind":
		return true
	default:
		if did == c.ownerDID {
			return true
		}
		return false
	}
}

// TODO: this method could be integrated in to `HasRPCAccess`
func (c *Controller) HasBitcoinRPCAccess(did, rpcCommand string) bool {
	var mode AccessMode
	if did == c.ownerDID {
		mode = AccessModeFull
	} else {
		mode = c.store.AccessMode(did)
	}

	if mode == AccessModeNotApplicant {
		return false
	}
	_, blocked := rpcBlockList[mode][rpcCommand]
	return !blocked
}
