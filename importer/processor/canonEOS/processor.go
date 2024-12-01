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

package canonEOS

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

	"github.com/barasher/go-exiftool"
)

const expectedVolumeName = "EOS_DIGITAL"

var (
	fileMatchPatterns = [...]string{
		`DCIM/(\d+)CANON/IMG_(\d+).CR2`,
		`DCIM/(\d+)CANON/MVI_(\d+).MOV`,
		`DCIM/(\d+)(CANON|EOS)([\w\d]{0,})/([\w\d_]{4}(\d{4})).(MOV|CR2|CR3|MP4)`,
	}
	logger *slog.Logger
)

type Processor struct {
	sourceDir    string
	volumeFormat string
	etHandle     *exiftool.Exiftool
}

func New(sourceDir string) *Processor {
	logger = slog.Default().With(slog.String("processor", "canonEOS"))

	return &Processor{
		sourceDir: sourceDir,
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

	// check for /DCIM and /MISC directories
	logger.Debug(fmt.Sprintf("[CheckSource]: Testing for required directories for volume '%s'", t.sourceDir))
	if !util.RequireDirs(t.sourceDir, []string{"DCIM", "MISC"}) {
		logger.Debug("[CheckSource]: One or more required directories does not exist on source, disqualified")
		return false
	}

	foundMiscDirAndFile := false
	if util.DirectoryExists(path.Join(t.sourceDir, "DCIM", "EOSMISC")) {
		// check for DCIM/EOSMISC/Mxxxx.CTG file
		logger.Debug(fmt.Sprintf("[CheckSource]: Testing for existence of DCIM/EOSMISC/Mxxxx.CTG file in volume '%s'", t.sourceDir))

		if exists, _ := util.RequireRegexFileMatch(path.Join(t.sourceDir, "DCIM", "EOSMISC"), `M(\d+).CTG`); exists {
			foundMiscDirAndFile = true
		}
	}

	if util.DirectoryExists(path.Join(t.sourceDir, "DCIM", "CANONMSC")) {
		// check for DCIM/EOSMISC/Mxxxx.CTG file
		logger.Debug(fmt.Sprintf("[CheckSource]: Testing for existence of DCIM/CANONMSC/Mxxxx.CTG file in volume '%s'", t.sourceDir))

		if exists, _ := util.RequireRegexFileMatch(path.Join(t.sourceDir, "DCIM", "CANONMSC"), `M(\d+).CTG`); exists {
			foundMiscDirAndFile = true
		}
	}

	if !foundMiscDirAndFile {
		logger.Debug("[CheckSource]: No '/DCIM/(EOSMISC|CANONMISC)/Mxxxx.CTG' file found, disqualified")
		return false
	}

	// check for DCIM/(\d+)(CANON|EOS)([A-Za-z0-9]+) directory
	logger.Debug(fmt.Sprintf(`[CheckSource]: Testing for existence of DCIM/(\d+)(CANON|EOS)([\w\d]{0,}) directory in volume '%s'`, t.sourceDir))
	if exists, _ := util.RequireRegexDirMatch(path.Join(t.sourceDir, "DCIM"), `(\d+)(CANON|EOS)([\w\d]{0,})`); !exists {
		logger.Debug(`[CheckSource]: No '(\d+)(CANON|EOS)([\w\d]{0,})/' directory found, disqualified`)
		return false
	}

	logger.Debug(fmt.Sprintf("[CheckSource]: Volume '%s' is compatible", t.sourceDir))
	return true
}

func (t *Processor) EnumerateFiles() []model.SourceFile {
	et, err := exiftool.NewExiftool()
	if err != nil {
		fmt.Printf("Error when intializing: %v\n", err)
		return nil
	}
	t.etHandle = et
	defer t.etHandle.Close()

	return t.scanDirectory(path.Join(t.sourceDir, "DCIM"), "DCIM")
}

func (t *Processor) readExif(imagePath string) *exiftool.FileMetadata {
	logger.Debug(fmt.Sprintf("Reading EXIF data from '%s'", imagePath))
	fileInfos := t.etHandle.ExtractMetadata(imagePath)

	// TODO: error handling
	if len(fileInfos) == 0 {
		return nil
	}

	// for k, v := range fileInfos[0].Fields {
	// 	fmt.Printf("[%v] %v\n", k, v)
	// }

	return &fileInfos[0]
}

// private functions
func (t *Processor) getCameraModel(exifData *exiftool.FileMetadata) string {
	camModelName := fmt.Sprintf("%v", exifData.Fields["Model"])
	return camModelName
}

func (t *Processor) getCaptureDate(exifData *exiftool.FileMetadata) time.Time {
	//[DateTimeOriginal] 2024:12:01 11:45:31
	dtmOriginal := fmt.Sprintf("%v", exifData.Fields["DateTimeOriginal"])[:10]

	zone, _ := time.Now().Zone()
	dtmStr := fmt.Sprintf("%s %s", dtmOriginal, zone)

	format := "2006:01:02 MST"
	date, err := time.Parse(format, dtmStr)

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

				if strings.HasSuffix(entry.Name(), "MOV") || strings.HasSuffix(entry.Name(), "MP4") {
					mediaType = "Video"
				}

				exif := t.readExif(fullPath)

				if exif == nil {
					logger.Warn(fmt.Sprintf("Could not read EXIF data for '%s', skipping!", fullPath))
					continue
				}

				newFile := model.SourceFile{
					FileName:     entry.Name(),
					SourcePath:   fullPath,
					MediaType:    mediaType,
					Size:         stat.Size(),
					SourceName:   t.getCameraModel(exif),
					CaptureDate:  t.getCaptureDate(exif),
					FileModTime:  stat.ModTime(),
					VolumeFormat: t.volumeFormat,
				}

				files = append(files, newFile)
			}
		}
	}

	return files
}
