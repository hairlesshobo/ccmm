//go:build linux

package util

import "log/slog"

func TestPlatform() {
	slog.Info("This is linux")
}
