package utils

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/btcsuite/btcutil/hdkeychain"
)

func ExtractGordianKeyDerivationPath(incompleteDescriptor string) string {
	re := regexp.MustCompile(`(?m)\[\<fingerprint\>(.+)\]`)
	match := re.FindStringSubmatch(incompleteDescriptor)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

func ParseDerivationPath(derivationPath string) ([]uint32, error) {
	path := make([]uint32, 0)

	parts := strings.Split(derivationPath, "/")
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}

		var i uint32
		if strings.HasSuffix(p, "h") || strings.HasSuffix(p, "H") || strings.HasSuffix(p, "'") {
			p = p[:len(p)-1]
			i = hdkeychain.HardenedKeyStart
		}

		val, ok := new(big.Int).SetString(p, 0)
		if !ok {
			return nil, fmt.Errorf("invalid component: %s", p)
		}
		i += uint32(val.Uint64())
		path = append(path, i)
	}
	return path, nil
}
