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

package jackRecorder

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"regexp"
	"time"

	"ccmm/model"
	"ccmm/util"
)

const mediaType = "Audio"

var (
	fileMatchPatterns = [...]string{
		`jack/\d{4}-\d{2}-\d{2}/([\w\d_-]+).wav`,
	}
	logger *slog.Logger
)

type Processor struct {
	sourceDir    string
	volumeFormat string
}

func New(sourceDir string) *Processor {
	logger = slog.Default().With(slog.String("processor", "jackRecorder"))

	return &Processor{
		sourceDir: sourceDir,
	}
}

func (t *Processor) CheckSource() bool {
	logger.Debug(fmt.Sprintf("[CheckSource]: Beginning to test volume compatibility for '%s'", t.sourceDir))

	t.volumeFormat = util.GetVolumeFormat(t.sourceDir)

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
	return t.scanDirectory(path.Join(t.sourceDir, "jack"), "jack")
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

func (t *Processor) scanDirectory(absoluteDirPath string, relativeDirPath string) []model.SourceFile {
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
			files = append(files, t.scanDirectory(fullPath, path.Join(relativeDirPath, entry.Name()))...)
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

				newFile := model.SourceFile{
					FileName:     entry.Name(),
					SourcePath:   fullPath,
					MediaType:    mediaType,
					Size:         stat.Size(),
					SourceName:   "Jack",
					CaptureDate:  getCaptureDate(dirName),
					FileModTime:  stat.ModTime(),
					VolumeFormat: t.volumeFormat,
				}

				files = append(files, newFile)
			}
		}
	}

	return files
}
