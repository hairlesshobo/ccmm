//go:build linux

package util

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

func TestPlatform() {
	platformNotSupported(TestPlatform)
}

func GetVolumeName(mountPath string) string {
	command := "findmnt -n --output label --mountpoint %s0"
	cmdName, cmdArgs := prepareCommand(command, mountPath)
	slog.Debug("Calling external command", slog.String("command", cmdName), slog.Any("args", cmdArgs))
	label, err := exec.Command(cmdName, cmdArgs...).Output()
	if err != nil {
		slog.Error(fmt.Sprintf("Error occurred while calling 'findmnt' command: %s", err.Error()))
	}

	return strings.TrimSuffix(string(label), "\n")
}

func MountVolume(device string, mountPath string) bool {
	return false
}

func UnmountVolume(device string) bool {
	// this should be provided with the full path of the mounted partition,
	// ex: /dev/sda1

	//  udisksctl unmount --block-device %s0
	return false
}

func PowerOffDevice(device string) bool {
	// this needs to be provided the path to the block device itself,
	// for example /dev/sda (without the partition index). will need to
	// look this value up. one way is this:
	//  readlink -f "/sys/class/block/sda1/.."
	// note 2: power off might work on the partitoin device, double check it

	//  udisksctl power-off --block-device %s0
	return false
}
