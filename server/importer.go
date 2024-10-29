package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/hairlesshobo/go-import-media/action"
	"github.com/hairlesshobo/go-import-media/model"
	"github.com/hairlesshobo/go-import-media/util"
)

var (
	shutdownChan    chan struct{}
	importQueueChan chan model.ImportVolume
)

//
// private functions
//

func initImporterThread() {
	shutdownChan = make(chan struct{})
	importQueueChan = make(chan model.ImportVolume)

	go importerRoutine(importQueueChan, shutdownChan)
}

func cleanupImporterThread() {
	defer close(shutdownChan)
	defer close(importQueueChan)
}

func triggerImport(w http.ResponseWriter, r *http.Request) {
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

}

func importerRoutine(importQueueChan chan model.ImportVolume, shutdownChan chan struct{}) {
out:
	for {
		select {
		// check for shutdown signal
		case <-shutdownChan:
			slog.Info("Shutting down importer routine")
			break out

		// check for import request
		case importConfig := <-importQueueChan:
			slog.Info("Starting import job for " + importConfig.VolumePath)
			fmt.Printf("%+v\n", importConfig)
			action.Import(importConfig)
		default:
			// continue processing here
			// // Queue draw
		}

		time.Sleep(200 * time.Millisecond)
	}
}
