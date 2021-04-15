package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/bitmark-inc/autonomy-pod-controller/config"
	"github.com/bitmark-inc/autonomy-pod-controller/key"
	"github.com/bitmark-inc/autonomy-pod-controller/messaging"
	"github.com/bitmark-inc/secp256k1-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type BindResponse struct {
	Identity  string `json:"identity"`
	Nonce     string `json:"nonce"`
	Timestamp string `json:"timestamp"`
	Signature string `json:"signature"`
}

// Sign returns the signature of a message using the given private key
func Sign(privateKey []byte, message string) (string, error) {
	hash := sha256.Sum256([]byte(message))
	s, err := secp256k1.Sign(hash[:], privateKey)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(s), nil
}

// bind invokes a bind request from pod
func bind(wsClient *messaging.WSMessagingClient, respCh <-chan *messaging.Message, podDID string) (string, error) {
	wsClient.SendWhisperMessages(podDID, 0, [][]byte{[]byte(`{"id":"1","command":"bind"}`)})
	r := <-respCh

	var resp struct {
		Error string       `json:"error"`
		Data  BindResponse `json:"data"`
	}

	if err := json.Unmarshal(r.Content, &resp); err != nil {
		return "", err
	}
	log.WithField("resp", resp).Info("bind resp")
	if resp.Error != "" {
		return "", fmt.Errorf(resp.Error)
	}

	if !key.VerifySignature(podDID, resp.Data.Nonce+resp.Data.Timestamp, resp.Data.Signature) {
		return "", fmt.Errorf("invalid bind info")
	}
	return resp.Data.Nonce, nil
}

// bindACK responses a bind request to pod
func bindACK(wsClient *messaging.WSMessagingClient, respCh <-chan *messaging.Message, podDID, nonce string, privateKey []byte) error {
	nowString := fmt.Sprint(int64(time.Now().UnixNano()) / int64(time.Millisecond))
	signature, err := Sign(privateKey, nonce+nowString)
	if err != nil {
		return err
	}

	bindAckReq := map[string]interface{}{
		"id":      "test",
		"command": "bind_ack",
		"args": map[string]string{
			"timestamp": nowString,
			"signature": signature,
		},
	}
	log.WithField("bindAckReq", bindAckReq).Info("bind ack request")

	b, err := json.Marshal(bindAckReq)
	if err != nil {
		return err
	}

	wsClient.SendWhisperMessages(podDID, 0, [][]byte{b})

	r := <-respCh

	var resp struct {
		Error string                 `json:"error"`
		Data  map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(r.Content, &resp); err != nil {
		return err
	}
	log.WithField("resp", resp).Info("bind ack resp")

	if resp.Error != "" {
		return fmt.Errorf(resp.Error)
	}

	return nil
}

// bitcoinCommand sends bitcoind RPC requests to pod
func bitcoinCommand(wsClient *messaging.WSMessagingClient, respCh <-chan *messaging.Message, podDID, jsonRPCBody string) (json.RawMessage, error) {
	req := map[string]interface{}{
		"id":      "test",
		"command": "bitcoind",
		"args":    json.RawMessage(jsonRPCBody),
	}

	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	wsClient.SendWhisperMessages(podDID, 0, [][]byte{b})
	r := <-respCh

	var bitcoindResp json.RawMessage
	if err := json.Unmarshal(r.Content, &bitcoindResp); err != nil {
		return nil, err
	}

	return bitcoindResp, nil
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "./config.yaml", "[optional] path of configuration file")
	flag.StringVar(&configFile, "config", "./config.yaml", "[optional] path of configuration file")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: binding-tool [options] [command]\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "Command:\n")
		fmt.Fprintf(os.Stderr, "  bind \t\t\t\t\t initiate a bind process\n")
		fmt.Fprintf(os.Stderr, "  bind_ack [nonce] \t\t\t respond a bind ack with a given nonce\n")
		fmt.Fprintf(os.Stderr, "  bitcoind [JSONRPC request body] \t call bitcoind command\n")
	}
	flag.Parse()

	commands := flag.Args()
	if len(commands) == 0 {
		flag.Usage()
		return
	}

	config.LoadConfig(configFile)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	messagingClient := messaging.New(http.DefaultClient,
		viper.GetString("messaging.server_url"),
		viper.GetString("messaging.client_jwt"),
		viper.GetString("messaging.store.leveldb_path"),
	)

	if err := messagingClient.RegisterAccount(); err != nil {
		log.WithError(err).Fatalf("registering account, error")
	}

	if err := messagingClient.RegisterKeys(); err != nil {
		log.WithError(err).Fatalf("failed to register keys")
	}

	wsClient, err := messagingClient.NewWSClient()
	if err != nil {
		log.WithError(err).Panic("fail to start a websocket connection")
	}
	defer wsClient.Close()

	done := make(chan struct{})

	msgCh := wsClient.WhisperMessages()

	privateKey, err := hex.DecodeString("61900f1b361bda95b318a219efda0685c9209f703f8638c269506860271c26d9")
	if err != nil {
		panic("unable to read private key")
	}

	go func() {
		podDID := viper.GetString("pod.identity")

		switch commands[0] {
		case "bind":
			nonce, err := bind(wsClient, msgCh, podDID)
			if err != nil {
				log.WithError(err).Error("bind request fail")
				os.Exit(1)
			}
			log.WithField("podDID", podDID).WithField("nonce", nonce).Info("bind ok")
		case "bind_ack":
			if len(commands) != 2 {
				flag.Usage()
				break
			}
			if err := bindACK(wsClient, msgCh, podDID, commands[1], privateKey); err != nil {
				log.WithError(err).Error("bind ack fail")
				os.Exit(1)
			}
			log.Info("bind ack ok")
		case "bitcoind":
			if len(commands) != 2 {
				flag.Usage()
				break
			}

			resp, err := bitcoinCommand(wsClient, msgCh, podDID, commands[1])
			if err != nil {
				log.WithError(err).Panic("bitcoin request fail")
				os.Exit(1)
			}

			log.WithField("resp", string(resp)).Info("bitcoin response")
		default:
			flag.Usage()
		}

		os.Exit(0)
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			wsClient.Close()

			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
