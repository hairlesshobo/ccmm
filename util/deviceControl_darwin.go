//go:build darwin

package util

import (
	"path"
)

func TestPlatform() {
	platformNotSupported(TestPlatform)
}

func GetVolumeName(mountPath string) string {
	_, label := path.Split(mountPath)
	return label
}
