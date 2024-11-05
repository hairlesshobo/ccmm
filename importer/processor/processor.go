// =================================================================================
//
//		ccmm - https://www.foxhollow.cc/projects/ccmm/
//
//	 go-import-media, aka gim, is a tool for automatically importing media
//	 from removable disks into a predefined folder structure automatically.
//
//		Copyright (c) 2024 Steve Cross <flip@foxhollow.cc>
//
//		Licensed under the Apache License, Version 2.0 (the "License");
//		you may not use this file except in compliance with the License.
//		You may obtain a copy of the License at
//
//		     http://www.apache.org/licenses/LICENSE-2.0
//
//		Unless required by applicable law or agreed to in writing, software
//		distributed under the License is distributed on an "AS IS" BASIS,
//		WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//		See the License for the specific language governing permissions and
//		limitations under the License.
//
// =================================================================================

package processor

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"reflect"
	"slices"
	"strings"
	"time"

	"ccmm/importer/processor/behringerX32"
	"ccmm/importer/processor/blackmagicIOS"
	"ccmm/importer/processor/canonEOS"
	"ccmm/importer/processor/canonXA"
	"ccmm/importer/processor/jackRecorder"
	"ccmm/model"
	"ccmm/util"
)

type Processor interface {
	CheckSource() bool
	EnumerateFiles() []model.SourceFile
}

func useProcessor(config model.ImporterConfig, name string) bool {
	return len(config.EnabledProcessors) == 0 || slices.Contains(config.EnabledProcessors, name)
}

func initProcessors(config model.ImporterConfig, volumePath string) []Processor {
	processors := []Processor{}

	// I hate this. it hurts me. this should be in a map that is referenced and
	// then dynamically creates a new instance using reflection, but i'm still
	// new and couldn't figure it out quickly in go land, so i decided to go this
	// route for now. gross, but works.

	if useProcessor(config, "behringerX32") {
		processors = append(processors, behringerX32.New(volumePath))
	}

	if useProcessor(config, "blackmagicIOS") {
		processors = append(processors, blackmagicIOS.New(volumePath))
	}

	if useProcessor(config, "canonEOS") {
		processors = append(processors, canonEOS.New(volumePath))
	}

	if useProcessor(config, "canonXA") {
		processors = append(processors, canonXA.New(volumePath))
	}

	if useProcessor(config, "jackRecorder") {
		processors = append(processors, jackRecorder.New(volumePath))
	}

	return processors
}

func FindProcessors(config model.ImporterConfig, volumePath string) []Processor {
	slog.Info(fmt.Sprintf("processor.FindProcessors: Looking for processors to handle path '%s'", volumePath))
	processors := initProcessors(config, volumePath)
	var foundProcessors []Processor

	for _, processor := range processors {
		if processor.CheckSource() {
			foundProcessors = append(foundProcessors, processor)
		}
	}

	if len(foundProcessors) == 0 {
		// TODO: eject and flash yellow if no processor found
		slog.Warn(fmt.Sprintf("processor.FindProcessors: No processor found for volume path '%s', skipping", volumePath))
		return nil
	}

	for _, processor := range foundProcessors {
		processorName := strings.Split(reflect.TypeOf(processor).String(), ".")[0][1:]
		slog.Info(fmt.Sprintf("processor.FindProcessors: Found processor '%s' to handle path '%s'", processorName, volumePath))
	}

	return foundProcessors
}

func EnumerateSources(processors []Processor, dump bool) []model.SourceFile {
	var allFiles []model.SourceFile

	for _, processor := range processors {
		processorFiles := processor.EnumerateFiles()

		allFiles = append(allFiles, processorFiles...)
	}

	if dump {
		j, _ := json.MarshalIndent(allFiles, "", "  ")
		fmt.Println(string(j))
	}

	return allFiles
}

func ImportFiles(config model.ImporterConfig, files []model.SourceFile, dryRun bool) {
	for _, sourceFile := range files {
		destDir := util.GetDestinationDirectory(config.LiveDataDir, sourceFile)
		destPath := path.Join(destDir, sourceFile.FileName)

		// Create the dir and parents, if needed
		if !dryRun {
			os.MkdirAll(destDir, 0755)
		}

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

		if dryRun {
			slog.Info(fmt.Sprintf("[Dry run] Would copy '%s' to '%s'", sourceFile.SourcePath, destPath))
		} else {
			slog.Info(fmt.Sprintf("Copying '%s' to '%s'", sourceFile.SourcePath, destPath))
			// TODO: find a way to add transfer speeds to this
			util.CopyFile(sourceFile.SourcePath, destPath)
		}

		os.Chtimes(destPath, time.Time{}, sourceFile.FileModTime)
	}
}
