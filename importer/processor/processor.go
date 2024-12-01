// =================================================================================
//
//		ccmm - https://www.foxhollow.cc/projects/ccmm/
//
//	 Connection Church Media Manager, aka ccmm, is a tool for managing all
//   aspects of produced media- initial import from removable media,
//   synchronization with clients and automatic data replication and backup
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
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"time"

	"ccmm/importer/processor/behringerX32"
	"ccmm/importer/processor/behringerXLIVE"
	"ccmm/importer/processor/blackmagicIOS"
	"ccmm/importer/processor/canonEOS"
	"ccmm/importer/processor/canonXA"
	"ccmm/importer/processor/jackRecorder"
	"ccmm/importer/processor/nikonD3300"
	"ccmm/importer/processor/zoomH1n"
	"ccmm/importer/processor/zoomH6"
	"ccmm/model"
	"ccmm/util"
)

type Processor interface {
	CheckSource() bool
	EnumerateFiles() []model.SourceFile
}

func useProcessor(enabledProcessors []string, name string) bool {
	return len(enabledProcessors) == 0 || slices.Contains(enabledProcessors, name)
}

func InitProcessors(enabledProcessors []string, volumePath string) []Processor {
	processors := []Processor{}

	// I hate this. it hurts me. this should be in a map that is referenced and
	// then dynamically creates a new instance using reflection, but i'm still
	// new and couldn't figure it out quickly in go land, so i decided to go this
	// route for now. gross, but works.

	if useProcessor(enabledProcessors, "behringerX32") {
		processors = append(processors, behringerX32.New(volumePath))
	}

	if useProcessor(enabledProcessors, "behringerXLIVE") {
		processors = append(processors, behringerXLIVE.New(volumePath))
	}

	if useProcessor(enabledProcessors, "blackmagicIOS") {
		processors = append(processors, blackmagicIOS.New(volumePath))
	}

	if useProcessor(enabledProcessors, "canonEOS") {
		processors = append(processors, canonEOS.New(volumePath))
	}

	if useProcessor(enabledProcessors, "canonXA") {
		processors = append(processors, canonXA.New(volumePath))
	}

	if useProcessor(enabledProcessors, "jackRecorder") {
		processors = append(processors, jackRecorder.New(volumePath))
	}

	if useProcessor(enabledProcessors, "nikonD3300") {
		processors = append(processors, nikonD3300.New(volumePath))
	}

	if useProcessor(enabledProcessors, "zoomH1n") {
		processors = append(processors, zoomH1n.New(volumePath))
	}

	if useProcessor(enabledProcessors, "zoomH6") {
		processors = append(processors, zoomH6.New(volumePath))
	}

	return processors
}

func FindProcessors(config model.ImporterConfig, volumePath string) []Processor {
	slog.Info(fmt.Sprintf("processor.FindProcessors: Looking for processors to handle path '%s'", volumePath))
	processors := InitProcessors(config.EnabledProcessors, volumePath)
	var foundProcessors []Processor

	for _, processor := range processors {
		if processor.CheckSource() {
			foundProcessors = append(foundProcessors, processor)
		}
	}

	if len(foundProcessors) == 0 {
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

	for idx := range allFiles {
		file := &allFiles[idx]

		// TODO: add global filtering here for the following:
		//   - empty directories
		//   - 0 byte files
		//   - ignore mac dot files
		//   - files older than a certain period
		//   - other stuff?

		// this is to fix how linux handles file mod time on xfat/exfat devices
		if runtime.GOOS == "linux" && time.Local != time.UTC && (file.VolumeFormat == util.FAT32 || file.VolumeFormat == util.ExFAT) {
			patchedDtm, _ := time.ParseInLocation(time.DateTime, file.FileModTime.UTC().Format(time.DateTime), time.Local)
			file.FileModTime = patchedDtm
		}
	}

	if dump {
		j, _ := json.MarshalIndent(allFiles, "", "  ")
		fmt.Println(string(j))
	}

	return allFiles
}

func ImportFiles(config model.ImporterConfig, files []model.SourceFile, dryRun bool) {
	for _, sourceFile := range files {
		destPath := path.Join(util.GetDestinationDirectory(config.LiveDataDir, sourceFile), sourceFile.FileName)

		// Create the dir and parents, if needed
		if !dryRun {
			destDir := filepath.Dir(destPath)
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
