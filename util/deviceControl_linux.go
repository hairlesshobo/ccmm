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
	label, err := exec.Command("findmnt", "-n", "--output", "label", "--mountpoint", mountPath).Output()
	if err != nil {
		slog.Error(fmt.Sprintf("Error occurred while calling 'findmnt' command: %s", err.Error()))
	}

	return strings.TrimSuffix(string(label), "\n")
}
