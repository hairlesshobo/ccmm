package action

import (
	"fmt"
	"log/slog"

	"github.com/hairlesshobo/go-import-media/model"
	"github.com/hairlesshobo/go-import-media/processor"
)

func Import(params model.ImportVolume) {
	slog.Info(fmt.Sprintf("Beginning import for volume '%s'", params.VolumePath))

	processors := processor.FindProcessors(params.VolumePath)
	processor.ProcessSources(processors, params.DryRun)

	// do empty, if enabled

	// do eject, if enabled

	slog.Info(fmt.Sprintf("Finished import for volume '%s'", params.VolumePath))
}
