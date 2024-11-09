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

package sync

import (
	"ccmm/model"
	"ccmm/util"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"
)

func ScanService(serviceDateStr string, allowedMediaTypes []string, serviceStorageRootPath string) []model.SyncFile {
	if len(serviceDateStr) < 10 {
		slog.Error(fmt.Sprintf("provided service doesn't appread to be a date: '%s'", serviceDateStr))
		return nil
	}

	date, err := time.Parse("2006-01-02", serviceDateStr)

	if err != nil {
		slog.Error(fmt.Sprintf("failed to parse service date '%s': %v", serviceDateStr, err))
		return nil
	}

	quarter := util.GetServiceQuarter(date)

	serviceRootDir := path.Join(serviceStorageRootPath, quarter, serviceDateStr)

	entries, err := os.ReadDir(serviceRootDir)

	if err != nil {
		slog.Error(fmt.Sprintf("Error occured while scanning directory '%s': %s", serviceRootDir, err.Error()))
		return nil
	}

	var allFiles []model.SyncFile

	for _, entry := range entries {
		fullPath := path.Join(serviceRootDir, entry.Name())

		// This top-level directory, relative to the service directory, defines the media type
		// being handled. For example: Audio, Video, Photo, etc. We only care about directories
		// and any files placed in this top level will be ignored. We also ignore any hidden
		// directories (those starting with a ".")
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			mediaType := entry.Name()

			if !mediaTypeRequested(allowedMediaTypes, mediaType) {
				slog.Info("Ignoring media type: " + mediaType)
				continue
			}

			allFiles = append(allFiles, scanDirectory(serviceDateStr, mediaType, fullPath, path.Join("/", entry.Name()))...)
		}
	}

	return allFiles
}

func mediaTypeRequested(allowedMediaTypes []string, requestedMediaType string) bool {
	if len(allowedMediaTypes) == 0 {
		return true
	}

	for _, allowedMediaType := range allowedMediaTypes {
		if strings.EqualFold(allowedMediaType, requestedMediaType) {
			return true
		}
	}

	return false
}

func scanDirectory(serviceDateStr string, mediaType string, absoluteDirPath string, relativeDirPath string) []model.SyncFile {
	slog.Debug(fmt.Sprintf("Scanning for files to sync at path '%s'", absoluteDirPath))

	var files []model.SyncFile

	entries, err := os.ReadDir(absoluteDirPath)

	if err != nil {
		slog.Error(fmt.Sprintf("Error occured while scanning directory '%s': %s", absoluteDirPath, err.Error()))
		return nil
	}

	for _, entry := range entries {
		fullPath := path.Join(absoluteDirPath, entry.Name())
		relativePath := path.Join(relativeDirPath, entry.Name())

		if entry.IsDir() {
			files = append(files, scanDirectory(serviceDateStr, mediaType, fullPath, path.Join(relativeDirPath, entry.Name()))...)
		} else {
			slog.Debug(fmt.Sprintf("[scanDirectory]: Found file '%s'", fullPath))

			stat, _ := os.Stat(fullPath)

			// TODO: add config setting for skip empty directories
			if stat.Size() == 0 {
				// TODO: add config setting for skip 0 byte files
				// slog.Info(fmt.Sprintf("[scanDirectory]: Skipping 0 byte file '%s'", fullPath))
				// continue
			}

			newFile := model.SyncFile{
				FileName:    entry.Name(),
				FilePath:    relativePath,
				Directory:   relativeDirPath,
				MediaType:   mediaType,
				Size:        stat.Size(),
				FileModTime: stat.ModTime(),
				Service:     serviceDateStr,

				// These will be set by the server so for now just an empty string
				ServerAction: "",
				ClientAction: "",
			}

			files = append(files, newFile)
		}
	}

	return files
}
