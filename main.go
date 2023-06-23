/*
 * Copyright (C) 2023 Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"io/fs"
	"log"
	"net/http"
	"nuts-foundation/nuts-monitor/api"
	"nuts-foundation/nuts-monitor/client"
	"nuts-foundation/nuts-monitor/config"
	"nuts-foundation/nuts-monitor/data"
	"os"
	"path"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const assetPath = "web"

//go:embed web/*
var embeddedFiles embed.FS

func main() {
	// first load the config
	config := config.LoadConfig()
	config.Print(log.Writer())

	// create the Node API Client
	client := client.HTTPClient{
		Config: config,
	}

	// then initialize the data storage and fill it with the initial transactions
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := data.NewStore(client)
	// connect to the NATS stream of the nuts node
	startConsumer(ctx, store, config)
	// load history async
	loadHistory(ctx, store, config)
	// start shifting windows
	store.Start(ctx)

	// start the web server
	e := newEchoServer(config, store)

	// Start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", 1313)))
}

// loadHistory uses a Go routine to load the transactions in the background
// On error it will retry every 10 seconds
func loadHistory(context context.Context, store *data.Store, c config.Config) {
	// initialize the client
	client := client.HTTPClient{
		Config: c,
	}

	go func() {
		// As long there's no error, keep retrying
		for {
			err := loadHistoryOnce(store, client)
			select {
			case <-context.Done():
				return
			default:
				if err == nil {
					return
				}
				log.Printf("failed to load historic transactions: %s", err)
				log.Printf("retrying in 10 seconds")
				// sleep for 10 seconds
				<-time.After(10 * time.Second)
			}
		}
	}()
}

// loadHistoryOnce loads the transactions from the Nuts node and stores them in the data store
func loadHistoryOnce(store *data.Store, client client.HTTPClient) error {
	// load the initial transactions
	// ListTransactions per batch of 100, stop if the list is empty
	// currentOffset is used to determine the offset for the next batch
	currentOffset := 0
	for {
		transactions, err := client.ListTransactions(context.Background(), currentOffset, currentOffset+100)
		if err != nil {
			return err
		}
		if len(transactions) == 0 {
			break
		}
		// the transactions need to be converted from string to Transaction
		for _, stringTransaction := range transactions {
			transaction, err := data.FromJWS(stringTransaction)
			if err != nil {
				log.Printf("failed to parse transaction: %s", err)
			}
			store.Add(*transaction)
		}
		// increase offset for next batch
		currentOffset += 100
	}
	return nil
}

// startConsumer will try to subscribe to NATS every 10 seconds
// it will retry until it succeeds
func startConsumer(ctx context.Context, store *data.Store, c config.Config) {
	go func() {
		for {
			err := startConsumerOnce(ctx, store, c)
			select {
			case <-ctx.Done():
				return
			default:
				if err == nil {
					return
				}
				log.Printf("failed to start NATS consumer: %s", err)
				log.Printf("retrying in 10 seconds")
				// sleep for 10 seconds
				<-time.After(10 * time.Second)
			}
		}
	}()
}

// startConsumerOnce starts the NATS consumer
// it creates a non-durable subscription to the nuts node for the "nuts-disposable" stream
// and stores the transactions in the data store
func startConsumerOnce(ctx context.Context, store *data.Store, c config.Config) error {
	// create a NATS connection
	conn, err := nats.Connect(c.NutsNodeStreamAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS stream: %w", err)
	}
	// setup NATS JetStream
	js, err := conn.JetStream()
	if err != nil {
		return fmt.Errorf("failed to connect to JetStream: %w", err)
	}

	// stream creation
	_, err = js.StreamInfo("nuts-monitor")
	if errors.Is(err, nats.ErrStreamNotFound) {
		_, err = js.AddStream(&nats.StreamConfig{
			Name:      "nuts-monitor",
			Subjects:  []string{"TRANSACTIONS.*"},
			MaxMsgs:   1000, // max buffer
			Retention: nats.LimitsPolicy,
			Storage:   nats.MemoryStorage,
			Discard:   nats.DiscardOld,
		})
		if err != nil {
			return fmt.Errorf("failed to create stream: %w", err)
		}
	} else if err != nil {
		return err
	}

	// Subscriber options
	opts := []nats.SubOpt{
		nats.BindStream("nuts-monitor"),
		nats.DeliverNew(),
		nats.Context(ctx),
	}

	// subscribe through JetStream
	_, err = js.Subscribe("TRANSACTIONS.*", func(msg *nats.Msg) {
		// parse the transaction, it's in JSON format
		event := transactionEvent{}
		err := json.Unmarshal(msg.Data, &event)
		if err != nil {
			log.Printf("failed to parse transaction event: %s", err)
		}
		transaction, err := data.FromJWS(event.Transaction)
		if err != nil {
			log.Printf("failed to parse transaction: %s", err)
		}
		// add transaction to store
		store.Add(*transaction)
	}, opts...)

	if err != nil {
		return fmt.Errorf("failed to subscribe to stream: %w", err)
	}
	return nil
}

type transactionEvent struct {
	// Transaction is in compacted JWS format
	Transaction string `json:"transaction"`
	// Payload is base64
	Payload string `json:"payload"`
}

func newEchoServer(config config.Config, store *data.Store) *echo.Echo {
	// http server
	e := echo.New()
	e.HideBanner = true
	loggerConfig := middleware.DefaultLoggerConfig
	e.Use(middleware.LoggerWithConfig(loggerConfig))

	// add status endpoint
	e.GET("/status", func(context echo.Context) error {
		return context.String(http.StatusOK, "OK")
	})

	// API endpoints from OAS spec
	apiWrapper := api.Wrapper{
		Config: config,
		Client: client.HTTPClient{
			Config: config,
		},
		DataStore: store,
	}
	api.RegisterHandlers(e, api.NewStrictHandler(apiWrapper, []api.StrictMiddlewareFunc{}))

	// Setup asset serving:
	// Check if we use live mode from the file system or using embedded files
	useFS := len(os.Args) > 1 && os.Args[1] == "live"
	assetHandler := http.FileServer(getFileSystem(useFS))
	e.GET("/*", echo.WrapHandler(assetHandler))

	return e
}

func getFileSystem(useFS bool) http.FileSystem {
	if useFS {
		log.Print("using live mode")
		return http.FS(os.DirFS(path.Join(assetPath, "dist")))
	}

	log.Print("using embed mode")
	fsys, err := fs.Sub(embeddedFiles, path.Join(assetPath, "dist"))
	if err != nil {
		panic(err)
	}

	return http.FS(fsys)
}
