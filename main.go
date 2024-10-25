package main

import (
	"log/slog"
	"os"

	"github.com/hairlesshobo/go-import-media/processor"
	"github.com/hairlesshobo/go-import-media/util"
)

func main() {
	util.LoadConfig()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	processors := processor.FindProcessors("/Volumes/CANON")
	processor.ProcessSources(processors)
}
