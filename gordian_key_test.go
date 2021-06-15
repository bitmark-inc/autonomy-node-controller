// SPDX-License-Identifier: ISC
// Copyright (c) 2019-2021 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateOrLoadMasterKey(t *testing.T) {
	mainKeyPath := "/tmp/main-master-key"
	testKeyPath := "/tmp/test-master-key"
	defer os.Remove(mainKeyPath)
	defer os.Remove(testKeyPath)

	// mainnet
	createdKey, err := createOrLoadMasterKey("main", mainKeyPath)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(createdKey.String(), "xprv"))

	loadedKey, err := createOrLoadMasterKey("main", mainKeyPath)
	assert.NoError(t, err)
	assert.Equal(t, createdKey.String(), loadedKey.String())

	// testnet
	createdKey, err = createOrLoadMasterKey("test", testKeyPath)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(createdKey.String(), "tprv"))

	loadedKey, err = createOrLoadMasterKey("test", testKeyPath)
	assert.NoError(t, err)
	assert.Equal(t, createdKey.String(), loadedKey.String())
}
