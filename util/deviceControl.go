//go:build !linux && !darwin

package util

func TestPlatform() {
	platformNotSupported(TestPlatform)
}

func GetVolumeName(mountPath string) string {
	platformNotSupported(GetVolumeName)
	return ""
}
