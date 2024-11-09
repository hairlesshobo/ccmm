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

package action

import (
	"ccmm/model"
	"ccmm/util/sync"
	"encoding/json"
	"fmt"
)

func DoSync(config model.ClientConfig, syncConfig model.SyncConfig) error {
	// TODO: add validation to ensure requested service date exists either locally or remotely

	if len(syncConfig.Services) == 0 {
		return fmt.Errorf("no service dates provided, nothing to do")
	}

	syncRequest := model.SyncRequest{
		ClientName:   config.ClientName,
		SyncType:     "request",
		Services:     syncConfig.Services,
		MediaTypes:   syncConfig.MediaTypes,
		ServiceFiles: make(map[string][]model.SyncFile),
	}

	for _, service := range syncConfig.Services {
		if len(service) < 10 {
			return fmt.Errorf("provided service doesn't appread to be a date: '%s'", service)
		}
		serviceDateStr := service[:10]

		files := sync.ScanService(service[:10], syncConfig.MediaTypes, config.DataDirs.Services)

		syncRequest.ServiceFiles[serviceDateStr] = files
	}

	if syncConfig.Dump {
		j, _ := json.MarshalIndent(syncRequest, "", "  ")
		fmt.Println(string(j))
	}

	return nil
}
