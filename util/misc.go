package util

import (
	"fmt"
	"log/slog"
	"os"
)

func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get hostname, error: %s", err.Error()))
		return ""
	}

	return hostname
}
