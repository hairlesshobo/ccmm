package util

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/hairlesshobo/go-import-media/model"
)

func GetQuarter(sourceFile model.SourceFile) string {
	quarter := 0
	year := sourceFile.CaptureDate.Year()
	month := int16(sourceFile.CaptureDate.Month())

	if month >= 1 && month <= 3 {
		quarter = 1
	} else if month >= 4 && month <= 6 {
		quarter = 2
	} else if month >= 7 && month <= 9 {
		quarter = 3
	} else {
		quarter = 4
	}

	return fmt.Sprintf("%d Q%d", year, quarter)
}

func GetDestinationDirectoryRelative(sourceFile model.SourceFile) string {
	quarter := GetQuarter(sourceFile)
	serviceDate := sourceFile.CaptureDate.Format("2006-01-02")

	return path.Join("_Services", quarter, serviceDate, sourceFile.MediaType, sourceFile.SourceName)
}

func GetDestinationDirectory(destRootDir string, sourceFile model.SourceFile) string {
	return path.Join(destRootDir, GetDestinationDirectoryRelative(sourceFile))
}

func CopyFile(sourcePath string, destPath string) (int64, error) {
	sourceFileStat, err := os.Stat(sourcePath)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", sourcePath)
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(destPath)
	if err != nil {
		return 0, err
	}
	defer destination.Close()

	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func FileExists(filePath string) bool {
	if stat, err := os.Stat(filePath); err == nil && stat.Mode().IsRegular() {
		return true
	}

	return false
}
