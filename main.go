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
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"nuts-foundation/nuts-monitor/api"
	"nuts-foundation/nuts-monitor/client"
	"nuts-foundation/nuts-monitor/config"
	"nuts-foundation/nuts-monitor/data"
	"os"
	"path"

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

	// then initialize the data storage and fill it with the initial transactions
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := data.NewStore()
	loadHistory(store, config)
	store.Start(ctx)

	// start the web server
	e := newEchoServer(config, store)

	// Start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", 1313)))
}

// loadHistory requests transactions from the Nuts node per 100 and stores them in the data store
func loadHistory(store *data.Store, c config.Config) {
	// initialize the client
	client := client.HTTPClient{
		Config: c,
	}

	// load the initial transactions
	// ListTransactions per batch of 100, stop if the list is empty
	// currentOffset is used to determine the offset for the next batch
	currentOffset := 0
	for {
		transactions, err := client.ListTransactions(context.Background(), currentOffset, currentOffset+100)
		if err != nil {
			log.Fatalf("failed to load historic transactions: %s", err)
		}
		if len(transactions) == 0 {
			break
		}
		// the transactions need to be converted from string to Transaction
		for _, stringTransactions := range transactions {
			transaction, err := data.FromJWS(stringTransactions)
			if err != nil {
				log.Printf("failed to parse transaction: %s", err)
			}
			store.Add(*transaction)
		}
		// increase offset for next batch
		currentOffset += 100
	}
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
