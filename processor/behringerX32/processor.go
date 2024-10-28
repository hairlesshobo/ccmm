package behringerX32

import (
	// "encoding/xml"
	"fmt"
	// "io"
	"log/slog"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/hairlesshobo/go-import-media/model"
	"github.com/hairlesshobo/go-import-media/util"
)

// TODO: create logger for this function?
const expectedVolumeName = "X32"
const mediaType = "Audio"

var fileMatchPatterns = [...]string{`R_(\d{8})-(\d{6}).wav`}

type Processor struct {
	sourceDir string
}

func (t *Processor) SetSourceDir(sourceDir string) {
	t.sourceDir = sourceDir
}

func (t *Processor) CheckSource() bool {
	slog.Debug(fmt.Sprintf("behringerX32.CheckSource: Beginning to test volume compatibility for '%s'", t.sourceDir))

	// verify volume label matches what is expected
	slog.Debug(fmt.Sprintf("behringerX32.CheckSource: Testing volume name at '%s'", t.sourceDir))
	if label := util.GetVolumeName(t.sourceDir); !strings.HasPrefix(label, expectedVolumeName) {
		slog.Debug(fmt.Sprintf("behringerX32.CheckSource: Volume label '%s' does not start with required '%s' value, disqualified", label, expectedVolumeName))
		return false
	}

	// check for recorded audio files
	slog.Debug(fmt.Sprintf("behringerX32.CheckSource: Testing for existence of '%s' file in volume '%s'", fileMatchPatterns[0], t.sourceDir))
	exists, _ := util.RequireRegexFileMatch(t.sourceDir, fileMatchPatterns[0])
	if !exists {
		slog.Debug(fmt.Sprintf("behringerX32.CheckSource: No '%s' file found, disqualified", fileMatchPatterns[0]))
		return false
	}

	slog.Debug(fmt.Sprintf("behringerX32.CheckSource: Volume '%s' is compatible", t.sourceDir))
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
		slog.Error(fmt.Sprintf("behringerX32.getCaptureDate: Failed to parse date '%s': %s", dtmStr, err.Error()))
	}

	return dtm
}

func scanDirectory(absoluteDirPath string, relativeDirPath string) []model.SourceFile {
	slog.Debug(fmt.Sprintf("behringerX32.scanDirectory: Scanning for source files at path '%s'", absoluteDirPath))

	var files []model.SourceFile

	// For this processor, we only care about .wav files
	entries, err := os.ReadDir(absoluteDirPath)

	if err != nil {
		slog.Error(fmt.Sprintf("behringerX32.scanDirectory: Error occured while scanning directory '%s': %s", absoluteDirPath, err.Error()))
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
				slog.Debug(fmt.Sprintf("behringerX32.scanDirectory: Matched file '%s'", fullPath))

				stat, _ := os.Stat(fullPath)

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
