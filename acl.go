package main

type AccessMode int

const (
	AccessModeNotApplicant = AccessMode(-1)
	AccessModeFull         = AccessMode(0)
	AccessModeLimited      = AccessMode(1)
	AccessModeMinimal      = AccessMode(2)
)

var (
	fullAccessBlockList = map[string]bool{
		"unloadwallet": true,
	}
	limitedAccessBlockList = map[string]bool{
		"unloadwallet":       true,
		"sendrawtransaction": true,
	}
	minimalAccessBlockList = map[string]bool{
		"unloadwallet":       true,
		"sendrawtransaction": true,
		"getbalances":        true,
	}

	rocBlockList = map[AccessMode]map[string]bool{
		AccessModeFull:    fullAccessBlockList,
		AccessModeLimited: limitedAccessBlockList,
		AccessModeMinimal: minimalAccessBlockList,
	}
)

func (c *Controller) HasBitcoinRPCAccess(did, rpcCommand string) bool {
	if did == c.ownerDID {
		_, blocked := fullAccessBlockList[rpcCommand]
		return !blocked
	}

	mode := c.store.AccessMode(did)
	if mode == AccessModeNotApplicant {
		return false
	}
	_, blocked := rocBlockList[mode][rpcCommand]
	return !blocked
}
