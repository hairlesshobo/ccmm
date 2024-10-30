package action

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hairlesshobo/go-import-media/model"
	"github.com/hairlesshobo/go-import-media/util"
)

var (
	unmountRetries          int = 5
	unmountRetryWaitSeconds int = 5
)

func DeviceAttached(params model.DeviceAttached) bool {
	slog.Info(fmt.Sprintf("Handle device attachment for '%s'", params.DevicePath))

	mountedPath := util.MountVolume(params.DevicePath)

	var importConfig model.ImportVolume
	importConfig.DryRun = params.DryRun
	importConfig.VolumePath = mountedPath

	Import(importConfig)

	// TODO: do empty, if enabled

	for i := 1; i <= unmountRetries; i++ {
		success := util.UnmountVolume(params.DevicePath)
		if success {
			break
		}

		slog.Info(fmt.Sprintf("Failed to unmount device '%s', waiting %d seconds and trying again [attempt %d/%d",
			params.DevicePath, unmountRetryWaitSeconds, i, unmountRetries))
		time.Sleep(time.Duration(unmountRetryWaitSeconds) * time.Second)
	}
	util.PowerOffDevice(params.DevicePath)

	slog.Info(fmt.Sprintf("Finished device attachment for '%s'", params.DevicePath))

	return true
}
