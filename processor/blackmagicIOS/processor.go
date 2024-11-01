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
package blackmagicIOS

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gim/model"
	"gim/util"
)

var (
	fileMatchPattern = `^(\w)(\d{3})_(\d{8})_C(\d{3}).mov`
	logger           *slog.Logger
)

type Processor struct {
	sourceDir string
}

func New(sourceDir string) *Processor {
	logger = slog.Default().With(slog.String("processor", "blackmagicIOS"))

	return &Processor{
		sourceDir: sourceDir,
	}
}

func (t *Processor) CheckSource() bool {
	logger.Debug(fmt.Sprintf("[CheckSource]: Beginning to test volume compatibility for '%s'", t.sourceDir))

	// check for DCIM/EOSMISC/Mxxxx.CTG file
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing for existence of '%s' file in volume '%s'", fileMatchPattern, t.sourceDir))
	exists, foundFile := util.RequireRegexFileMatch(t.sourceDir, fileMatchPattern)
	if !exists {
		logger.Debug(fmt.Sprintf("[CheckSource]: No '%s' file found, disqualified", fileMatchPattern))
		return false
	}

	modelName := util.MediaInfo_GetGeneralParameter(foundFile, "com.apple.quicktime.software")

	if !strings.HasPrefix(modelName, "Blackmagic Cam") {
		logger.Debug(fmt.Sprintf("[CheckSource]: Camera model '%s' does not begin with the required 'Blackmagic Cam', disqualified", modelName))
		return false
	}

	logger.Debug(fmt.Sprintf("[CheckSource]: Volume '%s' is compatible", t.sourceDir))
	return true
}

func (t *Processor) EnumerateFiles() []model.SourceFile {
	return t.scanDirectory(t.sourceDir, "")
}

//
// private functions
//

func (t *Processor) getCaptureDate(filePath string) time.Time {
	format := "2006-01-02 15:04:05 MST"
	dateOnlyFormat := "2006-01-02 MST"
	result := util.MediaInfo_GetGeneralParameter(filePath, "Encoded_Date")

	dtm, err := time.Parse(format, result)

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to parse dtm, error: %s", err.Error()))
	}

	date, err := time.Parse(dateOnlyFormat, dtm.Local().Format(dateOnlyFormat))

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to parse date, error: %s", err.Error()))
	}

	return date
}

func (t *Processor) scanDirectory(absoluteDirPath string, relativeDirPath string) []model.SourceFile {
	logger.Debug(fmt.Sprintf("[scanDirectory]: Scanning for source files at path '%s'", absoluteDirPath))

	var files []model.SourceFile

	entries, err := os.ReadDir(absoluteDirPath)

	if err != nil {
		logger.Error(fmt.Sprintf("[scanDirectory]: Error occured while scanning directory '%s': %s", absoluteDirPath, err.Error()))
		return nil
	}

	for _, entry := range entries {
		fullPath := path.Join(absoluteDirPath, entry.Name())
		relativePath := path.Join(relativeDirPath, entry.Name())

		if matched, _ := regexp.MatchString(fileMatchPattern, relativePath); matched {
			logger.Debug(fmt.Sprintf("[scanDirectory]: Matched file '%s'", fullPath))

			stat, _ := os.Stat(fullPath)
			label := filepath.Base(absoluteDirPath)

			var newFile model.SourceFile
			newFile.FileName = entry.Name()
			newFile.SourcePath = fullPath
			newFile.MediaType = "Video"
			newFile.Size = stat.Size()
			newFile.SourceName = label
			newFile.CaptureDate = t.getCaptureDate(fullPath)
			newFile.FileModTime = stat.ModTime()

			files = append(files, newFile)
		}
	}

	return files
}
