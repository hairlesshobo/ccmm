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
package canonXA

import (
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"ccmm/model"
	"ccmm/util"
)

const expectedVolumeName = "CANON"
const mediaType = "Video"

type xmlResult struct {
	Value string `xml:"Device>ModelName"`
}

var (
	fileMatchPatterns = [...]string{
		`CONTENTS/CLIPS(\d+)/(\w)(\d+)(\w)(\d+)_(\d{6})(\w{2})_CANON.(MXF|XML)`,
	}
	logger *slog.Logger
)

type Processor struct {
	sourceDir string
}

func New(sourceDir string) *Processor {
	logger = slog.Default().With(slog.String("processor", "canonXA"))

	return &Processor{
		sourceDir: sourceDir,
	}
}

func (t *Processor) CheckSource() bool {
	logger.Debug(fmt.Sprintf("[CheckSource]: Beginning to test volume compatibility for '%s'", t.sourceDir))

	// verify volume label matches what the camera sets
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing volume name at '%s'", t.sourceDir))
	if label := util.GetVolumeName(t.sourceDir); label != expectedVolumeName {
		logger.Debug(fmt.Sprintf("[CheckSource]: Volume label '%s' does not match required '%s' value, disqualified", label, expectedVolumeName))
		return false
	}

	// check for /CONTENTS and /DCIM directories
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing for required directories for volume '%s'", t.sourceDir))
	if !util.RequireDirs(t.sourceDir, []string{"CONTENTS", "DCIM"}) {
		logger.Debug("[CheckSource]: One or more required directories does not exist on source, disqualified")
		return false
	}

	// check for CONTENTS/CLIPS(\d+)
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing for existence of CONTENTS/CLIPSxxx directory in volume '%s'", t.sourceDir))
	exists, clipsPath := util.RequireRegexDirMatch(path.Join(t.sourceDir, "CONTENTS"), `CLIPS(\d+)`)
	if !exists {
		logger.Debug("[CheckSource]: No '/CONTENTS/CLIPSXXX/' directory found, disqualified")
		return false
	}

	// check for CONTENTS/CLIPS(\d+)/INDEX.MIF file
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing for existence of CONTENTS/CLIPSxxx/INDEX.MIF in volume '%s'", t.sourceDir))
	if !util.RequireFiles(clipsPath, []string{"INDEX.MIF"}) {
		logger.Debug("[CheckSource]: INDEX.MIF file not found in CLIPS directory, disqualified")
		return false
	}

	logger.Debug(fmt.Sprintf("[CheckSource]: Volume '%s' is compatible", t.sourceDir))
	return true
}

func (t *Processor) EnumerateFiles() []model.SourceFile {
	return scanDirectory(path.Join(t.sourceDir, "CONTENTS"), "CONTENTS")
}

// private functions

func getSourceName(mediaPath string) string {
	sidecarFile := mediaPath
	if strings.HasSuffix(sidecarFile, "MXF") {
		sidecarFile = strings.TrimSuffix(sidecarFile, "MXF") + "XML"
	}

	x := xmlResult{"Unknown"}

	logger.Debug(fmt.Sprintf("[getSourceName]: Reading SourceName from file '%s'", sidecarFile))

	xmlFile, err := os.Open(sidecarFile)
	if err != nil {
		logger.Error(fmt.Sprintf("[getSourceName]: Failed to open sidecar file '%s': %s", sidecarFile, err.Error()))
	} else {
		defer xmlFile.Close()

		byteValue, _ := io.ReadAll(xmlFile)
		err := xml.Unmarshal(byteValue, &x)
		if err != nil {
			logger.Error(fmt.Sprintf("error: %v", err))
		}
	}

	return x.Value
}

func getCaptureDate(fileName string) time.Time {
	datePart := strings.Split(fileName, "_")[1][:6]
	zone, _ := time.Now().Zone()
	dtm, err := time.Parse("060102 MST", fmt.Sprintf("%s %s", datePart, zone))

	if err != nil {
		logger.Error(fmt.Sprintf("[getCaptureDate]: Failed to parse date '%s': %s", datePart, err.Error()))
	}

	return dtm
}

func scanDirectory(absoluteDirPath string, relativeDirPath string) []model.SourceFile {
	logger.Debug(fmt.Sprintf("[scanDirectory]: Scanning for source files at path '%s'", absoluteDirPath))

	var files []model.SourceFile

	// For this processor, we only care about .MXF files and the sidecar XML files
	// and we read the source name from the sidecar XML
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

				var newFile model.SourceFile
				newFile.FileName = entry.Name()
				newFile.SourcePath = fullPath
				newFile.MediaType = mediaType
				newFile.Size = stat.Size()
				newFile.SourceName = getSourceName(fullPath)
				newFile.CaptureDate = getCaptureDate(entry.Name())
				newFile.FileModTime = stat.ModTime()

				files = append(files, newFile)
			}
		}
	}

	return files
}
