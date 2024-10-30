package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/hairlesshobo/go-import-media/model"
)

//
// public functoins
//

func StartServer(listenAddress string, listenPort int32) {
	initImporterThread()
	initDeviceAttachedThread()

	startServer(listenAddress, listenPort, setupRouting())

	cleanupImporterThread()
	cleanupDeviceAttachedThread()
}

//
// private functions
//

func startServer(listenAddress string, listenPort int32, router *chi.Mux) {
	listen := fmt.Sprintf("%s:%d", listenAddress, listenPort)
	slog.Info("starting server on " + listen)
	err := http.ListenAndServe(listen, router)

	if err != nil {
		slog.Error("http server error", "error", err)
		os.Exit(1)
	}
}

func getLogger() *httplog.Logger {
	logger := httplog.NewLogger("server.listener", httplog.Options{
		LogLevel:       slog.Level(model.Config.LogLevel),
		Concise:        true,
		RequestHeaders: false,
	})

	logger.Logger = slog.Default().With(slog.String("component", "router"))

	return logger
}

func setupRouting() *chi.Mux {
	router := chi.NewRouter()

	router.Use(httplog.RequestLogger(getLogger()))

	router.Get("/health", healthCheck)
	router.Post("/trigger_import", triggerImport)
	router.Post("/device_attached", triggerDeviceAttached)

	return router
}
