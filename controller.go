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

type Controller struct {
	ownerDID     string
	bindingNonce string
	bindingFile  string
	httpClient   *http.Client
	Identity     *PodIdentity
	store        Store
}

func NewController(ownerDID string, i *PodIdentity) *Controller {
	return &Controller{
		ownerDID:    ownerDID,
		bindingFile: config.AbsoluteApplicationFilePath(viper.GetString("binding_file")),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Identity: i,
		store:    NewBoltStore(config.AbsoluteApplicationFilePath(viper.GetString("db_name"))),
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

	if req.Command != "bind" && req.Command != "bind_ack" {
		bound, err := c.IsPodBound()
		if err != nil {
			return CommandResponse{ErrorResponse(err)}
		}

		if !bound {
			return CommandResponse{ErrorResponse(err)}
		}
	}

	if !c.HasRPCAccess(m.Source, req.Command) {
		return CommandResponse{ErrorResponse(errors.New("not allowed to use this RPC"))}
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
	case "create_wallet":
		var params CreateWalletRPCParams
		if err := json.Unmarshal(req.Args, &params); err != nil {
			return CommandResponse{ErrorResponse(fmt.Errorf("bad request for createwallet. error: %s", err.Error()))}
		}

		resp := c.createWallet(params.Descriptor)
		return CommandResponse{resp}
	case "set_member":
		var params UpdateMemberAccessModeRPCParams
		if err := json.Unmarshal(req.Args, &params); err != nil {
			return CommandResponse{ErrorResponse(fmt.Errorf("bad request for setmember. error: %s", err.Error()))}
		}

		resp := c.setMember(params.MemberDID, params.AccessMode)
		return CommandResponse{resp}
	case "remove_member":
		var params RemoveMemberAccessModeRPCParams
		if err := json.Unmarshal(req.Args, &params); err != nil {
			return CommandResponse{ErrorResponse(fmt.Errorf("bad request for removemember. error: %s", err.Error()))}
		}

		resp := c.removeMember(params.MemberDID)
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
	if !c.HasBitcoinRPCAccess(did, bitcoindParams.Method) {
		return ErrorResponse(errors.New("not allowed to use this RPC"))
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

// createWallet creates a new descriptor wallet if it does not exist
// according to the semi-finished descriptor (already including the platform and recovery key information).
// An example of an incomplete descriptor:
//
// wsh(sortedmulti(2,[119dbcab/48h/1h/0h/2h]tpubDFYr9xD4WtT3yDBdX2qT2j2v6ZruqccwPKFwLguuJL99bWBrk6D2Lv1aPpRbFnw1sQUU9DM7ScMAkPRJqR1iXKhWMBNMAJ45QCTuvSZbzzv/0/*,[e650dc93/48h/1h/0h/2h]tpubDEijNAeHVNmm6wHwspPv4fV8mRkoMimeVCk47dExpN9e17jFti12BdjzL8MX17GvKEekRzknNuDoLy1Q8fujYfsWfCvjwYmjjENUpzwDy6B/0/*,[<fingerprint>/48h/1h/0h/2h]<xpub>/0/*))
func (c *Controller) createWallet(incompleteDescriptor string) []byte {
	derivationPath := utils.ExtractGordianKeyDerivationPath(incompleteDescriptor)
	if derivationPath == "" {
		return ErrorResponse(fmt.Errorf("gordian key derivation path not found"))
	}

	path, err := utils.ParseDerivationPath(derivationPath)
	if err != nil {
		return ErrorResponse(err)
	}

	u, err := url.Parse(viper.GetString("bitcoind.rpcconnect"))
	if err != nil {
		return ErrorResponse(err)
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
		return ErrorResponse(err)
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
				shouldCreateWallet = true
			default:
				return ErrorResponse(err)
			}
		} else {
			return ErrorResponse(err)
		}
	} else {
		if walletInfo.KeyPoolSize == 0 && *walletInfo.KeyPoolSizeHDInternal == 0 {
			shouldImportDescriptors = true
		}
	}

	blockchainInfo, err := client.GetBlockChainInfo()
	if err != nil {
		return ErrorResponse(err)
	}
	keyFilePath := config.AbsoluteApplicationFilePath(viper.GetString("gordian_master_key_file"))
	masterKey, err := createOrLoadMasterKey(blockchainInfo.Chain, keyFilePath)
	if err != nil {
		return ErrorResponse(err)
	}
	masterFingerprint, err := computeFingerprint(masterKey)
	if err != nil {
		return ErrorResponse(err)
	}
	gordianPrivateKey := masterKey
	for _, i := range path {
		gordianPrivateKey, err = gordianPrivateKey.Derive(i)
		if err != nil {
			return ErrorResponse(err)
		}
	}
	gordianPublicKey, err := gordianPrivateKey.Neuter()
	if err != nil {
		return ErrorResponse(err)
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
			return ErrorResponse(err)
		}
	}

	if shouldImportDescriptors {
		importedDescriptorReplacer := strings.NewReplacer(
			"<fingerprint>", masterFingerprint,
			"<xpub>", gordianPrivateKey.String(),
		)
		externalDescriptorWithoutChecksum := importedDescriptorReplacer.Replace(incompleteDescriptor)
		externalDescriptorInfo, err := client.GetDescriptorInfo(externalDescriptorWithoutChecksum)
		if err != nil {
			return ErrorResponse(err)
		}

		internalDescriptorWithoutChecksum := strings.ReplaceAll(externalDescriptorWithoutChecksum, "/0/*", "/1/*")
		internalDescriptorInfo, err := client.GetDescriptorInfo(internalDescriptorWithoutChecksum)
		if err != nil {
			return ErrorResponse(err)
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
			return ErrorResponse(err)
		}
	}

	accountMapDescriptorReplacer := strings.NewReplacer(
		"<fingerprint>", masterFingerprint,
		"<xpub>", gordianPublicKey.String(),
	)
	gordianWalletDescriptor := accountMapDescriptorReplacer.Replace(incompleteDescriptor)
	return ObjectResponse(map[string]interface{}{"descriptor": gordianWalletDescriptor})
}

func (c *Controller) setMember(memberDID string, accessMode AccessMode) []byte {
	if err := c.store.UpdateMemberAccessMode(memberDID, accessMode); err != nil {
		return ErrorResponse(err)
	}
	return ObjectResponse(map[string]string{"status": "ok"})
}

func (c *Controller) removeMember(memberDID string) []byte {
	if err := c.store.RemoveMember(memberDID); err != nil {
		return ErrorResponse(err)
	}
	return ObjectResponse(map[string]string{"status": "ok"})
}
