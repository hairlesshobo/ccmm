package action

import (
	"fmt"
	"log/slog"

	"github.com/hairlesshobo/go-import-media/model"
	"github.com/hairlesshobo/go-import-media/processor"
	"github.com/hairlesshobo/go-import-media/util"
)

func Import(params model.ImportVolume) bool {
	slog.Info(fmt.Sprintf("Beginning import for volume '%s'", params.VolumePath))

	if !util.DirectoryExists(params.VolumePath) {
		slog.Error(fmt.Sprintf("Directory not found: %s", params.VolumePath))
		return false
	}

	processors := processor.FindProcessors(params.VolumePath)
	processor.ProcessSources(processors, params.DryRun)

	slog.Info(fmt.Sprintf("Finished import for volume '%s'", params.VolumePath))

	return true
}
