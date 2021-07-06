// SPDX-License-Identifier: ISC
// Copyright (c) 2019-2021 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

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
	"github.com/gin-gonic/gin"
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
	var retryInterval, retryCounts int

	flag.StringVar(&configFile, "c", "./config.yaml", "[optional] path of configuration file")
	flag.StringVar(&configFile, "config", "./config.yaml", "[optional] path of configuration file")
	flag.IntVar(&retryInterval, "retry", 10, "[optional] retry intervals between each messaging API")
	flag.IntVar(&retryCounts, "count", 12, "[optional] retry counts for messaging API")
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

	for c := 0; c < retryCounts; c++ {
		if c != 0 {
			time.Sleep(time.Duration(retryInterval) * time.Second)
		}

		if c == retryCounts-1 {
			log.WithField("retry", retryCounts).Fatal("maximum retries exceeded for pod authentication")
		}

		if err := i.Auth(); err != nil {
			log.WithError(err).Error("pod authentication fail")
		} else {
			break
		}
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
				log.WithError(err).Panic("fail to parse token")
			}

			if claims, ok := t.Claims.(*jwt.StandardClaims); ok {
				if time.Now().Unix() > (claims.ExpiresAt - int64(renewBefore)) {
					if err := i.Auth(); err != nil {
						log.WithError(err).Error("fail to refresh token")
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

	go func() {
		router := gin.New()
		router.POST("/tx-notification", controller.transactionNotify)
		router.Run(":" + viper.GetString("server_port"))
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

CONNECTION_LOOP:
	// This is the main loop for maintaining a persistant websocket connection to API server
	for {
		authToken := i.AuthToken()

		messagingClient := messaging.New(
			&http.Client{Timeout: 10 * time.Second},
			viper.GetString("messaging.endpoint"),
			authToken,
			config.AbsoluteApplicationFilePath(viper.GetString("messaging.db_name")))

		var ws *messaging.WSMessagingClient
		for i := 0; i < retryCounts; i++ {
			if i != 0 {
				time.Sleep(time.Duration(retryInterval) * time.Second)
			}
			if err := messagingClient.RegisterAccount(); err != nil {
				log.WithError(err).Error("registering account")
				continue
			}

			if err := messagingClient.RegisterKeys(); err != nil {
				log.WithError(err).Error("failed to register pre-keys on startup")
				continue
			}

			c, err := messagingClient.NewWSClient()
			if err != nil {
				log.WithError(err).Error("fail to establish websocket connection")
				continue
			}

			ws = c
			break
		}

		if ws == nil {
			log.WithField("retry", retryCounts).Fatalf("maximum retries exceeded for establishing the websocket connection")
		}

		addKey := make(chan struct{}, 0)
		go func(refillInterval time.Duration) {
			for range addKey {
				log.Debug("refill pre-keys")
				messagingClient.RefreshToken(i.AuthToken())

				if err := messagingClient.RegisterKeys(); err != nil {
					log.WithError(err).Fatalf("failed to refill pre-keys")
				}
				time.Sleep(refillInterval)
			}
			log.Debug("add-key loop closed")
		}(time.Second)

		// This is a loop to watch and process new messages
	MESSAGE_LOOP:
		for {
			select {
			case <-interrupt:
				log.Info("service interrupted")
				close(addKey)
				ws.Close()
				messagingClient.Close()
				break CONNECTION_LOOP
			case m := <-ws.WhisperMessages():
				// there will be a very last nil message if a connection is closed from the server
				if m == nil {
					log.Info("connection closed by server")
					close(addKey)
					ws.Close()
					messagingClient.Close()
					// break the MESSAGE_LOOP so that the service will start re-connecting the API server
					break MESSAGE_LOOP
				}

				log.WithField("message", m).Debug("receive message")
				responseMessage := controller.Process(m)
				if responseMessage != nil {
					ws.SendWhisperMessages(m.Source, m.SourceDevice, responseMessage)
				}

				select {
				case addKey <- struct{}{}:
				default:
				}
			}
		}
	}
}
