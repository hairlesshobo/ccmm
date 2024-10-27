package server

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hairlesshobo/go-import-media/action"
	"github.com/hairlesshobo/go-import-media/model"
)

func importer(importQueueChan chan model.ImportVolume, shutdownChan chan struct{}) {
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
