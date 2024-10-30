// =================================================================================
//
//		gim - https://www.foxhollow.cc/projects/gim/
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
	// "encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"gim/model"
	"gim/processor/behringerX32"
	"gim/processor/canonEOS"
	"gim/processor/canonXA"
	"gim/processor/jackRecorder"
	"gim/util"
)

type Processor interface {
	SetSourceDir(sourceDir string)
	CheckSource() bool
	EnumerateFiles() []model.SourceFile
}

func initProcessors(volumePath string) []Processor {
	processors := []Processor{
		&canonXA.Processor{},
		&canonEOS.Processor{},
		&behringerX32.Processor{},
		&jackRecorder.Processor{},
	}

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

func ProcessSources(processors []Processor, dryRun bool) bool {
	success := true
	for _, processor := range processors {
		success = success && ProcessSource(processor, dryRun)
	}

	// TODO: is this necessary?
	return success
}

func ProcessSource(processor Processor, dryRun bool) bool {
	files := processor.EnumerateFiles()

	//j, _ := json.MarshalIndent(files, "", "  ")
	//fmt.Println(string(j))

	//return true

	for _, sourceFile := range files {
		destDir := util.GetDestinationDirectory(model.Config.LiveDataDir, sourceFile)
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
			util.CopyFile(sourceFile.SourcePath, destPath)
		}

		os.Chtimes(destPath, time.Time{}, sourceFile.FileModTime)
	}

	// TODO: umm.. do i need a return? maybe some error handling...
	return true
}
