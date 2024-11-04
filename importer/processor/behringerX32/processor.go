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

package behringerX32

import (
	// "encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"ccmm/model"
	"ccmm/util"
)

const expectedVolumeName = "X32"
const mediaType = "Audio"

var (
	fileMatchPatterns = [...]string{
		`^R_(\d{8})-(\d{6}).wav`,
	}
	logger *slog.Logger
)

type Processor struct {
	sourceDir string
}

func New(sourceDir string) *Processor {
	logger = slog.Default().With(slog.String("processor", "behringerX32"))

	return &Processor{
		sourceDir: sourceDir,
	}
}

func (t *Processor) CheckSource() bool {
	logger.Debug(fmt.Sprintf("[CheckSource]: Beginning to test volume compatibility for '%s'", t.sourceDir))

	// verify volume label matches what is expected
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing volume name at '%s'", t.sourceDir))
	if label := util.GetVolumeName(t.sourceDir); !strings.HasPrefix(label, expectedVolumeName) {
		logger.Debug(fmt.Sprintf("[CheckSource]: Volume label '%s' does not start with required '%s' value, disqualified", label, expectedVolumeName))
		return false
	}

	// check for recorded audio files
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing for existence of '%s' file in volume '%s'", fileMatchPatterns[0], t.sourceDir))
	exists, _ := util.RequireRegexFileMatch(t.sourceDir, fileMatchPatterns[0])
	if !exists {
		logger.Debug(fmt.Sprintf("[CheckSource]: No '%s' file found, disqualified", fileMatchPatterns[0]))
		return false
	}

	logger.Debug(fmt.Sprintf("[CheckSource]: Volume '%s' is compatible", t.sourceDir))
	return true
}

func (t *Processor) EnumerateFiles() []model.SourceFile {
	return scanDirectory(t.sourceDir, "")
}

// private functions

func getCaptureDate(fileName string) time.Time {
	datePart := fileName[2:10]

	zone, _ := time.Now().Zone()
	dtmStr := fmt.Sprintf("%s %s", datePart, zone)
	dtm, err := time.Parse("20060102 MST", dtmStr)

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

		if !entry.IsDir() {
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

				if stat.Size() == 0 {
					logger.Info(fmt.Sprintf("[scanDirectory]: Skipping 0 byte file '%s'", fullPath))
					continue
				}

				var newFile model.SourceFile
				newFile.FileName = entry.Name()
				newFile.SourcePath = fullPath
				newFile.MediaType = mediaType
				newFile.Size = stat.Size()
				newFile.SourceName = "X32"
				newFile.CaptureDate = getCaptureDate(entry.Name())
				newFile.FileModTime = stat.ModTime()

				files = append(files, newFile)
			}
		}
	}

	return files
}
