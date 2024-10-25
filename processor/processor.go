package processor

import (
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/hairlesshobo/go-import-media/model"
	"github.com/hairlesshobo/go-import-media/util"
)

type Processor interface {
	SetSourceDir(sourceDir string)
	CheckSource() bool
	EnumerateFiles() []model.SourceFile
}

func FindProcessors(volumePath string) []Processor {
	processors := []Processor{&CanonXA{}}
	var foundProcessors []Processor

	for _, processor := range processors {
		processor.SetSourceDir(volumePath)

		if processor.CheckSource() {
			foundProcessors = append(foundProcessors, processor)
		}
	}

	if len(foundProcessors) == 0 {
		// TODO: eject and flash yellow if not processor found
		slog.Info(fmt.Sprintf("No processor found for volume path '%s', skipping", volumePath))
		return nil
	}

	return foundProcessors
}

func ProcessSources(processors []Processor) bool {
	success := true
	for _, processor := range processors {
		success = success && ProcessSource(processor)
	}

	// TODO: is this necessary?
	return success
}

func ProcessSource(processor Processor) bool {
	files := processor.EnumerateFiles()
	// j, _ := json.MarshalIndent(files, "", "  ")
	// fmt.Println(string(j))

	for _, sourceFile := range files {
		destDir := util.GetDestinationDirectory(model.Config.LiveDataDir, sourceFile)
		destPath := path.Join(destDir, sourceFile.FileName)

		fmt.Printf("%s    -->   %s\n", sourceFile.SourcePath, destPath)

		// Create the dir and parents, if needed
		os.MkdirAll(destDir, 0755)
	}

	// TODO: umm.. do i need a return? maybe some error handling...
	return true
}
