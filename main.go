package main

import (
	"log/slog"
	"os"

	"github.com/hairlesshobo/go-import-media/model"
	"github.com/hairlesshobo/go-import-media/processor"
	"github.com/hairlesshobo/go-import-media/util"
)

func main() {
	util.LoadConfig()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.Level(model.Config.LogLevel)}))
	slog.SetDefault(logger)

	processors := processor.FindProcessors("/Volumes/EOS_DIGITAL")
	processor.ProcessSources(processors)
}
