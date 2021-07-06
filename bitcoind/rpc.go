// SPDX-License-Identifier: ISC
// Copyright (c) 2019-2020 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bitcoind

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

type rpc struct {
	Jsonrpc string        `json:"jsonrpc"`
	Id      int           `json:"id,string"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

var sequence uint32
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

type RPCResult struct {
	Result json.RawMessage
	Error  map[string]interface{}
}

func rpccall(url string, method string, params []interface{}) (json.RawMessage, error) {

	id := int(atomic.AddUint32(&sequence, 1) & 0xffff)

	call := rpc{
		Jsonrpc: "1.0",
		Id:      id,
		Method:  method,
		Params:  params,
	}

	rpc, err := json.Marshal(call)
	if err != nil {
		return nil, err
	}

	log.Debugf("call: %s", rpc)

	buffer := bytes.NewReader(rpc)
	resp, err := httpClient.Post(url, "application/json", buffer)
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("bitcoind request failed")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Debugf("body: %s\n", body)

	var v RPCResult
	err = json.Unmarshal(body, &v)
	if err != nil {
		return nil, err
	}

	e := v.Error
	if e != nil {
		return nil, fmt.Errorf("%s", e["message"])
	}

	return v.Result, nil
}
