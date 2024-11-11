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

package zoomH1n

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

const expectedVolumeName = "H1N_SD"

var (
	fileMatchPatterns = [...]string{
		`STEREO/FOLDER\d{2}/ZOOM\d{4}.WAV`,
	}
	logger *slog.Logger
)

type Processor struct {
	sourceDir    string
	sourceName   string
	volumeFormat string
	fileRegexes  []regexp.Regexp
}

func New(sourceDir string) *Processor {
	logger = slog.Default().With(slog.String("processor", "zoomH1n"))

	processor := &Processor{
		sourceDir:    sourceDir,
		sourceName:   "",
		volumeFormat: "",
		fileRegexes:  make([]regexp.Regexp, 0),
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

	// verify volume label matches what the recorder sets
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing volume name at '%s'", t.sourceDir))
	if label := util.GetVolumeName(t.sourceDir); label != expectedVolumeName {
		logger.Debug(fmt.Sprintf("[CheckSource]: Volume label '%s' does not match required '%s' value, disqualified", label, expectedVolumeName))
		return false
	}

	// check for /STEREO directory
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing for required directories for volume '%s'", t.sourceDir))
	if !util.RequireDirs(t.sourceDir, []string{"STEREO"}) {
		logger.Debug("[CheckSource]: One or more required directories does not exist on source, disqualified")
		return false
	}

	// check for /STEREO/FOLDERxx directories
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing for required directories for volume '%s'", t.sourceDir))

	if exists, _ := util.RequireRegexDirMatch(path.Join(t.sourceDir, "STEREO"), `FOLDER\d{2}`); !exists {
		logger.Debug("[CheckSource]: One or more required directories does not exist on source, disqualified")
		return false
	}

	logger.Debug(fmt.Sprintf("[CheckSource]: Volume '%s' is compatible", t.sourceDir))
	return true
}

func (t *Processor) EnumerateFiles() []model.SourceFile {
	// TODO: does this thing use a dir other than STEREO, perhaps if recording in dual mono?
	return t.scanDirectory(path.Join(t.sourceDir, "STEREO"), "STEREO")
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

	// For this processor, we only care about .MXF files and the sidecar XML files
	// and we read the source name from the sidecar XML
	// TODO: Create a shared ReadDir that includes global filtering but mimmics the API of os.ReadDir
	entries, err := os.ReadDir(absoluteDirPath)

	if err != nil {
		logger.Error(fmt.Sprintf("[scanDirectory]: Error occured while scanning directory '%s': %s", absoluteDirPath, err.Error()))
		return nil
	}

	folderRegex, _ := regexp.Compile(`FOLDER\d{2}`)

	for _, entry := range entries {
		fullPath := path.Join(absoluteDirPath, entry.Name())
		relativePath := path.Join(relativeDirPath, entry.Name())

		if entry.IsDir() {
			// only ascend into known top-level directories named FOLDERxx
			if relativeDirPath == "" {
				if !folderRegex.MatchString(entry.Name()) {
					slog.Debug("Skipping unknown top-level directory: " + entry.Name())
					continue
				}
			}

			files = append(files, t.scanDirectory(fullPath, path.Join(relativeDirPath, entry.Name()))...)
		} else {
			foundMatch := false

			for _, regexC := range t.fileRegexes {
				if regexC.MatchString(relativePath) {
					foundMatch = true
					break
				}
			}

			if foundMatch {
				logger.Debug(fmt.Sprintf("[scanDirectory]: Matched file '%s'", fullPath))

				stat, _ := os.Stat(fullPath)

				newFile := model.SourceFile{
					FileName:     entry.Name(),
					SourcePath:   fullPath,
					MediaType:    "Audio",
					Size:         stat.Size(),
					SourceName:   "Zoom H6",
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
