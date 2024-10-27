package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/hairlesshobo/go-import-media/model"
	"github.com/hairlesshobo/go-import-media/util"
)

func StartServer(listenPort int32) {
	shutdownChan := make(chan struct{})
	defer close(shutdownChan)

	importQueueChan := make(chan model.ImportVolume)
	defer close(importQueueChan)

	listen := fmt.Sprintf(":%d", listenPort)
	fmt.Println("starting server on " + listen)

	r := chi.NewRouter()

	go importer(importQueueChan, shutdownChan)

	r.Use(middleware.Logger)

	r.Post("/trigger_import", func(w http.ResponseWriter, r *http.Request) {
		var importConfig model.ImportVolume

		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Failed to read request body: " + err.Error())
		}

		if err = json.Unmarshal(body, &importConfig); err != nil {
			slog.Error("Failed to unmarshal JSON: " + err.Error())
		}

		fmt.Printf("%+v\n", importConfig)

		if !util.DirectoryExists(importConfig.VolumePath) {
			w.WriteHeader(500)
			// TODO: write an error response (define a model?)
		} else {
			importQueueChan <- importConfig
			w.WriteHeader(201)
		}

	})

	err := http.ListenAndServe(listen, r)

	if err != nil {
		slog.Error("http server error", "error", err)
		os.Exit(1)
	}
}
