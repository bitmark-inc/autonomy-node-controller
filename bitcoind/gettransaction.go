// SPDX-License-Identifier: ISC
// Copyright (c) 2019-2020 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bitcoind

import (
	"encoding/json"
)

const (
	TxSend    = "send"
	TxReceive = "receive"
)

type Transaction struct {
	Amount        float64              `json:"amount"`
	Fee           float64              `json:"fee"`
	Blockheight   int                  `json:"blockheight"`
	TxID          string               `json:"txid"`
	Timereceived  int                  `json:"timereceived"`
	Blockhash     string               `json:"blockhash"`
	Confirmations int                  `json:"confirmations"`
	Time          int                  `json:"time"`
	Blocktime     int                  `json:"blocktime"`
	Details       []TransactionDetails `json:"details"`
	Decoded       RawTransaction       `json:"decoded"`
}

type TransactionDetails struct {
	Address  string  `json:"address"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Label    string  `json:"label"`
	Vout     int     `json:"vout"`
}

func (c *HttpBitcoind) GetTransaction(txId string) (*Transaction, error) {

	params := make([]interface{}, 3)

	params[0] = txId
	params[1] = true // include_watchonly
	params[2] = true // verbose - also decode raw data, same as getrawtransaction

	b, err := c.rpcCall("gettransaction", params)
	if err != nil {
		return nil, err
	}

	var result Transaction
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
