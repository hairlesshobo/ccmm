package processor

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

type CanonXA struct {
	sourceDir string
}

func (t *CanonXA) SetSourceDir(sourceDir string) {
	t.sourceDir = sourceDir
}

func (t *CanonXA) CheckSource() bool {
	// verify volume label matches what the camera sets
	_, label := path.Split(t.sourceDir)
	if label != expectedVolumeName {
		slog.Debug(fmt.Sprintf("CanonXA70: Volume label '%s' does not match required '%s' value, disqualified", label, expectedVolumeName))
		return false
	}

	// check for /CONTENTS and /DCIM directories
	if !util.RequireDirs(t.sourceDir, []string{"CONTENTS", "DCIM"}) {
		slog.Debug("CanonXA: One or more required directories does not exist on source, disqualified")
		return false
	}

	// check for CONTENTS/CLIPS(\d+)
	exists, clipsPath := util.RequireRegexDirMatch(path.Join(t.sourceDir, "CONTENTS"), `CLIPS(\d+)`)
	if !exists {
		slog.Debug("CanonXA: No '/CONTENTS/CLIPSXXX/' directory found, disqualified")
		return false
	}

	// check for CONTENTS/CLIPS(\d+)/INDEX.MIF file
	if !util.RequireFiles(clipsPath, []string{"INDEX.MIF"}) {
		slog.Debug("CanonXA: INDEX.MIF file not found in CLIPS directory, disqualified")
		return false
	}

	return true
}

func (t *CanonXA) EnumerateFiles() []model.SourceFile {
	return scanDirectory(path.Join(t.sourceDir, "CONTENTS"), "CONTENTS")
}

func (t *CanonXA) ProcessSource() {

}

func getSourceName(mediaPath string) string {
	sidecarFile := mediaPath
	if strings.HasSuffix(sidecarFile, "MXF") {
		sidecarFile = strings.TrimSuffix(sidecarFile, "MXF") + "XML"
	}

	x := xmlResult{"Unknown"}

	xmlFile, err := os.Open(sidecarFile)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to open sidecar file '%s': %s", sidecarFile, err.Error()))
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
	// A095C001_241020HP_CANON
	datePart := strings.Split(fileName, "_")[1][:6]
	zone, _ := time.Now().Zone()
	dtm, err := time.Parse("060102 MST", fmt.Sprintf("%s %s", datePart, zone))

	if err != nil {
		slog.Error(fmt.Sprintf("Failed to parse date '%s': %s", datePart, err.Error()))
	}

	return dtm
}

func scanDirectory(absoluteDirPath string, relativeDirPath string) []model.SourceFile {
	var files []model.SourceFile

	// For this processor, we only care about .MXF files and the sidecar XML files
	// and we read the source name from the sidecar XML
	entries, err := os.ReadDir(absoluteDirPath)

	if err != nil {
		slog.Error(fmt.Sprintf("Error occured while scanning directory '%s': %s", absoluteDirPath, err.Error()))
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
				stat, _ := os.Stat(fullPath)

				var newFile model.SourceFile
				newFile.FileName = entry.Name()
				newFile.SourcePath = fullPath
				newFile.MediaType = mediaType
				newFile.Size = uint64(stat.Size())
				newFile.SourceName = getSourceName(fullPath)
				newFile.CaptureDate = getCaptureDate(entry.Name())

				files = append(files, newFile)
			}
		}
	}

	return files
}
