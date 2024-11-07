package action

import (
	"ccmm/model"
	"ccmm/util"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"
)

func DoSync(config model.ClientConfig, syncConfig model.SyncConfig) error {
	// TODO: add validation to ensure requested service date exists either locally or remotely

	if len(syncConfig.Services) == 0 {
		return fmt.Errorf("no service dates provided, nothing to do")
	}

	for _, service := range syncConfig.Services {
		if len(service) < 10 {
			return fmt.Errorf("provided service doesn't appread to be a date: '%s'", service)
		}

		serviceDateStr := service[0:10]
		date, err := time.Parse("2006-01-02", serviceDateStr)

		if err != nil {
			return fmt.Errorf("failed to parse service date '%s': %v", serviceDateStr, err)
		}

		quarter := util.GetServiceQuarter(date)

		serviceRootDir := path.Join(config.DataDirs.Services, quarter, serviceDateStr)

		if util.DirectoryExists(serviceRootDir) {
			// TODO: filter by media type
			files := scanDirectory("", serviceRootDir, "/")

			j, _ := json.MarshalIndent(files, "", "  ")
			fmt.Println(string(j))
		} else {
			// TODO: directory doesn't exist so we can skip scanning and just generate an empty file set
		}

		fmt.Println(serviceRootDir)
	}

	fmt.Println("meow")

	return nil
}

func scanDirectory(mediaType string, absoluteDirPath string, relativeDirPath string) []model.SyncFile {
	slog.Debug(fmt.Sprintf("Scanning for files to sync at path '%s'", absoluteDirPath))

	var files []model.SyncFile

	// For this processor, we only care about .wav files
	entries, err := os.ReadDir(absoluteDirPath)

	if err != nil {
		slog.Error(fmt.Sprintf("Error occured while scanning directory '%s': %s", absoluteDirPath, err.Error()))
		return nil
	}

	for _, entry := range entries {
		fullPath := path.Join(absoluteDirPath, entry.Name())
		relativePath := path.Join(relativeDirPath, entry.Name())

		if entry.IsDir() {
			// if we don't currently know the media tpye, that means we are at the top
			// level of a service scan and therefore use the directory name as the media
			// type that we then pass on to the recursive calls to this function
			subMediaType := mediaType
			if mediaType == "" {
				subMediaType = entry.Name()
			}

			files = append(files, scanDirectory(subMediaType, fullPath, path.Join(relativeDirPath, entry.Name()))...)
		} else {
			slog.Debug(fmt.Sprintf("[scanDirectory]: Matched file '%s'", fullPath))

			stat, _ := os.Stat(fullPath)

			if stat.Size() == 0 {
				slog.Info(fmt.Sprintf("[scanDirectory]: Skipping 0 byte file '%s'", fullPath))
				continue
			}

			newFile := model.SyncFile{
				FileName:    entry.Name(),
				FilePath:    relativePath,
				Directory:   relativeDirPath,
				MediaType:   mediaType,
				Size:        stat.Size(),
				FileModTime: stat.ModTime(),

				// These will be set by the server so for now just an empty string
				ServerAction: "",
				ClientAction: "",
			}

			files = append(files, newFile)
		}
	}

	return files
}
