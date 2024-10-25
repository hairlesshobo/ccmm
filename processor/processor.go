package processor

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/hairlesshobo/go-import-media/model"
	"github.com/hairlesshobo/go-import-media/util"
)

type Processor interface {
	SetSourceDir(sourceDir string)
	CheckSource() bool
	EnumerateFiles() []model.SourceFile
}

func initProcessors(volumePath string) []Processor {
	processors := []Processor{&CanonXA{}}

	for _, processor := range processors {
		processor.SetSourceDir(volumePath)
	}

	return processors
}

func FindProcessors(volumePath string) []Processor {
	slog.Info(fmt.Sprintf("processor.FindProcessors: Looking for processors to handle path '%s'", volumePath))
	processors := initProcessors(volumePath)
	var foundProcessors []Processor

	for _, processor := range processors {
		if processor.CheckSource() {
			processorName := strings.Split(reflect.TypeOf(processor).String(), ".")[1]
			slog.Info(fmt.Sprintf("processor.FindProcessors: Found processor '%s' to handle path '%s'", processorName, volumePath))
			foundProcessors = append(foundProcessors, processor)
		}
	}

	if len(foundProcessors) == 0 {
		// TODO: eject and flash yellow if not processor found
		slog.Warn(fmt.Sprintf("processor.FindProcessors: No processor found for volume path '%s', skipping", volumePath))
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

		// Create the dir and parents, if needed
		os.MkdirAll(destDir, 0755)

		stat, err := os.Stat(destPath)
		fileExists := err == nil && stat.Mode().IsRegular()
		sameSize := false
		if stat != nil {
			sameSize = stat.Size() == sourceFile.Size
		}

		if fileExists && sameSize {
			slog.Debug(fmt.Sprintf("Not copying file because the destination already exists and is same size at '%s'", destPath))
			continue
		}

		if fileExists && !sameSize {
			slog.Debug(fmt.Sprintf("File already exists but is different size, will copy to '%s'", destPath))
		}

		slog.Info(fmt.Sprintf("Copying '%s' to '%s'", sourceFile.SourcePath, destPath))
		util.CopyFile(sourceFile.SourcePath, destPath)

		os.Chtimes(destPath, time.Time{}, sourceFile.FileModTime)
	}

	// TODO: umm.. do i need a return? maybe some error handling...
	return true
}
