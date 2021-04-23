package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
)

func createOrLoadMasterKey(chain, path string) (*hdkeychain.ExtendedKey, error) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		// read master key from file
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("can't read the master key file: %s", err)
		}
		return hdkeychain.NewKeyFromString(string(content))
	}

	seed := make([]byte, 64)
	if _, err := rand.Read(seed); err != nil {
		return nil, fmt.Errorf("can't generate seed: %s", err)
	}
	c := chaincfg.MainNetParams
	if chain == "test" {
		c = chaincfg.TestNet3Params
	}
	masterKey, err := hdkeychain.NewMaster(seed, &c)
	if err != nil {
		return nil, fmt.Errorf("can't generate the master key: %s", err)
	}

	// write master key to file
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("")
	}
	defer f.Close()

	if _, err := f.WriteString(masterKey.String()); err != nil {
		return nil, fmt.Errorf("can't create the master key file: %s", err)
	}

	return masterKey, nil
}

func computeFingerprint(key *hdkeychain.ExtendedKey) (string, error) {
	p, err := key.ECPubKey()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(btcutil.Hash160(p.SerializeCompressed())[:4]), nil
}
