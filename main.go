package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/bitmark-inc/autonomy-pod-controller/config"
	"github.com/bitmark-inc/autonomy-pod-controller/messaging"
)

type RequestCommand struct {
	ID      string          `json:"id"`
	Command string          `json:"command"`
	Args    json.RawMessage `json:"args"`
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "./config.yaml", "[optional] path of configuration file")
	flag.StringVar(&configFile, "config", "./config.yaml", "[optional] path of configuration file")
	flag.Parse()

	config.LoadConfig(configFile)

	i, created, err := CreateOrLoadPodIdentityFromKey(viper.GetString("keyfile"))
	if err != nil {
		log.WithError(err).Panic("fail to create or load identity")
	}

	ownerDID := viper.GetString("owner_did")
	controller := NewController(ownerDID, i)
	log.WithField("owner_did", ownerDID).
		WithField("identity", i.DID).
		WithField("created", created).
		Info("controller initialized")

	jwt, err := i.Auth()
	if err != nil {
		log.WithError(err).Panic("pod authentication fail")
	}

	messagingClient := messaging.New(&http.Client{
		Timeout: 10 * time.Second,
	}, viper.GetString("messaging.endpoint"), jwt, viper.GetString("messaging.db_path"))

	if err := messagingClient.RegisterAccount(); err != nil {
		log.Fatalf("registering account, error: %s", err)
	}

	if err := messagingClient.RegisterKeys(); err != nil {
		log.WithError(err).Fatalf("failed to register keys")
	}

	ws, err := messagingClient.NewWSClient()
	if err != nil {
		log.WithError(err).Panic("fail to establish websocket connection")
	}

	for m := range ws.WhisperMessages() {
		log.WithField("message", m).Debug("receive message")
		responseMessage := controller.Process(m)
		if responseMessage != nil {
			ws.SendWhisperMessages(m.Source, m.SourceDevice, responseMessage)
		}
	}
}
