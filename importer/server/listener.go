// =================================================================================
//
//		ccmm - https://www.foxhollow.cc/projects/ccmm/
//
//	 go-import-media, aka gim, is a tool for automatically importing media
//	 from removable disks into a predefined folder structure automatically.
//
//		Copyright (c) 2024 Steve Cross <flip@foxhollow.cc>
//
//		Licensed under the Apache License, Version 2.0 (the "License");
//		you may not use this file except in compliance with the License.
//		You may obtain a copy of the License at
//
//		     http://www.apache.org/licenses/LICENSE-2.0
//
//		Unless required by applicable law or agreed to in writing, software
//		distributed under the License is distributed on an "AS IS" BASIS,
//		WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//		See the License for the specific language governing permissions and
//		limitations under the License.
//
// =================================================================================

package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"ccmm/model"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
)

//
// public functoins
//

func StartServer(config model.ImporterConfig, listenAddress string, listenPort int32) {
	initDeviceAttachedThread(config)

	startServer(listenAddress, listenPort, setupRouting(config))

	cleanupDeviceAttachedThread()
}

//
// private functions
//

func startServer(listenAddress string, listenPort int32, router *chi.Mux) {
	listen := fmt.Sprintf("%s:%d", listenAddress, listenPort)
	slog.Info("Started ccmm importer server on " + listen)
	err := http.ListenAndServe(listen, router)

	if err != nil {
		slog.Error("http server error", "error", err)
		os.Exit(1)
	}
}

func getLogger(config model.ImporterConfig) *httplog.Logger {
	logger := httplog.NewLogger("server.listener", httplog.Options{
		LogLevel:       slog.Level(config.LogLevel),
		Concise:        true,
		RequestHeaders: false,
	})

	logger.Logger = slog.Default().With(slog.String("component", "router"))

	return logger
}

func setupRouting(config model.ImporterConfig) *chi.Mux {
	router := chi.NewRouter()

	router.Use(httplog.RequestLogger(getLogger(config)))

	router.Get("/health", healthCheck)
	router.Post("/trigger_import", func(w http.ResponseWriter, r *http.Request) {
		importPost(config, w, r)
	})
	router.Post("/device_attached", func(w http.ResponseWriter, r *http.Request) {
		deviceAttachedPost(config, w, r)
	})
	// TODO: add /status route

	return router
}
