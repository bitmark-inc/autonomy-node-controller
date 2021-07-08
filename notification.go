// SPDX-License-Identifier: ISC
// Copyright (c) 2019-2021 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/bitmark-inc/autonomy-pod-controller/bitcoind"
)

func (c *Controller) transactionNotify(context *gin.Context) {
	txID := context.Param("txid")
	client, err := bitcoind.NewHttpRPCClient(c.httpClient)
	tx, err := client.GetTransaction(txID)
	if err != nil {
		logFields := map[string]interface{}{
			"txid": txID,
		}
		responseWithError(context, err, "failed to get tx from rpc", logFields)
		return
	}

	// forloop vins to get addresses
	vins := make([]string, 0)
	for _, vin := range tx.Decoded.Vins {
		addr, err := getVinAddresses(client, vin.TxID, int(vin.Vout))
		if err != nil {
			logFields := map[string]interface{}{
				"txid": vin.TxID,
			}
			responseWithError(context, err, "failed to get tx vin address", logFields)
			return
		}
		vins = append(vins, addr...)
	}

	// forloop vouts to get addresses and value
	vouts := make([]map[string]interface{}, 0)
	for _, vout := range tx.Decoded.Vouts {
		vouts = append(vouts, map[string]interface{}{
			"addresses": vout.ScriptPubKey.Addresses,
			"value":     vout.Value,
		})
	}

	// prepare the notify params
	type notifyFormat struct {
		AccountID string
		Contents  map[string]string
		Data      map[string]interface{}
	}
	notifyContent := map[string]string{
		"en": "Transaction Notification",
	}
	notifyData := map[string]interface{}{
		"TxID":          tx.TxID,
		"Confirmations": tx.Confirmations,
		"Category":      tx.Details[0].Category,
		"Amount":        tx.Amount,
		"Vins":          vins,
		"Vouts":         vouts,
	}

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(notifyFormat{AccountID: c.ownerDID, Contents: notifyContent, Data: notifyData}); err != nil {
		logFields := map[string]interface{}{
			"notifyData": notifyData,
		}
		responseWithError(context, err, "failed to encode tx", logFields)
		return
	}

	// start to call notification api
	notifyURL := viper.GetString("api_endpoint") + "/api/accounts/notification"

	notifyReq, _ := http.NewRequest("POST", notifyURL, body)
	notifyReq.Header.Add("Content-Type", "application/json")
	notifyReq.Header.Add("Authorization", "Bearer "+c.Identity.authToken)

	resp, err := c.httpClient.Do(notifyReq)
	if err != nil {
		logFields := map[string]interface{}{
			"notifyData": notifyData,
		}
		responseWithError(context, err, "failed to notify", logFields)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		r, _ := ioutil.ReadAll(resp.Body)
		logFields := map[string]interface{}{
			"status": resp.StatusCode,
			"body":   string(r),
		}
		responseWithError(context, err, "failed to notify", logFields)
		return
	}
	context.JSON(200, gin.H{"ok": 1})
}

func getVinAddresses(client *bitcoind.HttpBitcoind, txId string, vout int) ([]string, error) {
	tx, err := client.GetRawTransaction(txId)
	if err != nil {
		return nil, err
	}
	var addresses []string
	for i := range tx.Vouts {
		if tx.Vouts[i].N == vout {
			// Find and append the address
			addresses = append(addresses, tx.Vouts[i].ScriptPubKey.Addresses...)
			break
		}
	}
	return addresses, nil
}

func responseWithError(context *gin.Context, err error, message string, fields ...map[string]interface{}) {
	var withFields map[string]interface{}
	if len(fields) > 0 {
		withFields = fields[0]
	}
	log.WithError(err).WithFields(withFields).Error(message)
	context.JSON(500, gin.H{"error": errors.New(message)})
}
