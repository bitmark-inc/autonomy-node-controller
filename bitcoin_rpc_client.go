package main

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/spf13/viper"
)

type rpcClient interface {
	Shutdown()
	GetWalletInfo() (*btcjson.GetWalletInfoResult, error)
	GetBlockChainInfo() (*btcjson.GetBlockChainInfoResult, error)
	RawRequest(method string, params []json.RawMessage) (json.RawMessage, error)
	GetDescriptorInfo(descriptor string) (*btcjson.GetDescriptorInfoResult, error)
	WalletProcessPsbt(psbt string, sign *bool, sighashType rpcclient.SigHashType, bip32Derivs *bool) (*btcjson.WalletProcessPsbtResult, error)
}

func NewBitcoinRPCClient() (rpcClient, error) {
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
	return client, nil
}
