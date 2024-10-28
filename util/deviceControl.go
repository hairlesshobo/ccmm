//go:build !linux && !darwin

package util

import "log/slog"

func TestPlatform() {
	slog.Info("This is unknown")
}
