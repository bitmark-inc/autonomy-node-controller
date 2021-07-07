// SPDX-License-Identifier: ISC
// Copyright (c) 2019-2021 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/rpcclient"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/bitmark-inc/autonomy-pod-controller/config"
	"github.com/bitmark-inc/autonomy-pod-controller/key"
	"github.com/bitmark-inc/autonomy-pod-controller/messaging"
	"github.com/bitmark-inc/autonomy-pod-controller/utils"
)

// BindACKParams is the parameters for command `bind_ack`
type BindACKParams struct {
	Timestamp string `json:"timestamp"`
	Signature string `json:"signature"`
}

// BindACKParams is the parameters for command `bind_ack`
type BitcoindRPCParams struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type CreateWalletRPCParams struct {
	Descriptor string `json:"descriptor"`
}

type UpdateMemberAccessModeRPCParams struct {
	MemberDID  string     `json:"member_did"`
	AccessMode AccessMode `json:"access_mode"`
}

type RemoveMemberAccessModeRPCParams struct {
	MemberDID string `json:"member_did"`
}

type FinishPSBTRPCParams struct {
	PSBT string `json:"psbt"`
}

type BitcoindResponse struct {
	StatusCode   int    `json:"statusCode"`
	ResponseBody []byte `json:"responseBody"`
}

type Controller struct {
	ownerDID       string
	httpClient     *http.Client
	Identity       *PodIdentity
	store          Store
	LastActiveTime time.Time
}

func NewController(ownerDID string, i *PodIdentity) *Controller {
	return &Controller{
		ownerDID: ownerDID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Identity:       i,
		store:          NewBoltStore(config.AbsoluteApplicationFilePath(viper.GetString("db_name"))),
		LastActiveTime: time.Now(),
	}
}

// Process handles messages from clients and returns a response message
func (c *Controller) Process(m *messaging.Message) [][]byte {
	defer func() {
		c.LastActiveTime = time.Now()
		if r := recover(); r != nil {
			log.WithField("recover", r).Error("panic caught")
		}
	}()

	var req RequestCommand
	if err := json.Unmarshal(m.Content, &req); err != nil {
		log.WithError(err).Error("fail to decode content")
	}

	accessMode := c.accessMode(m.Source)

	if !HasCommandAccess(req.Command, accessMode) {
		return CommandResponse(req.ID, nil, errors.New("not allowed to use this command"))
	}

	if !c.hasCorrectBindingState(m.Source, req.Command) {
		return CommandResponse(req.ID, nil, errors.New("incorrect binding state"))
	}

	log.WithField("command request", req).Debug("parse command")
	switch req.Command {
	case "bind":
		resp, err := c.bind(m.Source)
		return CommandResponse(req.ID, resp, err)
	case "bind_ack":
		var params BindACKParams
		if err := json.Unmarshal(req.Args, &params); err != nil {
			return CommandResponse(req.ID, nil, fmt.Errorf("bad request for bind_ack: %s", err.Error()))
		}

		resp, err := c.bindACK(m.Source, params)
		return CommandResponse(req.ID, resp, err)
	case "create_wallet":
		var params CreateWalletRPCParams
		if err := json.Unmarshal(req.Args, &params); err != nil {
			return CommandResponse(req.ID, nil, fmt.Errorf("bad request for create_wallet: %s", err.Error()))
		}

		resp, err := c.createWallet(params.Descriptor)
		return CommandResponse(req.ID, resp, err)
	case "finish_psbt":
		var params FinishPSBTRPCParams
		if err := json.Unmarshal(req.Args, &params); err != nil {
			return CommandResponse(req.ID, nil, fmt.Errorf("bad request for finish_psbt: %s", err.Error()))
		}

		resp, err := c.finishPSBT(params.PSBT)
		return CommandResponse(req.ID, resp, err)
	case "set_member":
		var params UpdateMemberAccessModeRPCParams
		if err := json.Unmarshal(req.Args, &params); err != nil {
			return CommandResponse(req.ID, nil, fmt.Errorf("bad request for set_member: %s", err.Error()))
		}

		resp, err := c.setMember(params.MemberDID, params.AccessMode)
		return CommandResponse(req.ID, resp, err)
	case "remove_member":
		var params RemoveMemberAccessModeRPCParams
		if err := json.Unmarshal(req.Args, &params); err != nil {
			return CommandResponse(req.ID, nil, fmt.Errorf("bad request for remove_member: %s", err.Error()))
		}
		resp, err := c.removeMember(params.MemberDID)
		return CommandResponse(req.ID, resp, err)
	case "start_bitcoind":
		resp, err := c.startBitcoind()
		return CommandResponse(req.ID, resp, err)
	case "stop_bitcoind":
		resp, err := c.stopBitcoind()
		return CommandResponse(req.ID, resp, err)
	case "get_bitcoind_status":
		resp, err := c.getBitcoindStatus()
		return CommandResponse(req.ID, resp, err)
	case "bitcoind":
		var params BitcoindRPCParams
		if err := json.Unmarshal(req.Args, &params); err != nil {
			return CommandResponse(req.ID, nil, fmt.Errorf("bad request for bitcoind: %s", err.Error()))
		}

		if !HasBitcoinRPCAccess(params.Method, accessMode) {
			return CommandResponse(req.ID, nil, errors.New("not allowed to use this RPC"))
		}

		resp, err := c.bitcoinRPC(m.Source, params)
		return CommandResponse(req.ID, resp, err)
	default:
		return CommandResponse(req.ID, nil, fmt.Errorf("unsupported command"))
	}
}

func (c *Controller) accessMode(did string) AccessMode {
	if did == c.ownerDID {
		return AccessModeFull
	}

	mode := c.store.MemberAccessMode(did)
	if mode > AccessModeMinimal {
		return AccessModeNotApplicant
	}
	return mode
}

func (c *Controller) hasCorrectBindingState(did, command string) bool {
	switch command {
	case "bind", "bind_ack":
		return !c.store.HasBinding(did)
	default:
		return c.store.HasBinding(did)
	}
}

// bind triggers the bind process which is triggerred by a client.
func (c *Controller) bind(did string) (map[string]string, error) {
	b, err := utils.GenerateRandomBytes(4)
	if err != nil {
		return nil, err
	}

	nonce := hex.EncodeToString(b)
	nowString := fmt.Sprint(int64(time.Now().UnixNano()) / int64(time.Millisecond))
	signature, err := key.Sign(c.Identity.PrivateKey, nonce+nowString)
	if err != nil {
		return nil, err
	}

	if err := c.store.SetBinding(did, nonce); err != nil {
		return nil, err
	}

	return map[string]string{
		"identity":  c.Identity.DID,
		"nonce":     nonce,
		"timestamp": nowString,
		"signature": signature,
	}, nil
}

// bindACK process the client response of a binding process. It checks the nonce
// and the signature using owner DID.
func (c *Controller) bindACK(did string, ackParams BindACKParams) (map[string]string, error) {
	nonce := c.store.BindingNonce(did)

	if !key.VerifySignature(did, nonce+ackParams.Timestamp, ackParams.Signature) {
		err := fmt.Errorf("invalid binding ack signature")
		log.WithError(err).Error("fail to bind account")
		return nil, err
	}

	if err := c.store.CompleteBinding(did); err != nil {
		log.WithError(err).Error("fail to bind account")
		return nil, err
	}
	log.WithField("did", did).Println("bind account successfully")

	return map[string]string{"status": "ok"}, nil
}

// bitcoinRPC runs bitcoind rpc for clients
func (c *Controller) bitcoinRPC(did string, bitcoindParams BitcoindRPCParams) (map[string]interface{}, error) {
	var reqBody bytes.Buffer

	if err := json.NewEncoder(&reqBody).Encode(bitcoindParams); err != nil {
		return nil, err
	}

	// log.WithField("reqBody", reqBody.String()).Info("request body")
	req, err := http.NewRequest("POST", viper.GetString("bitcoind.rpcconnect"), &reqBody)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(viper.GetString("bitcoind.rpcuser"), viper.GetString("bitcoind.rpcpassword"))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseBody json.RawMessage
	if resp.StatusCode != 401 {
		responseBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	}

	return map[string]interface{}{
		"statusCode":   resp.StatusCode,
		"responseBody": responseBody,
	}, nil
}

// createWallet creates a new descriptor wallet if it does not exist
// according to the semi-finished descriptor (already including the platform and recovery key information).
// An example of an incomplete descriptor:
//
// wsh(sortedmulti(2,[119dbcab/48h/1h/0h/2h]tpubDFYr9xD4WtT3yDBdX2qT2j2v6ZruqccwPKFwLguuJL99bWBrk6D2Lv1aPpRbFnw1sQUU9DM7ScMAkPRJqR1iXKhWMBNMAJ45QCTuvSZbzzv/0/*,[e650dc93/48h/1h/0h/2h]tpubDEijNAeHVNmm6wHwspPv4fV8mRkoMimeVCk47dExpN9e17jFti12BdjzL8MX17GvKEekRzknNuDoLy1Q8fujYfsWfCvjwYmjjENUpzwDy6B/0/*,[<fingerprint>/48h/1h/0h/2h]<xpub>/0/*))
func (c *Controller) createWallet(incompleteDescriptor string) (map[string]string, error) {
	derivationPath := utils.ExtractGordianKeyDerivationPath(incompleteDescriptor)
	if derivationPath == "" {
		return nil, fmt.Errorf("gordian key derivation path not found")
	}

	path, err := utils.ParseDerivationPath(derivationPath)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(viper.GetString("bitcoind.rpcconnect"))
	if err != nil {
		return nil, err
	}

	connCfg := &rpcclient.ConnConfig{
		Host:         fmt.Sprintf("%s:%s", u.Hostname(), u.Port()),
		User:         viper.GetString("bitcoind.rpcuser"),
		Pass:         viper.GetString("bitcoind.rpcpassword"),
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err
	}
	defer client.Shutdown()

	shouldCreateWallet := false
	shouldImportDescriptors := false
	walletInfo, err := client.GetWalletInfo()
	if err != nil {
		if jerr, ok := err.(*btcjson.RPCError); ok {
			switch jerr.Code {
			case btcjson.ErrRPCWalletNotFound:
				shouldCreateWallet = true
				shouldImportDescriptors = true
			default:
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		if walletInfo.KeyPoolSize == 0 && *walletInfo.KeyPoolSizeHDInternal == 0 {
			shouldImportDescriptors = true
		}
	}

	blockchainInfo, err := client.GetBlockChainInfo()
	if err != nil {
		return nil, err
	}
	keyFilePath := config.AbsoluteApplicationFilePath(viper.GetString("gordian_master_key_file"))
	masterKey, err := createOrLoadMasterKey(blockchainInfo.Chain, keyFilePath)
	if err != nil {
		return nil, err
	}
	masterFingerprint, err := computeFingerprint(masterKey)
	if err != nil {
		return nil, err
	}
	gordianPrivateKey := masterKey
	for _, i := range path {
		gordianPrivateKey, err = gordianPrivateKey.Derive(i)
		if err != nil {
			return nil, err
		}
	}
	gordianPublicKey, err := gordianPrivateKey.Neuter()
	if err != nil {
		return nil, err
	}

	if shouldCreateWallet {
		walletName, _ := json.Marshal(btcjson.String("gordian"))
		passphrase, _ := json.Marshal(btcjson.String(""))
		t, _ := json.Marshal(btcjson.Bool(true))
		f, _ := json.Marshal(btcjson.Bool(false))
		createWalletParams := []json.RawMessage{
			walletName, // wallet_name
			f,          // disable_private_keys
			t,          // blank
			passphrase, // passphrase
			t,          // avoid_reuse
			t,          // descriptors
			f,          // load_on_startup
		}

		if _, err := client.RawRequest("createwallet", createWalletParams); err != nil {
			return nil, err
		}
	}

	if shouldImportDescriptors {
		// FIXME: check network
		importedDescriptorReplacer := strings.NewReplacer(
			"<fingerprint>", masterFingerprint,
			"<xpub>", gordianPrivateKey.String(),
			"<tpub>", gordianPrivateKey.String(),
		)
		externalDescriptorWithoutChecksum := importedDescriptorReplacer.Replace(incompleteDescriptor)
		externalDescriptorInfo, err := client.GetDescriptorInfo(externalDescriptorWithoutChecksum)
		if err != nil {
			return nil, err
		}

		internalDescriptorWithoutChecksum := strings.ReplaceAll(externalDescriptorWithoutChecksum, "/0/*", "/1/*")
		internalDescriptorInfo, err := client.GetDescriptorInfo(internalDescriptorWithoutChecksum)
		if err != nil {
			return nil, err
		}

		descriptors := []map[string]interface{}{
			{
				"desc":      fmt.Sprintf("%s#%s", externalDescriptorWithoutChecksum, externalDescriptorInfo.Checksum),
				"active":    true,
				"timestamp": "now",
				"internal":  false,
			},
			{
				"desc":      fmt.Sprintf("%s#%s", internalDescriptorWithoutChecksum, internalDescriptorInfo.Checksum),
				"active":    true,
				"timestamp": "now",
				"internal":  true,
			},
		}
		b, _ := json.Marshal(descriptors)
		if _, err := client.RawRequest("importdescriptors", []json.RawMessage{b}); err != nil {
			return nil, err
		}
	}

	// FIXME: check network
	accountMapDescriptorReplacer := strings.NewReplacer(
		"<fingerprint>", masterFingerprint,
		"<xpub>", gordianPublicKey.String(),
		"<tpub>", gordianPrivateKey.String(),
	)
	gordianWalletDescriptor := accountMapDescriptorReplacer.Replace(incompleteDescriptor)
	return map[string]string{"descriptor": gordianWalletDescriptor}, nil
}

// finishPSBT finalizes the PSBT and broadcasts the transaction
func (c *Controller) finishPSBT(psbt string) (map[string]string, error) {
	u, err := url.Parse(viper.GetString("bitcoind.rpcconnect"))
	if err != nil {
		return nil, err
	}
	connCfg := &rpcclient.ConnConfig{
		Host:         fmt.Sprintf("%s:%s", u.Hostname(), u.Port()),
		User:         viper.GetString("bitcoind.rpcuser"),
		Pass:         viper.GetString("bitcoind.rpcpassword"),
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err
	}
	defer client.Shutdown()

	processedPSBT, err := client.WalletProcessPsbt(psbt, btcjson.Bool(true), rpcclient.SigHashAll, btcjson.Bool(true))
	if err != nil {
		return nil, err
	}
	if !processedPSBT.Complete {
		return nil, fmt.Errorf("psbt not completed: %s", err)
	}

	psbtBytes, _ := json.Marshal(btcjson.String(processedPSBT.Psbt))
	r, err := client.RawRequest("finalizepsbt", []json.RawMessage{psbtBytes})
	if err != nil {
		return nil, err
	}
	var finalizePSBTResult struct {
		PSBT     string `json:"psbt"`
		Hex      string `json:"hex"`
		Complete bool   `json:"complete"`
	}
	if err := json.Unmarshal(r, &finalizePSBTResult); err != nil {
		return nil, fmt.Errorf("unexpected response from finalizepsbt: %s", err)
	}
	if !finalizePSBTResult.Complete {
		return nil, fmt.Errorf("psbt not finalized: %s", err)
	}

	txBytes, _ := json.Marshal(btcjson.String(finalizePSBTResult.Hex))
	r, err = client.RawRequest("sendrawtransaction", []json.RawMessage{txBytes})
	if err != nil {
		return nil, err
	}
	var txID string
	if err := json.Unmarshal(r, &txID); err != nil {
		return nil, fmt.Errorf("unexpected response from sendrawtransaction: %s", err)
	}

	return map[string]string{"txid": txID}, nil
}

func (c *Controller) setMember(memberDID string, accessMode AccessMode) (map[string]string, error) {
	if err := c.store.UpdateMemberAccessMode(memberDID, accessMode); err != nil {
		return nil, err
	}
	return map[string]string{"status": "ok"}, nil
}

func (c *Controller) removeMember(memberDID string) (map[string]string, error) {
	if err := c.store.RemoveMember(memberDID); err != nil {
		return nil, err
	}
	return map[string]string{"status": "ok"}, nil
}

func (c *Controller) startBitcoind() (*BitcoindResponse, error) {
	req, err := http.NewRequest("POST", viper.GetString("bitcoind_ctl.endpoint")+"/start", nil)
	if err != nil {
		log.WithError(err).Error("fail to create bitcoind-ctl api request")
		return nil, fmt.Errorf("fail to create bitcoind-ctl api request")
	}
	result, err := c.doHttpRequest(req)
	if err != nil {
		log.WithError(err).Error("fail to call bitcoind-ctl api")
		return nil, fmt.Errorf("fail to call bitcoind-ctl api")
	}
	return result, nil
}

func (c *Controller) stopBitcoind() (*BitcoindResponse, error) {
	req, err := http.NewRequest("POST", viper.GetString("bitcoind_ctl.endpoint")+"/stop", nil)
	if err != nil {
		log.WithError(err).Error("fail to create bitcoind-ctl api request")
		return nil, fmt.Errorf("fail to create bitcoind-ctl api request")
	}
	result, err := c.doHttpRequest(req)
	if err != nil {
		log.WithError(err).Error("fail to call bitcoind-ctl api")
		return nil, fmt.Errorf("fail to call bitcoind-ctl api")
	}
	return result, nil
}

func (c *Controller) getBitcoindStatus() (*BitcoindResponse, error) {
	req, err := http.NewRequest("GET", viper.GetString("bitcoind_ctl.endpoint")+"/status", nil)
	if err != nil {
		log.WithError(err).Error("fail to create bitcoind-ctl api request")
		return nil, fmt.Errorf("fail to create bitcoind-ctl api request")
	}
	result, err := c.doHttpRequest(req)
	if err != nil {
		log.WithError(err).Error("fail to call bitcoind-ctl api")
		return nil, fmt.Errorf("fail to call bitcoind-ctl api")
	}
	return result, nil
}

func (c *Controller) doHttpRequest(req *http.Request) (*BitcoindResponse, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &BitcoindResponse{}, err
	}
	defer resp.Body.Close()

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &BitcoindResponse{}, err
	}

	return &BitcoindResponse{
		StatusCode:   resp.StatusCode,
		ResponseBody: resBody,
	}, nil
}
