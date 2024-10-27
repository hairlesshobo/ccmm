package action

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hairlesshobo/go-import-media/model"
	"github.com/hairlesshobo/go-import-media/processor"
)

func Import(params model.ImportVolume) {
	slog.Info(fmt.Sprintf("Beginning import for volume '%s'", params.VolumePath))

	if stat, err := os.Stat(params.VolumePath); err != nil || !stat.IsDir() {
		slog.Error(fmt.Sprintf("Directory not found: %s", params.VolumePath))
		os.Exit(1)
	}

	processors := processor.FindProcessors(params.VolumePath)
	processor.ProcessSources(processors, params.DryRun)

	// do empty, if enabled

	// do eject, if enabled

	slog.Info(fmt.Sprintf("Finished import for volume '%s'", params.VolumePath))
}
