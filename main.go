package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/hairlesshobo/go-import-media/processor"
	"github.com/hairlesshobo/go-import-media/util"
)

func main() {
	util.LoadConfig()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	processors := FindProcessors("/Volumes/CANON")

	for _, processor := range processors {
		files := processor.EnumerateFiles()
		j, _ := json.MarshalIndent(files, "", "  ")
		fmt.Println(string(j))
	}
}

func FindProcessors(volumePath string) []processor.Processor {
	processors := []processor.Processor{&processor.CanonXA{}}
	var foundProcessors []processor.Processor

	for _, processor := range processors {
		processor.SetSourceDir(volumePath)

		if processor.CheckSource() {
			slog.Debug("matched")
			foundProcessors = append(foundProcessors, processor)
		}
	}

	if len(foundProcessors) == 0 {
		slog.Info(fmt.Sprintf("No processor found for volume path '%s', skipping", volumePath))
		return nil
	}

	return foundProcessors
}
