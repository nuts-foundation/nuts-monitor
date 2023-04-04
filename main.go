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
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"nuts-foundation/nuts-monitor/api"
	"nuts-foundation/nuts-monitor/client"
	"nuts-foundation/nuts-monitor/config"
	"os"
	"path"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const assetPath = "web"

//go:embed web/*
var embeddedFiles embed.FS

func main() {
	e := newEchoServer()

	// Start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", 1313)))
}

func newEchoServer() *echo.Echo {
	// config
	config := config.LoadConfig()
	config.Print(log.Writer())

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
		return http.FS(os.DirFS(assetPath))
	}

	log.Print("using embed mode")
	fsys, err := fs.Sub(embeddedFiles, path.Join(assetPath, "dist"))
	if err != nil {
		panic(err)
	}

	return http.FS(fsys)
}
