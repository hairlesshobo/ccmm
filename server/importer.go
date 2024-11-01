// =================================================================================
//
//		gim - https://www.foxhollow.cc/projects/gim/
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
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"gim/action"
	"gim/model"
	"gim/util"
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
			// TODO: is this block even necessary?
			// continue processing here
			// // Queue draw
		}

		time.Sleep(200 * time.Millisecond)
	}
}
