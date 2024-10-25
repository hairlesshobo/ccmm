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

	"github.com/hairlesshobo/go-import-media/model"
	"github.com/hairlesshobo/go-import-media/util"
)

// TODO: create logger for this function?
const expectedVolumeName = "CANON"
const mediaType = "Video"

type xmlResult struct {
	Value string `xml:"Device>ModelName"`
}

var fileMatchPatterns = [...]string{`CONTENTS/CLIPS(\d+)/(\w)(\d+)(\w)(\d+)_(\d{6})(\w{2})_CANON.(MXF|XML)`}

type Processor struct {
	sourceDir string
}

func (t *Processor) SetSourceDir(sourceDir string) {
	t.sourceDir = sourceDir
}

func (t *Processor) CheckSource() bool {
	slog.Debug(fmt.Sprintf("canonXA.CheckSource: Beginning to test volume compatibility for '%s'", t.sourceDir))

	// verify volume label matches what the camera sets
	slog.Debug(fmt.Sprintf("canonXA.CheckSource: Testing volume name at '%s'", t.sourceDir))
	_, label := path.Split(t.sourceDir)
	if label != expectedVolumeName {
		slog.Debug(fmt.Sprintf("canonXA.CheckSource: Volume label '%s' does not match required '%s' value, disqualified", label, expectedVolumeName))
		return false
	}

	// check for /CONTENTS and /DCIM directories
	slog.Debug(fmt.Sprintf("canonXA.CheckSource: Testing for required directories for volume '%s'", t.sourceDir))
	if !util.RequireDirs(t.sourceDir, []string{"CONTENTS", "DCIM"}) {
		slog.Debug("canonXA.CheckSource: One or more required directories does not exist on source, disqualified")
		return false
	}

	// check for CONTENTS/CLIPS(\d+)
	slog.Debug(fmt.Sprintf("canonXA.CheckSource: Testing for existence of CONTENTS/CLIPSxxx directory in volume '%s'", t.sourceDir))
	exists, clipsPath := util.RequireRegexDirMatch(path.Join(t.sourceDir, "CONTENTS"), `CLIPS(\d+)`)
	if !exists {
		slog.Debug("canonXA.CheckSource: No '/CONTENTS/CLIPSXXX/' directory found, disqualified")
		return false
	}

	// check for CONTENTS/CLIPS(\d+)/INDEX.MIF file
	slog.Debug(fmt.Sprintf("canonXA.CheckSource: Testing for existence of CONTENTS/CLIPSxxx/INDEX.MIF in volume '%s'", t.sourceDir))
	if !util.RequireFiles(clipsPath, []string{"INDEX.MIF"}) {
		slog.Debug("canonXA.CheckSource: INDEX.MIF file not found in CLIPS directory, disqualified")
		return false
	}

	slog.Debug(fmt.Sprintf("canonXA.CheckSource: Volume '%s' is compatible", t.sourceDir))
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

	slog.Debug(fmt.Sprintf("canonXA.getSourceName: Reading SourceName from file '%s'", sidecarFile))

	xmlFile, err := os.Open(sidecarFile)
	if err != nil {
		slog.Error(fmt.Sprintf("canonXA.getSourceName: Failed to open sidecar file '%s': %s", sidecarFile, err.Error()))
	} else {
		defer xmlFile.Close()

		byteValue, _ := io.ReadAll(xmlFile)
		err := xml.Unmarshal(byteValue, &x)
		if err != nil {
			slog.Error(fmt.Sprintf("error: %v", err))
		}
	}

	return x.Value
}

func getCaptureDate(fileName string) time.Time {
	datePart := strings.Split(fileName, "_")[1][:6]
	zone, _ := time.Now().Zone()
	dtm, err := time.Parse("060102 MST", fmt.Sprintf("%s %s", datePart, zone))

	if err != nil {
		slog.Error(fmt.Sprintf("canonXA.getCaptureDate: Failed to parse date '%s': %s", datePart, err.Error()))
	}

	return dtm
}

func scanDirectory(absoluteDirPath string, relativeDirPath string) []model.SourceFile {
	slog.Debug(fmt.Sprintf("canonXA.scanDirectory: Scanning for source files at path '%s'", absoluteDirPath))

	var files []model.SourceFile

	// For this processor, we only care about .MXF files and the sidecar XML files
	// and we read the source name from the sidecar XML
	entries, err := os.ReadDir(absoluteDirPath)

	if err != nil {
		slog.Error(fmt.Sprintf("canonXA.scanDirectory: Error occured while scanning directory '%s': %s", absoluteDirPath, err.Error()))
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
				slog.Debug(fmt.Sprintf("canonXA.scanDirectory: Matched file '%s'", fullPath))

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
