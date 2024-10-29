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
	return false
}

func PowerOffDevice(device string) bool {
	return false
}
