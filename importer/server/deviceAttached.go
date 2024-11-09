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
	"fmt"
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

func initDeviceAttachedThread(config model.ImporterConfig) {
	shutdownDeviceAttacherChan = make(chan struct{})
	deviceAttacherQueueChan = make(chan model.DeviceAttached, 10)

	go deviceAttachedRoutine(config, deviceAttacherQueueChan, shutdownDeviceAttacherChan)
}

func cleanupDeviceAttachedThread() {
	defer close(shutdownDeviceAttacherChan)
	defer close(deviceAttacherQueueChan)
}

func deviceAttachedPost(config model.ImporterConfig, w http.ResponseWriter, r *http.Request) {
	if config.DisableAutoProcessing {
		slog.Info("Auto processing is disabled by config file, taking no action")
		w.WriteHeader(403)
		return
	}

	attachDeviceConfig := util.ReadJsonBody[model.DeviceAttached](r)

	if !util.FileExists(attachDeviceConfig.DevicePath) {
		w.WriteHeader(500)
	} else {
		deviceAttacherQueueChan <- attachDeviceConfig
		w.WriteHeader(201)
	}

}

func deviceAttachedRoutine(config model.ImporterConfig, deviceAttachedQueueChan chan model.DeviceAttached, shutdownDeviceAttachedChan chan struct{}) {
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
			action.DeviceAttached(config, attachDeviceConfig)
		default:
		}

		time.Sleep(200 * time.Millisecond)
	}
}
