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

package nikonD3300

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"ccmm/model"
	"ccmm/util"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

const expectedVolumeName = "NIKON D3300"

var (
	fileMatchPatterns = [...]string{
		`DCIM/\d{3}D3300/DSC_\d{4}.(NEF|MOV)`,
	}
	logger *slog.Logger
)

type Processor struct {
	sourceDir    string
	sourceName   string
	volumeFormat string
}

func New(sourceDir string) *Processor {
	logger = slog.Default().With(slog.String("processor", "nikonD3300"))

	return &Processor{
		sourceDir:  sourceDir,
		sourceName: "",
	}
}

func (t *Processor) CheckSource() bool {
	logger.Debug(fmt.Sprintf("[CheckSource]: Beginning to test volume compatibility for '%s'", t.sourceDir))

	t.volumeFormat = util.GetVolumeFormat(t.sourceDir)

	// verify volume label matches what the camera sets
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing volume name at '%s'", t.sourceDir))
	if label := util.GetVolumeName(t.sourceDir); label != expectedVolumeName {
		logger.Debug(fmt.Sprintf("[CheckSource]: Volume label '%s' does not match required '%s' value, disqualified", label, expectedVolumeName))
		return false
	}

	// check for /DCIM directories
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing for required directories for volume '%s'", t.sourceDir))
	if !util.RequireDirs(t.sourceDir, []string{"DCIM"}) {
		logger.Debug("[CheckSource]: One or more required directories does not exist on source, disqualified")
		return false
	}

	// check for DCIM/\d+{3}D3300 directory
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing for existence of DCIM/xxxD3300 directory in volume '%s'", t.sourceDir))
	if exists, _ := util.RequireRegexDirMatch(path.Join(t.sourceDir, "DCIM"), `\d{3}D3300`); !exists {
		logger.Debug("[CheckSource]: No '/DCIM/xxxD3300/' directory found, disqualified")
		return false
	}

	logger.Debug(fmt.Sprintf("[CheckSource]: Volume '%s' is compatible", t.sourceDir))
	return true
}

func (t *Processor) EnumerateFiles() []model.SourceFile {
	return t.scanDirectory(path.Join(t.sourceDir, "DCIM"), "DCIM")
}

// private functions
func (t *Processor) getCameraModel(imagePath string) string {
	// TODO: Read warning below
	//!! This type of camera model caching could be an issue if we swapped
	//!! cards mid-event without first formatting. practically speaking, it
	//!! shouldn't be a problem since that's not something we've ever done
	if t.sourceName != "" {
		return t.sourceName
	}

	logger.Debug(fmt.Sprintf("Reading EXIF data from '%s'", imagePath))
	imageFile, err := os.Open(imagePath)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to open image file: %s", err.Error()))
		return ""
	}
	defer imageFile.Close()

	// Optionally register camera makenote data parsing - currently Nikon and
	// Canon are supported.
	exif.RegisterParsers(mknote.All...)

	x, err := exif.Decode(imageFile)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to decode exif data in image file: %s", err.Error()))
		return ""
	}

	camModel, _ := x.Get(exif.Model) // normally, don't ignore errors!
	camModelName, _ := camModel.StringVal()

	camModelName = strings.Replace(camModelName, "NIKON", "Nikon", -1)

	t.sourceName = camModelName
	return camModelName
}

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

				mediaType := "Photo"

				if strings.HasSuffix(entry.Name(), "MOV") {
					mediaType = "Video"
				}

				newFile := model.SourceFile{
					FileName:     entry.Name(),
					SourcePath:   fullPath,
					MediaType:    mediaType,
					Size:         stat.Size(),
					SourceName:   t.getCameraModel(fullPath),
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
