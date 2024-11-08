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

package server

import (
	"ccmm/model"
	"ccmm/util"
	"encoding/json"
	"fmt"
	"slices"

	"ccmm/util/sync"
	"net/http"
)

//
// private functions
//

func syncRequest(w http.ResponseWriter, r *http.Request) {
	config := getConfig(r)

	syncRequest := util.ReadJsonBody[model.SyncRequest](r)

	for _, serviceDateStr := range syncRequest.Services {
		theirFiles := syncRequest.ServiceFiles[serviceDateStr]
		myFiles := sync.ScanService(serviceDateStr, syncRequest.MediaTypes, config.DataDirs.Services)

		// look for files that we have and the they also have
		for _, myFile := range myFiles {
			idx := slices.IndexFunc(theirFiles, func(f model.SyncFile) bool { return f.FilePath == myFile.FilePath })

			// found the file
			if idx >= 0 {
				theirFile := &syncRequest.ServiceFiles[serviceDateStr][idx]

				if theirFile.Size == myFile.Size {
					if theirFile.FileModTime == myFile.FileModTime {
						theirFile.ClientAction = "none"
						theirFile.ServerAction = "none"
					} else {
						fmt.Printf("   My DTM: %v\n", myFile.FileModTime)
						fmt.Printf("Their DTM: %v\n", theirFile.FileModTime)
					}
				}

			}
		}

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(syncRequest)
}
