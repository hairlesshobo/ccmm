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
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"ccmm/importer/action"
	"ccmm/model"
	"ccmm/util"
)

var (
	shutdownDeviceAttacherChan chan struct{}
	deviceAttacherQueueChan    chan model.DeviceAttached
)

//
// private functions
//

func initDeviceAttachedThread() {
	shutdownDeviceAttacherChan = make(chan struct{})
	deviceAttacherQueueChan = make(chan model.DeviceAttached)

	go deviceAttachedRoutine(deviceAttacherQueueChan, shutdownDeviceAttacherChan)
}

func cleanupDeviceAttachedThread() {
	defer close(shutdownDeviceAttacherChan)
	defer close(deviceAttacherQueueChan)
}

func triggerDeviceAttached(w http.ResponseWriter, r *http.Request) {
	if model.Config.DisableAutoProcessing {
		slog.Info("Auto processing is disabled by config file, taking no action")
		w.WriteHeader(403)
		return
	}

	var attachDeviceConfig model.DeviceAttached

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read request body: " + err.Error())
	}

	if err = json.Unmarshal(body, &attachDeviceConfig); err != nil {
		slog.Error("Failed to unmarshal JSON: " + err.Error())
	}

	fmt.Printf("%+v\n", attachDeviceConfig)

	if !util.FileExists(attachDeviceConfig.DevicePath) {
		w.WriteHeader(500)
	} else {
		// TODO: how to make this non-blocking and add to a queue to work off of
		deviceAttacherQueueChan <- attachDeviceConfig
		w.WriteHeader(201)
	}

}

func deviceAttachedRoutine(deviceAttachedQueueChan chan model.DeviceAttached, shutdownDeviceAttachedChan chan struct{}) {
out:
	for {
		select {
		// check for shutdown signal
		case <-shutdownDeviceAttachedChan:
			slog.Info("Shutting down device attached routine")
			break out

		// check for attach device request
		case attachDeviceConfig := <-deviceAttachedQueueChan:
			slog.Info("Starting device attached job for " + attachDeviceConfig.DevicePath)
			fmt.Printf("%+v\n", attachDeviceConfig)
			action.DeviceAttached(attachDeviceConfig)
		default:
			// TODO: is this block even necessary?
			// continue processing here
			// // Queue draw
		}

		time.Sleep(200 * time.Millisecond)
	}
}
