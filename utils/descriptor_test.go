// SPDX-License-Identifier: ISC
// Copyright (c) 2019-2021 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package utils

import (
	"testing"

	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/stretchr/testify/assert"
)

func TestExtractGordianKeyDerivationPath(t *testing.T) {
	d := "wsh(sortedmulti(2,[119dbcab/48h/1h/0h/2h]tpubDFYr9xD4WtT3yDBdX2qT2j2v6ZruqccwPKFwLguuJL99bWBrk6D2Lv1aPpRbFnw1sQUU9DM7ScMAkPRJqR1iXKhWMBNMAJ45QCTuvSZbzzv/0/*,[e650dc93/48h/1h/0h/2h]tpubDEijNAeHVNmm6wHwspPv4fV8mRkoMimeVCk47dExpN9e17jFti12BdjzL8MX17GvKEekRzknNuDoLy1Q8fujYfsWfCvjwYmjjENUpzwDy6B/0/*,[<fingerprint>/48h/1h/0h/2h]<xpub>/0/*))"
	path := ExtractGordianKeyDerivationPath(d)
	assert.Equal(t, "/48h/1h/0h/2h", path)
}

func TestParseDerivationPath(t *testing.T) {
	cases := map[string][]uint32{
		"/48h/0h/0h/2h": {hdkeychain.HardenedKeyStart + 48, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart + 2},
		"/48h/1h/0h/2h": {hdkeychain.HardenedKeyStart + 48, hdkeychain.HardenedKeyStart + 1, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart + 2},
	}

	for derivationPath, path := range cases {
		p, err := ParseDerivationPath(derivationPath)
		assert.NoError(t, err)
		assert.Equal(t, path, p)
	}
}
