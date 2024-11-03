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

func StartServer(config model.ManagerConfig) {
	// initImporterThread()
	// initDeviceAttachedThread()

	router := setupRouting(config)
	startServer(config, router)

	// cleanupImporterThread()
	// cleanupDeviceAttachedThread()
}

//
// private functions
//

func startServer(config model.ManagerConfig, router *chi.Mux) {
	listen := fmt.Sprintf("%s:%d", config.ListenAddress, config.ListenPort)
	slog.Info("Started gim server on " + listen)
	err := http.ListenAndServe(listen, router)

	if err != nil {
		slog.Error("http server error", "error", err)
		os.Exit(1)
	}
}

func getLogger(config model.ManagerConfig) *httplog.Logger {
	logger := httplog.NewLogger("server.listener", httplog.Options{
		LogLevel:       slog.Level(config.LogLevel),
		Concise:        true,
		RequestHeaders: false,
	})

	logger.Logger = slog.Default().With(slog.String("component", "router"))

	return logger
}

func setupRouting(config model.ManagerConfig) *chi.Mux {
	router := chi.NewRouter()

	router.Use(httplog.RequestLogger(getLogger(config)))

	router.Get("/health", healthCheck)
	router.Get("/quarters", func(w http.ResponseWriter, r *http.Request) {
		getQuarters(config, w, r)
	})
	// router.Post("/trigger_import", triggerImport)
	// router.Post("/device_attached", triggerDeviceAttached)

	return router
}
