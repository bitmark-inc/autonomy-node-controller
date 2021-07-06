// SPDX-License-Identifier: ISC
// Copyright (c) 2019-2020 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bitcoind

import (
	"encoding/json"
)

type Vin struct {
	TxID string `json:"txid"`
	Vout int64  `json:"vout"`
}

type Vout struct {
	Value        float64 `json:"value"`
	N            int     `json:"n"`
	ScriptPubKey struct {
		Asm       string   `json:"asm"`
		Hex       string   `json:"hex"`
		ReqSigs   int      `json:"reqSigs"`
		Type      string   `json:"type"`
		Addresses []string `json:"addresses"`
	} `json:"scriptPubKey,omitempty"`
}
type RawTransaction struct {
	TxID          string `json:"txid"`
	Hash          string `json:"hash"`
	Version       int    `json:"version"`
	Size          int    `json:"size"`
	Vsize         int    `json:"vsize"`
	Weight        int    `json:"weight"`
	Locktime      int    `json:"locktime"`
	Vins          []Vin  `json:"vin"`
	Vouts         []Vout `json:"vout"`
	Hex           string `json:"hex"`
	Blockhash     string `json:"blockhash"`
	Confirmations int    `json:"confirmations"`
	Time          int    `json:"time"`
	Blocktime     int    `json:"blocktime"`
}

func GetRawTransaction(url string, txId string) (*RawTransaction, error) {

	params := make([]interface{}, 2)

	params[0] = txId
	params[1] = true

	b, err := rpccall(url, "getrawtransaction", params)
	if err != nil {
		return nil, err
	}

	var result RawTransaction
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
