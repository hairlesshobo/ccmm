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
//go:build darwin

package util

import (
	"bufio"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
)

func TestPlatform() {
	// WatchForDeviceMount(func(devicePath, volumePath string) {
	// 	fmt.Println("device mounted!")
	// 	fmt.Println("Device path: " + devicePath)
	// 	fmt.Println("Volume path: " + volumePath)
	// })

	fmt.Println(GetVolumeFormat("/Volumes/X32-LEX120"))
	// platformNotSupported(TestPlatform)
}

func GetVolumeName(mountPath string) string {
	label := filepath.Base(mountPath)
	return label
}

func MountVolume(device string) string {
	panic("Not implemented")
}

func UnmountVolume(path string) bool {
	// test to see if the provided device exists and is even mounted
	// device is mounted
	if !pathMounted(path) {
		slog.Debug(fmt.Sprintf("Device is not mounted: '%s', nothing to do.", path))

		// we return true because the desire is for the provided device to be
		// unmounted, and it already is
		return true
	}

	slog.Info(fmt.Sprintf("Unmounting device '%s'", path))
	command := "diskutil unmount %s0"
	_, exitCode, _ := callExternalCommand(command, path)

	// 0 means it unmounted successfully
	return exitCode == 0
}

func GetVolumeFormat(mountPath string) string {
	command := "diskutil info %s0"
	output, exitCode, _ := callExternalCommand(command, mountPath)

	if exitCode != 0 {
		return ""
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "File System Personality:") {
			format := strings.TrimSpace(strings.Split(line, ":")[1])

			if format == "ExFAT" {
				return ExFAT
			}

			if format == "MS-DOS FAT32" {
				return FAT32
			}

			return ""
		}
	}

	return ""
}

func PowerOffDevice(device string) bool {
	// test to see if the provided device exists
	if !FileExists(device) {
		slog.Debug(fmt.Sprintf("Device does not exist: '%s', nothing to do.", device))

		// we return true because the desire is for the provided device to be ejected,
		// and it already is (or the wrong one was provided, but never mind that)
		return true
	}

	slog.Info(fmt.Sprintf("Ejecting device '%s'", device))
	command := "diskutil eject %s0"
	_, exitCode, _ := callExternalCommand(command, device)

	// 0 means it unmounted successfully
	return exitCode == 0
}

func WatchForDeviceMount(deviceMountedCallback func(devicePath string, volumePath string)) {
	// run diskutil actiity and watch for "***DiskDescriptionChanged"
	cmd := exec.Command("diskutil", "activity")
	stdout, err := cmd.StdoutPipe()
	// cmd.Stderr = cmd.Stdout
	if err != nil {
		slog.Error("Error occurred running 'diskutil activity' command: " + err.Error())
		return
	}

	if err = cmd.Start(); err != nil {
		slog.Error("Error occurred starting 'diskutil activity' command: " + err.Error())
		return
	}

	watchStarted := false

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		if !watchStarted {
			if strings.HasPrefix(line, "***DAIdle") {
				watchStarted = true
			}
			continue
		}

		// this is apparently what is logged when a disk has been mounted
		if strings.HasPrefix(line, "***DiskDescriptionChanged") {
			if !strings.Contains(line, "file://") {
				// this happens when a disk is unmounted, so we need to ignore these
				continue
			}
			// starting line:
			// ***DiskDescriptionChanged ('disk7s1', DAVolumePath = 'file:///Volumes/CANON/') Time=20241108-16:25:39.9676

			mountLog := strings.TrimPrefix(line, "***DiskDescriptionChanged (")          // 'disk7s1', DAVolumePath = 'file:///Volumes/CANON/') Time=20241108-16:25:39.9676
			mountLog = strings.Split(mountLog, "') Time=")[0]                            // 'disk7s1', DAVolumePath = 'file:///Volumes/CANON/
			devicePath := "/dev/" + strings.Trim(strings.Split(mountLog, ",")[0], "'")   // /dev/disk7s1
			volumePath := strings.TrimSuffix(strings.Split(mountLog, "file://")[1], "/") // /Volumes/CANON

			deviceMountedCallback(devicePath, volumePath)
		}
	}
}

func pathMounted(path string) bool {
	findmntCommand := "diskutil info %s0"
	_, exitCode, _ := callExternalCommand(findmntCommand, path)

	// device is mounted
	return exitCode == 0
}
