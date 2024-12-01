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

package behringerXLIVE

import (
	// "encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"ccmm/model"
	"ccmm/util"
)

const expectedVolumeName = "XLIVE"
const mediaType = "Audio"

var (
	fileMatchPatterns = [...]string{
		`X_LIVE/[A-Z|0-9]{8}/[A-Z0-9]{8}.WAV`,
		`X_LIVE/[A-Z|0-9]{8}/SE_LOG.BIN`,
	}
	logger *slog.Logger
)

type Processor struct {
	sourceDir    string
	volumeFormat string
	fileRegexes  []regexp.Regexp
}

func New(sourceDir string) *Processor {
	logger = slog.Default().With(slog.String("processor", "behringerXLIVE"))

	processor := &Processor{
		sourceDir: sourceDir,
	}

	for _, pattern := range fileMatchPatterns {
		regexC, err := regexp.Compile(pattern)

		if err != nil {
			slog.Error("Invalid regex pattern provided: '" + pattern + "'")
			return nil
		}

		processor.fileRegexes = append(processor.fileRegexes, *regexC)
	}

	return processor
}

func (t *Processor) CheckSource() bool {
	logger.Debug(fmt.Sprintf("[CheckSource]: Beginning to test volume compatibility for '%s'", t.sourceDir))

	t.volumeFormat = util.GetVolumeFormat(t.sourceDir)

	// verify volume label matches what is expected
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing volume name at '%s'", t.sourceDir))
	if label := util.GetVolumeName(t.sourceDir); !strings.HasPrefix(label, expectedVolumeName) {
		logger.Debug(fmt.Sprintf("[CheckSource]: Volume label '%s' does not start with required '%s' value, disqualified", label, expectedVolumeName))
		return false
	}

	// check for /X_LIVE directory
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing for required directories for volume '%s'", t.sourceDir))
	if !util.RequireDirs(t.sourceDir, []string{"X_LIVE"}) {
		logger.Debug("[CheckSource]: One or more required directories does not exist on source, disqualified")
		return false
	}

	// check for recorded audio files
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing for existence of '%s' file in volume '%s'", fileMatchPatterns[0], t.sourceDir))
	exists, subDir := util.RequireRegexDirMatch(path.Join(t.sourceDir, "X_LIVE"), `[A-Z|0-9]{8}`)
	if !exists {
		logger.Debug("[CheckSource]: No directory found matching regex '[A-Z|0-9]{8}', disqualified")
		return false
	}

	// check for X_LIVE/[A-Z|0-9]{8}/SE_LOG.BIN file
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing for existence of X_LIVE/XXXXXXXX/SE_LOG.BIN file in volume '%s'", t.sourceDir))
	if !util.RequireFiles(subDir, []string{"SE_LOG.BIN"}) {
		logger.Debug("[CheckSource]: No '/X_LIVE/XXXXXXXX/SE_LOG.BIN' file found, disqualified")
		return false
	}

	logger.Debug(fmt.Sprintf("[CheckSource]: Volume '%s' is compatible", t.sourceDir))
	return true
}

func (t *Processor) EnumerateFiles() []model.SourceFile {
	return t.scanDirectory(path.Join(t.sourceDir, "X_LIVE"), "X_LIVE")
}

// private functions

func (t *Processor) getCaptureDate(dtm time.Time) time.Time {
	format := "2006-01-02 MST"
	date, err := time.Parse(format, dtm.Format(format))

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to parse date, error: %s", err.Error()))
	}

	return date
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

			for _, regexC := range t.fileRegexes {
				fmt.Println(relativePath)
				if regexC.MatchString(relativePath) {
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

				parentDirName := filepath.Base(absoluteDirPath)

				newFile := model.SourceFile{
					FileName:     fmt.Sprintf("%s/%s", parentDirName, entry.Name()),
					SourcePath:   fullPath,
					MediaType:    mediaType,
					Size:         stat.Size(),
					SourceName:   "X-Live",
					CaptureDate:  t.getCaptureDate(stat.ModTime()),
					FileModTime:  stat.ModTime(),
					VolumeFormat: t.volumeFormat,
				}

				files = append(files, newFile)
			}
		}
	}

	return files
}
