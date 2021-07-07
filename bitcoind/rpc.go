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
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

type rpc struct {
	Jsonrpc string        `json:"jsonrpc"`
	Id      int           `json:"id,string"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type RPCResult struct {
	Result json.RawMessage
	Error  map[string]interface{}
}

func (c *HttpBitcoind) rpcCall(method string, params []interface{}) (json.RawMessage, error) {
	status, body, err := c.Call(method, params)
	if err != nil {
		return nil, err
	}

	if status == 401 {
		return nil, fmt.Errorf("%s", "bitcoind return 401 status code")
	}

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

func (c *HttpBitcoind) Call(method string, params []interface{}) (int, json.RawMessage, error) {
	id := int(atomic.AddUint32(c.sequence, 1) & 0xffff)

	call := rpc{
		Jsonrpc: "1.0",
		Id:      id,
		Method:  method,
		Params:  params,
	}

	rpc, err := json.Marshal(call)
	if err != nil {
		return 0, nil, err
	}

	log.Debugf("call: %s", rpc)

	buffer := bytes.NewReader(rpc)
	resp, err := c.HttpClient.Post(c.serverURL, "application/json", buffer)
	if err != nil {
		log.Error(err)
		return 0, nil, fmt.Errorf("bitcoind request failed")
	}
	defer resp.Body.Close()

	var body json.RawMessage
	if resp.StatusCode != 401 {
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, nil, err
		}
		log.Debugf("body: %s\n", body)
	}

	return resp.StatusCode, body, nil
}
