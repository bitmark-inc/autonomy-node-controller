package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bitmark-inc/autonomy-pod-controller/key"
	"github.com/bitmark-inc/autonomy-pod-controller/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const PrivateKeyLen = 32

// KeyFile is an object to save private
type KeyFile struct {
	PrivateKey string `json:"private_key"`
}

// PodIdentity is an identity object
type PodIdentity struct {
	PrivateKey []byte
	DID        string
}

// Auth do authentication from the autonomy API server as the role autonomy-pod
// and get a JWT token for sending and receiving messages
func (p *PodIdentity) Auth() (string, error) {
	nowString := fmt.Sprint(int64(time.Now().UnixNano()) / int64(time.Millisecond))
	signature, err := key.Sign(p.PrivateKey, nowString+"autonomy-node")
	if err != nil {
		log.WithField("signature", signature).Error("sign error")
		return "", err
	}

	req := map[string]string{
		"signature": signature,
		"requester": p.DID,
		"timestamp": nowString,
		"role":      "autonomy-node",
	}

	var body bytes.Buffer

	if err := json.NewEncoder(&body).Encode(req); err != nil {
		return "", err
	}

	resp, err := http.Post(viper.GetString("messaging.endpoint")+"/api/auth", "application/json", &body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("auth api request fail. status: %d", resp.StatusCode)
	}

	var respBody struct {
		Token string `json:"jwt_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", err
	}

	return respBody.Token, nil
}

// NewPodIdentity creates a new identity
func NewPodIdentity() (*PodIdentity, error) {
	privateKey, err := utils.GenerateRandomBytes(PrivateKeyLen)
	if err != nil {
		return nil, err
	}

	return &PodIdentity{
		PrivateKey: privateKey,
		DID:        key.DID(privateKey),
	}, nil
}

// SaveKey saves the private key of a PodIdentity to a key file
func (p *PodIdentity) SaveKey(keyFile string) error {
	f, err := os.OpenFile(keyFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	k := KeyFile{
		PrivateKey: hex.EncodeToString(p.PrivateKey),
	}

	return json.NewEncoder(f).Encode(&k)
}

// LoadKey loads the private of a PodIdentity from a key file
func (p *PodIdentity) LoadKey(keyFile string) (*PodIdentity, error) {
	f, err := os.Open(keyFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var k KeyFile
	if err := json.NewDecoder(f).Decode(&k); err != nil {
		return nil, err
	}

	privateKey, err := hex.DecodeString(k.PrivateKey)
	if err != nil {
		return nil, err
	}

	p.PrivateKey = privateKey
	p.DID = key.DID(privateKey)
	return p, nil
}

// LoadPodIdentity loads the identity from a key file

// CreateOrLoadPodIdentityFromKey tries to create an identity by loading existing key file first.
// If the key file does not exist, it creates an identity by generating a new private key
// and saving it into the key file
func CreateOrLoadPodIdentityFromKey(keyFile string) (*PodIdentity, bool, error) {
	_, err := os.Stat(keyFile)
	if err != nil {
		if os.IsNotExist(err) {
			i, err := NewPodIdentity()
			if err != nil {
				return nil, false, err
			}
			if err := i.SaveKey(keyFile); err != nil {
				return nil, false, err
			}
			return i, true, nil
		}
	}

	if i, err := new(PodIdentity).LoadKey(keyFile); err != nil {
		return nil, false, err
	} else {
		return i, false, err
	}
}
