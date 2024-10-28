package canonEOS

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/hairlesshobo/go-import-media/model"
	"github.com/hairlesshobo/go-import-media/util"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

const expectedVolumeName = "EOS_DIGITAL"

// TODO: do i need to add support for multiple xxxCANON folders in the copy structure?

// TODO: add support for handling videos as well
var fileMatchPatterns = [...]string{
	`DCIM/(\d+)CANON/IMG_(\d+).CR2`,
	`DCIM/(\d+)CANON/MVI_(\d+).MOV`,
}

type Processor struct {
	sourceDir  string
	sourceName string
}

func (t *Processor) SetSourceDir(sourceDir string) {
	t.sourceDir = sourceDir
	t.sourceName = ""
}

func (t *Processor) CheckSource() bool {
	slog.Debug(fmt.Sprintf("canonEOS.CheckSource: Beginning to test volume compatibility for '%s'", t.sourceDir))

	// verify volume label matches what the camera sets
	slog.Debug(fmt.Sprintf("canonEOS.CheckSource: Testing volume name at '%s'", t.sourceDir))
	if label := util.GetVolumeName(t.sourceDir); label != expectedVolumeName {
		slog.Debug(fmt.Sprintf("canonEOS.CheckSource: Volume label '%s' does not match required '%s' value, disqualified", label, expectedVolumeName))
		return false
	}

	// check for /DCIM and /MISC and /DCIM/EOSMISC directories
	slog.Debug(fmt.Sprintf("canonEOS.CheckSource: Testing for required directories for volume '%s'", t.sourceDir))
	if !util.RequireDirs(t.sourceDir, []string{"DCIM", "MISC", "DCIM/EOSMISC"}) {
		slog.Debug("canonEOS.CheckSource: One or more required directories does not exist on source, disqualified")
		return false
	}

	// check for DCIM/(\d+)CANON directory
	slog.Debug(fmt.Sprintf("canonEOS.CheckSource: Testing for existence of DCIM/xxxCANON directory in volume '%s'", t.sourceDir))
	if exists, _ := util.RequireRegexDirMatch(path.Join(t.sourceDir, "DCIM"), `(\d+)CANON`); !exists {
		slog.Debug("canonEOS.CheckSource: No '/DCIM/xxxCANON/' directory found, disqualified")
		return false
	}

	// check for DCIM/EOSMISC/Mxxxx.CTG file
	slog.Debug(fmt.Sprintf("canonEOS.CheckSource: Testing for existence of DCIM/EOSMISC/Mxxxx.CTG file in volume '%s'", t.sourceDir))
	if exists, _ := util.RequireRegexFileMatch(path.Join(t.sourceDir, "DCIM", "EOSMISC"), `M(\d+).CTG`); !exists {
		slog.Debug("canonEOS.CheckSource: No '/DCIM/EOSMISC/Mxxxx.CTG' file found, disqualified")
		return false
	}

	slog.Debug(fmt.Sprintf("canonEOS.CheckSource: Volume '%s' is compatible", t.sourceDir))
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

	slog.Debug(fmt.Sprintf("Reading EXIF data from '%s'", imagePath))
	imageFile, err := os.Open(imagePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to open image file: %s", err.Error()))
		return ""
	}
	defer imageFile.Close()

	// Optionally register camera makenote data parsing - currently Nikon and
	// Canon are supported.
	exif.RegisterParsers(mknote.All...)

	x, err := exif.Decode(imageFile)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to decode exif data in image file: %s", err.Error()))
		return ""
	}

	camModel, _ := x.Get(exif.Model) // normally, don't ignore errors!
	camModelName, _ := camModel.StringVal()

	t.sourceName = camModelName
	return camModelName
}

func (t *Processor) getCaptureDate(dtm time.Time) time.Time {
	format := "2006-01-02 MST"
	date, err := time.Parse(format, dtm.Format(format))

	if err != nil {
		slog.Error(fmt.Sprintf("Failed to parse date, error: %s", err.Error()))
	}

	return date
}

func (t *Processor) scanDirectory(absoluteDirPath string, relativeDirPath string) []model.SourceFile {
	slog.Debug(fmt.Sprintf("canonEOS.scanDirectory: Scanning for source files at path '%s'", absoluteDirPath))

	var files []model.SourceFile

	// For this processor, we only care about .MXF files and the sidecar XML files
	// and we read the source name from the sidecar XML
	entries, err := os.ReadDir(absoluteDirPath)

	if err != nil {
		slog.Error(fmt.Sprintf("canonEOS.scanDirectory: Error occured while scanning directory '%s': %s", absoluteDirPath, err.Error()))
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
				slog.Debug(fmt.Sprintf("canonEOS.scanDirectory: Matched file '%s'", fullPath))

				stat, _ := os.Stat(fullPath)

				mediaType := "Photo"

				if strings.HasSuffix(entry.Name(), "MOV") {
					mediaType = "Video"
				}

				var newFile model.SourceFile
				newFile.FileName = entry.Name()
				newFile.SourcePath = fullPath
				newFile.MediaType = mediaType
				newFile.Size = stat.Size()
				newFile.SourceName = t.getCameraModel(fullPath)
				newFile.CaptureDate = t.getCaptureDate(stat.ModTime())
				newFile.FileModTime = stat.ModTime()

				files = append(files, newFile)
			}
		}
	}

	return files
}
