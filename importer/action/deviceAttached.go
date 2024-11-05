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

package action

import (
	"fmt"
	"log/slog"
	"time"

	"ccmm/model"
	"ccmm/util"
)

var (
	mountRetries          = 5
	mountRetryWaitSeconds = 5
)

// DeviceAttached Perform the woek necessary for a newly attached device. This
// will automaticlly mount the device, call an import, then unmount and power off the device
func DeviceAttached(config model.ImporterConfig, params model.DeviceAttached) {
	slog.Info(fmt.Sprintf("Handle device attachment for '%s'", params.DevicePath))

	mountedPath := ""

	for i := 1; i <= mountRetries; i++ {
		mountedPath = util.MountVolume(params.DevicePath)
		if mountedPath != "" {
			break
		}

		slog.Info(fmt.Sprintf("Failed to unmount device '%s', waiting %d seconds and trying again [attempt %d/%d",
			params.DevicePath, mountRetryWaitSeconds, i, mountRetries))
		time.Sleep(time.Duration(mountRetryWaitSeconds) * time.Second)
	}

	var importConfig model.ImportVolume
	importConfig.DryRun = params.DryRun
	importConfig.VolumePath = mountedPath

	Import(config, importConfig, func(_ *ImportQueueItem) {
		// TODO: do empty, if enabled

		for i := 1; i <= mountRetries; i++ {
			success := util.UnmountVolume(params.DevicePath)
			if success {
				break
			}

			slog.Info(fmt.Sprintf("Failed to unmount device '%s', waiting %d seconds and trying again [attempt %d/%d",
				params.DevicePath, mountRetryWaitSeconds, i, mountRetries))
			time.Sleep(time.Duration(mountRetryWaitSeconds) * time.Second)
		}
		util.PowerOffDevice(params.DevicePath)

		slog.Info(fmt.Sprintf("Finished device attachment for '%s'", params.DevicePath))
	})
}
