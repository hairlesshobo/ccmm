// =================================================================================
//
//		ccmm - https://www.foxhollow.cc/projects/ccmm/
//
//	 go-import-media, aka gim, is a tool for automatically importing media
//	 from removable disks into a predefined folder structure automatically.
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
	"log/slog"

	"github.com/hairlesshobo/go-mediainfo"
)

func MediaInfo_GetGeneralParameter(filePath string, parameter string) string {
	mi := mediainfo.New()
	defer mi.Close()
	if err := mi.Open(filePath); err != nil {
		slog.Error(fmt.Sprintf("Failed to open file '%s' for reading mediainfo", filePath))
		return ""
	}

	return mi.Get(mediainfo.StreamGeneral, 0, parameter)
}
