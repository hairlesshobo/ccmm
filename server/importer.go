package server

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hairlesshobo/go-import-media/model"
)

func startImporter(importQueueChan chan model.ImportVolume, shutdownChan chan struct{}) {
out:
	for {
		select {
		// check for shutdown signal
		case <-shutdownChan:
			// if shutdown {
			slog.Info("Shutting down importer routine")
			break out
			// }
		case importConfig := <-importQueueChan:
			slog.Info("Queueing import job for " + importConfig.VolumePath)
			fmt.Printf("%+v\n", importConfig)
		default:
			// continue processing here
			// // Queue draw
		}

		time.Sleep(200 * time.Millisecond)
	}
}
