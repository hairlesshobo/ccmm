// =================================================================================
//
//		ccmm - https://www.foxhollow.cc/projects/ccmm/
//
//	 Connection Church Media Manager, aka ccmm, is a tool for managing all
//   aspects of produced media- initial import from removable media,
//   synchronization with clients and automatic data replication and backup
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
	"context"
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
	slog.Info("Started importer server on " + listen)
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

// HTTP middleware setting a value on the request context
func MyMiddleware(config model.ManagerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// create new context from `r` request context
			ctx := context.WithValue(r.Context(), model.ManagerConfigContext, config)

			// call the next handler in the chain, passing the response writer and
			// the updated request object with the new context value.
			//
			// note: context.Context values are nested, so any previously set
			// values will be accessible as well, and the new `"user"` key
			// will be accessible from this point forward.
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func setupRouting(config model.ManagerConfig) *chi.Mux {
	r := chi.NewRouter()

	r.Use(httplog.RequestLogger(getLogger(config)))

	r.Use(MyMiddleware(config))

	r.Get("/health", healthCheck)
	r.Get("/api/v1/quarters", getQuarters)
	r.Post("/api/v1/sync_request", syncRequest)

	return r
}

func getConfig(r *http.Request) model.ManagerConfig {
	return r.Context().Value(model.ManagerConfigContext).(model.ManagerConfig)
}
