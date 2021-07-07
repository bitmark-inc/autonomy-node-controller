package bitcoind

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/spf13/viper"
)

// HttpBitcoind is a client to bitcoind
type HttpBitcoind struct {
	serverURL  string
	HttpClient *http.Client
	sequence   *uint32
}

func NewHttpRPCClient(client *http.Client) (*HttpBitcoind, error) {
	u, err := url.Parse(viper.GetString("bitcoind.rpcconnect"))
	if err != nil {
		return nil, err
	}

	bitcoindURL := fmt.Sprintf(
		"http://%s:%s@%s:%s",
		viper.GetString("bitcoind.rpcuser"),
		viper.GetString("bitcoind.rpcpassword"),
		u.Hostname(),
		u.Port(),
	)

	return &HttpBitcoind{
		serverURL:  bitcoindURL,
		HttpClient: client,
	}, nil
}

func NewBtcdRPCClient() (*rpcclient.Client, error) {

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
	return rpcclient.New(connCfg, nil)
}
