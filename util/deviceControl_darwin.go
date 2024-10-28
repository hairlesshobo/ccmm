//go:build darwin

package util

import "log/slog"

func TestPlatform() {
	slog.Info("This is darwin")
}
