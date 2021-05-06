package main

import (
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
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

// renewBefore: 10 minutes in seconds
var renewBefore = 10 * time.Minute / time.Second

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "./config.yaml", "[optional] path of configuration file")
	flag.StringVar(&configFile, "config", "./config.yaml", "[optional] path of configuration file")
	flag.Parse()

	config.LoadConfig(configFile)

	i, created, err := CreateOrLoadPodIdentityFromKey(config.AbsoluteApplicationFilePath(viper.GetString("auth_key_file")))
	if err != nil {
		log.WithError(err).Panic("fail to create or load identity")
	}

	ownerDID := viper.GetString("owner_did")
	controller := NewController(ownerDID, i)
	log.WithField("owner_did", ownerDID).
		WithField("identity", i.DID).
		WithField("created", created).
		Info("controller initialized")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	if err := i.Auth(); err != nil {
		log.WithError(err).Panic("pod authentication fail")
	}

	// The goroutine will continuously check auth_token and re-request a new one if
	// a token is going to be expired.
	go func(checkInterval time.Duration) {
		for {
			token := i.AuthToken()
			if token == "" {
				// wait for the first token to be acquired from the first authentication
				time.Sleep(time.Second)
				continue
			}

			// parse the jwt claim without verified the token signature
			t, _, err := new(jwt.Parser).ParseUnverified(token, &jwt.StandardClaims{})
			if err != nil {
				log.WithError(err).Panic("fail to refresh token")
			}

			if claims, ok := t.Claims.(*jwt.StandardClaims); ok {
				if time.Now().Unix() > (claims.ExpiresAt - int64(renewBefore)) {
					if err := i.Auth(); err != nil {
						log.WithError(err).Panic("fail to refresh token")
					}
					log.Info("successfully refresh a new token")
				} else {
					log.WithField("expires_at", claims.ExpiresAt).Debug("token not expired")
				}
				time.Sleep(checkInterval)
				continue
			} else {
				log.WithError(errors.New("error casting token claims")).Panic("fail to refresh token")
			}
		}
	}(time.Minute)

CONNECTION_LOOP:
	// This is the main loop for maintaining a persistant websocket connection to API server
	for {
		authToken := i.AuthToken()

		messagingClient := messaging.New(
			&http.Client{Timeout: 10 * time.Second},
			viper.GetString("messaging.endpoint"),
			authToken,
			config.AbsoluteApplicationFilePath(viper.GetString("messaging.db_name")))

		if err := messagingClient.RegisterAccount(); err != nil {
			log.Fatalf("registering account, error: %s", err)
		}

		if err := messagingClient.RegisterKeys(); err != nil {
			log.WithError(err).Fatalf("failed to register pre-keys on startup")
		}

		ws, err := messagingClient.NewWSClient()
		if err != nil {
			log.WithError(err).Panic("fail to establish websocket connection")
		}

		// This is a loop to watch and process new messages
		for {
			select {
			case <-interrupt:
				log.Info("service interrupted")
				ws.Close()
				break CONNECTION_LOOP
			case m := <-ws.WhisperMessages():
				// there will be a very last nil message if a connection is closed from the server
				if m == nil {
					log.Info("connection closed by server")
					ws.Close()
					break CONNECTION_LOOP
				}

				log.WithField("message", m).Debug("receive message")
				responseMessage := controller.Process(m)
				if responseMessage != nil {
					ws.SendWhisperMessages(m.Source, m.SourceDevice, responseMessage)
				}
			}
		}
	}
}
