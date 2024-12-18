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

package util

import (
	"fmt"
	"io"
	"os"
	"path"

	"ccmm/model"
)

func GetDestinationDirectoryRelative(sourceFile model.SourceFile) string {
	quarter := GetServiceQuarter(sourceFile.CaptureDate)
	serviceDate := sourceFile.CaptureDate.Format("2006-01-02")

	// TODO: add ability to configure destination folder structure
	// return path.Join("_Services", quarter, serviceDate, sourceFile.MediaType, sourceFile.SourceName)
	return path.Join(quarter, serviceDate, sourceFile.MediaType, sourceFile.SourceName)
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
