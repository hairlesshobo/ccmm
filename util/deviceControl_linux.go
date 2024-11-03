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
//go:build linux

package util

import (
	"fmt"
	"log/slog"
	"strings"
)

func TestPlatform() {
	fmt.Println(UnmountVolume("/dev/sda1"))
	platformNotSupported(TestPlatform)
}

func GetVolumeName(mountPath string) string {
	slog.Debug(fmt.Sprintf("Querying volume name at '%s'", mountPath))
	command := "findmnt -n --output label --mountpoint %s0"
	output, _, err := callExternalCommand(command, mountPath)

	if err != nil {
		slog.Debug(fmt.Sprintf("Could not get volume name from path '%s'", mountPath))
		return ""
	}

	return strings.TrimSuffix(output, "\n")
}

func pathMounted(device string) bool {
	findmntCommand := "findmnt %s0"
	_, exitCode, _ := callExternalCommand(findmntCommand, device)

	// device is mounted
	return exitCode == 0
}

func MountVolume(device string) string {
	// lets first make sure that the device isn't already mounted elsewhere,
	// if it is, we'll use the path it is already mounted to
	// TODO: add retry logic
	findmntCommand := "findmnt -o target -n %s0"
	output, exitCode, _ := callExternalCommand(findmntCommand, device)

	if exitCode == 0 {
		existingPath := strings.TrimSuffix(output, "\n")
		slog.Error(fmt.Sprintf("Device is already mounted at: '%s', no need to mount", existingPath))
		return existingPath
	}

	slog.Info(fmt.Sprintf("Mounting device '%s'", device))
	command := "udisksctl mount --block-device %s0"
	output, exitCode, _ = callExternalCommand(command, device)

	if exitCode != 0 {
		// TODO: improve logging here to pull from command stderr output
		slog.Error(fmt.Sprintf("Failed to mount device '%s'", device))
		return ""
	}

	return strings.Split(strings.TrimSuffix(output, "\n"), " at ")[1]
}

func UnmountVolume(device string) bool {
	// test to see if the provided device exists and is even mounted
	// device is mounted
	if !pathMounted(device) {
		slog.Debug(fmt.Sprintf("Device is not mounted: '%s', nothing to do.", device))

		// we return true because the desire is for the provided device to be
		// unmounted, and it already is
		return true
	}

	// we use udisksctl to unmount because, supposedly, it does buffer flushing
	// and crap that might not get done otherwise. ¯\_(ツ)_/¯
	slog.Info(fmt.Sprintf("Unmounting device '%s'", device))
	command := "udisksctl unmount --block-device %s0"
	_, exitCode, _ := callExternalCommand(command, device)

	// 0 means it unmounted successfully
	return exitCode == 0
}

func GetVolumeFormat(device string) string {
	// this should look up the filesystem type of a given device.
	// This will allow us to filter what types of devices are
	// automatically mounted and scanned. I clearly won't be needing
	// to automount and import from ext4, xfs, etc.
	//
	// initial thoughts are only FAT-based disks should be imported
	// there may be a need for NTFS or AFS at some point, but i doubt-ish
	// it (maybe AFS for blackmagic gear)
	command := "blkid -o value -s TYPE %s0"
	output, exitCode, err := callExternalCommand(command, device)

	// testing err and exit code may be redundant, but whatever
	if err != nil || exitCode != 0 {
		return ""
	}

	return strings.TrimSuffix(output, "\n")
}

func PowerOffDevice(device string) bool {
	// this needs to be provided the path to the block device itself,
	// for example /dev/sda (without the partition index). will need to
	// look this value up. one way is this:
	//  readlink -f "/sys/class/block/sda1/.."
	// note 2: power off might work on the partitoin device, double check it

	// the goal is to ensure that the deviec specified is not powered on
	// if the file doesn't exist, we can say "yes - it is no longer powered on"
	if !FileExists(device) {
		slog.Debug(fmt.Sprintf("Device '%s' doesn't exist, already powered off?", device))
		return true
	}

	slog.Info(fmt.Sprintf("Powering off device '%s'", device))
	command := "udisksctl power-off --block-device %s0"
	_, _, err := callExternalCommand(command, device)

	return err == nil
}
