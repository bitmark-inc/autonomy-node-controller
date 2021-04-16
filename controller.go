package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/bitmark-inc/autonomy-pod-controller/key"
	"github.com/bitmark-inc/autonomy-pod-controller/messaging"
	"github.com/bitmark-inc/autonomy-pod-controller/utils"
)

type CommandResponse [][]byte

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

type Controller struct {
	ownerDID     string
	bindingNonce string
	bindingFile  string
	httpClient   *http.Client
	Identity     *PodIdentity
}

func NewController(ownerDID string, i *PodIdentity) *Controller {
	return &Controller{
		ownerDID:    ownerDID,
		bindingFile: viper.GetString("binding_file"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Identity: i,
	}
}

// Process handles messages from clients and returns a response message
func (c *Controller) Process(m *messaging.Message) [][]byte {
	defer func() {
		if r := recover(); r != nil {
			log.WithField("recover", r).Error("panic caught")
		}
	}()

	var req RequestCommand
	if err := json.Unmarshal(m.Content, &req); err != nil {
		log.WithError(err).Error("fail to decode content")
	}

	log.WithField("command request", req).Debug("parse command")
	switch req.Command {
	case "bind":
		resp := c.bind(m.Source)
		return CommandResponse{resp}
	case "bind_ack":
		var params BindACKParams
		if err := json.Unmarshal(req.Args, &params); err != nil {
			return CommandResponse{ErrorResponse(fmt.Errorf("bad request for bind_ack. error: %s", err.Error()))}
		}

		resp := c.bindACK(m.Source, params)
		return CommandResponse{resp}
	case "bitcoind":
		var params BitcoindRPCParams
		if err := json.Unmarshal(req.Args, &params); err != nil {
			return CommandResponse{ErrorResponse(fmt.Errorf("bad request for bitcoind. error: %s", err.Error()))}
		}

		resp := c.bitcoinRPC(m.Source, params)
		return CommandResponse{resp}
	default:
		return CommandResponse{ErrorResponse(fmt.Errorf("unsupported command"))}
	}
}

// bind triggers the bind process which is triggerred by a client
// Only pre-defined owner DID is allowed to initiate the binding process.
func (c *Controller) bind(did string) []byte {
	if did != c.ownerDID {
		return ErrorResponse(errors.New("illegal owner"))
	}

	bound, err := c.IsPodBound()
	if err != nil {
		return ErrorResponse(err)
	}

	if bound {
		log.Warn("node has bound")
		return ErrorResponse(fmt.Errorf("node has bound"))
	}

	b, err := utils.GenerateRandomBytes(4)
	if err != nil {
		return ErrorResponse(err)
	}

	nonce := hex.EncodeToString(b)
	nowString := fmt.Sprint(int64(time.Now().UnixNano()) / int64(time.Millisecond))
	signature, err := key.Sign(c.Identity.PrivateKey, nonce+nowString)
	if err != nil {
		return ErrorResponse(err)
	}

	c.bindingNonce = nonce

	return ObjectResponse(map[string]string{
		"identity":  c.Identity.DID,
		"nonce":     nonce,
		"timestamp": nowString,
		"signature": signature,
	})
}

// bindACK process the client response of a binding process. It checks the nonce
// and the signature using owner DID.
func (c *Controller) bindACK(did string, ackParams BindACKParams) []byte {
	defer func() {
		c.bindingNonce = ""
	}()

	if did != c.ownerDID {
		return ErrorResponse(errors.New("illegal owner"))
	}

	if c.bindingNonce == "" {
		err := fmt.Errorf("binding request not found")
		log.WithError(err).Error("fail to bind account")
		return ErrorResponse(err)
	}

	if !key.VerifySignature(did, c.bindingNonce+ackParams.Timestamp, ackParams.Signature) {
		err := fmt.Errorf("invalid binding ack signature")
		log.WithError(err).Error("fail to bind account")
		return ErrorResponse(err)
	}

	if err := c.BindAccount(); err != nil {
		log.WithError(err).Error("fail to bind account")
		return ErrorResponse(err)
	}
	log.WithField("did", did).Println("bind account successfully")

	return ObjectResponse(map[string]string{
		"status": "ok",
	})
}

// bitcoinRPC runs bitcoind rpc for clients
func (c *Controller) bitcoinRPC(did string, bitcoindParams BitcoindRPCParams) []byte {
	bound, err := c.IsPodBound()
	if err != nil {
		return ErrorResponse(err)
	}

	if !bound {
		return ErrorResponse(fmt.Errorf("node is not bound"))
	}

	var reqBody bytes.Buffer

	if err := json.NewEncoder(&reqBody).Encode(bitcoindParams); err != nil {
		return ErrorResponse(err)
	}

	// log.WithField("reqBody", reqBody.String()).Info("requset body")
	req, err := http.NewRequest("POST", viper.GetString("bitcoind.rpcconnect"), &reqBody)
	if err != nil {
		return ErrorResponse(err)
	}
	req.SetBasicAuth(viper.GetString("bitcoind.rpcuser"), viper.GetString("bitcoind.rpcpassword"))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ErrorResponse(err)
	}
	defer resp.Body.Close()

	var responseBody json.RawMessage
	if resp.StatusCode != 401 {
		responseBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return ErrorResponse(err)
		}
	}

	return ObjectResponse(map[string]interface{}{
		"statusCode":   resp.StatusCode,
		"responseBody": responseBody,
	})
}
