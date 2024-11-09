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
//go:build linux

package util

import (
	"fmt"
	"log/slog"
	"strings"
)

// TestPlatform is really intended only for development purposes while
// building platform-specific functionality. Shouldn't be used for anything
// else and will eventually be removed
func TestPlatform() {
	// FSTYPE
	// FSVER
	fmt.Println(GetVolumeFormat("/media/flip/CANON"))
	platformNotSupported(TestPlatform)
}

// GetVolumeName requires a path to a mounted volume
// and will return the name of the volume as a string
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

// MountVolume requires a device node (ex: /dev/sda1) be provided
// and will either return an empty string on failure, or will
// return the path to the newly mounted volume. The mounted path
// is automatically determined by udisksctl and we have no control
// over where it chooses to mount
func MountVolume(device string) string {
	// lets first make sure that the device isn't already mounted elsewhere,
	// if it is, we'll use the path it is already mounted to
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

// UnmountVolume requires a device node (ex: /dev/sda1) be passed
// and will unmount the volume. Returns true on success or false if
// the unmount process failed
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

// GetVolumeFormat requires that a mount path is provided (ex: /media/user/CANON)
// and will return the format of the mounted filesystem at that point.
// If an error occurs or an unknown format is mounted there, an empty
// string will be returned instead
func GetVolumeFormat(mountPath string) string {
	// this should look up the filesystem type of a given device.
	// This will allow us to filter what types of devices are
	// automatically mounted and scanned. I clearly won't be needing
	// to automount and import from ext4, xfs, etc.
	//
	// initial thoughts are only FAT-based disks should be imported
	// there may be a need for NTFS or AFS at some point, but i doubt-ish
	// it (maybe AFS for blackmagic gear)

	command := "findmnt -o source -n %s0"
	output, exitCode, err := callExternalCommand(command, mountPath)

	// testing err and exit code may be redundant, but whatever
	if err != nil || exitCode != 0 {
		slog.Warn("Could not find a device mounted to the following path: " + mountPath)
		return ""
	}

	devicePath := strings.TrimSpace(output)

	command = "lsblk -o FSTYPE,FSVER -n -r %s0"
	output, exitCode, err = callExternalCommand(command, devicePath)

	// testing err and exit code may be redundant, but whatever
	if err != nil || exitCode != 0 {
		return ""
	}

	output = strings.TrimSpace(output)
	parts := strings.Split(output, " ")
	fstype := parts[0]
	fsver := parts[1]

	if fstype == "vfat" && fsver == "FAT32" {
		return FAT32
	}

	if fstype == "exfat" {
		return ExFAT
	}

	slog.Warn("Unknown filesystem type: " + output)

	return ""
}

// PowerOffDevice attempts to eject or power off the device
// specified by the device parameter. This requires that a
// device path (ex: /dev/sda1) be provided. Returns true if
// it is successfully powered off (or was already powered off)
// and false on failure
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

// WatchForDeviceAttached is not currently built for linux, but will be in the future
func WatchForDeviceAttached(deviceMountedCallback func(devicePath string, volumePath string)) {
	// TODO: build this instead of requiring udev rules be installed to the system
	platformNotSupported(WatchForDeviceAttached)
}

//
// private functions
//

func pathMounted(device string) bool {
	findmntCommand := "findmnt %s0"
	_, exitCode, _ := callExternalCommand(findmntCommand, device)

	// device is mounted
	return exitCode == 0
}
