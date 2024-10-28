package jackRecorder

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/hairlesshobo/go-import-media/model"
	"github.com/hairlesshobo/go-import-media/util"
)

// TODO: create logger for this function?
const mediaType = "Audio"

var fileMatchPatterns = [...]string{`(.+).wav`}

type Processor struct {
	sourceDir string
}

func (t *Processor) SetSourceDir(sourceDir string) {
	t.sourceDir = sourceDir
}

func (t *Processor) CheckSource() bool {
	slog.Debug(fmt.Sprintf("jackRecorder.CheckSource: Beginning to test volume compatibility for '%s'", t.sourceDir))

	// check for jack directory
	slog.Debug(fmt.Sprintf("jackRecorder.CheckSource: Testing for required directories for volume '%s'", t.sourceDir))
	if !util.RequireDirs(t.sourceDir, []string{"jack"}) {
		slog.Debug("jackRecorder.CheckSource: One or more required directories does not exist on source, disqualified")
		return false
	}

	// check for jack/(\d{4})-(\d{2})-(\d{2})
	slog.Debug(fmt.Sprintf(`jackRecorder.CheckSource: Testing for existence of jack/(\d{4})-(\d{2})-(\d{2}) directory in volume '%s'`, t.sourceDir))
	exists, _ := util.RequireRegexDirMatch(path.Join(t.sourceDir, "jack"), `(\d{4})-(\d{2})-(\d{2})`)
	if !exists {
		slog.Debug(`jackRecorder.CheckSource: No '/jack/(\d{4})-(\d{2})-(\d{2})/' directory found, disqualified`)
		return false
	}

	slog.Debug(fmt.Sprintf("jackRecorder.CheckSource: Volume '%s' is compatible", t.sourceDir))
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
		slog.Error(fmt.Sprintf("jackRecorder.getCaptureDate: Failed to parse date '%s': %s", dtmStr, err.Error()))
	}

	return dtm
}

func scanDirectory(absoluteDirPath string, relativeDirPath string) []model.SourceFile {
	slog.Debug(fmt.Sprintf("jackRecorder.scanDirectory: Scanning for source files at path '%s'", absoluteDirPath))

	var files []model.SourceFile

	// For this processor, we only care about .wav files
	entries, err := os.ReadDir(absoluteDirPath)

	if err != nil {
		slog.Error(fmt.Sprintf("jackRecorder.scanDirectory: Error occured while scanning directory '%s': %s", absoluteDirPath, err.Error()))
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
				slog.Debug(fmt.Sprintf("jackRecorder.scanDirectory: Matched file '%s'", fullPath))

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
