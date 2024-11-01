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
package jackRecorder

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"regexp"
	"time"

	"gim/model"
	"gim/util"
)

const mediaType = "Audio"

var (
	fileMatchPatterns = [...]string{`(.+).wav`}
	logger            *slog.Logger
)

type Processor struct {
	sourceDir string
}

func New(sourceDir string) *Processor {
	logger = slog.Default().With(slog.String("processor", "jackRecorder"))

	return &Processor{
		sourceDir: sourceDir,
	}
}

func (t *Processor) CheckSource() bool {
	logger.Debug(fmt.Sprintf("[CheckSource]: Beginning to test volume compatibility for '%s'", t.sourceDir))

	// check for jack directory
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing for required directories for volume '%s'", t.sourceDir))
	if !util.RequireDirs(t.sourceDir, []string{"jack"}) {
		logger.Debug("[CheckSource]: One or more required directories does not exist on source, disqualified")
		return false
	}

	// check for jack/(\d{4})-(\d{2})-(\d{2})
	logger.Debug(fmt.Sprintf(`[CheckSource]: Testing for existence of jack/(\d{4})-(\d{2})-(\d{2}) directory in volume '%s'`, t.sourceDir))
	exists, _ := util.RequireRegexDirMatch(path.Join(t.sourceDir, "jack"), `(\d{4})-(\d{2})-(\d{2})`)
	if !exists {
		logger.Debug(`[CheckSource]: No '/jack/(\d{4})-(\d{2})-(\d{2})/' directory found, disqualified`)
		return false
	}

	logger.Debug(fmt.Sprintf("[CheckSource]: Volume '%s' is compatible", t.sourceDir))
	return true
}

func (t *Processor) EnumerateFiles() []model.SourceFile {
	return scanDirectory(path.Join(t.sourceDir, "jack"), "jack")
}

// private functions

func getCaptureDate(directoryName string) time.Time {
	zone, _ := time.Now().Zone()
	dtmStr := fmt.Sprintf("%s %s", directoryName, zone)
	dtm, err := time.Parse("2006-01-02 MST", dtmStr)

	if err != nil {
		logger.Error(fmt.Sprintf("[getCaptureDate]: Failed to parse date '%s': %s", dtmStr, err.Error()))
	}

	return dtm
}

func scanDirectory(absoluteDirPath string, relativeDirPath string) []model.SourceFile {
	logger.Debug(fmt.Sprintf("[scanDirectory]: Scanning for source files at path '%s'", absoluteDirPath))

	var files []model.SourceFile

	// For this processor, we only care about .wav files
	entries, err := os.ReadDir(absoluteDirPath)

	if err != nil {
		logger.Error(fmt.Sprintf("[scanDirectory]: Error occured while scanning directory '%s': %s", absoluteDirPath, err.Error()))
		return nil
	}

	for _, entry := range entries {
		fullPath := path.Join(absoluteDirPath, entry.Name())
		relativePath := path.Join(relativeDirPath, entry.Name())

		if entry.IsDir() {
			files = append(files, scanDirectory(fullPath, path.Join(relativeDirPath, entry.Name()))...)
		} else {
			foundMatch := false

			for _, pattern := range fileMatchPatterns {
				if matched, _ := regexp.MatchString(pattern, relativePath); matched {
					foundMatch = true
					break
				}
			}

			if foundMatch {
				logger.Debug(fmt.Sprintf("[scanDirectory]: Matched file '%s'", fullPath))

				stat, _ := os.Stat(fullPath)
				_, dirName := path.Split(absoluteDirPath)

				var newFile model.SourceFile
				newFile.FileName = entry.Name()
				newFile.SourcePath = fullPath
				newFile.MediaType = mediaType
				newFile.Size = stat.Size()
				newFile.SourceName = "Jack"
				newFile.CaptureDate = getCaptureDate(dirName)
				newFile.FileModTime = stat.ModTime()

				files = append(files, newFile)
			}
		}
	}

	return files
}
