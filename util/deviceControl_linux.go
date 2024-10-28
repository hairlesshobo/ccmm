//go:build linux

package util

func TestPlatform() {
	platformNotSupported(TestPlatform)
}

func GetVolumeName(mountPath string) string {
	platformNotSupported(GetVolumeName)
	return ""
}
