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
	"net/http"

	"ccmm/importer/action"
	"ccmm/model"
	"ccmm/util"
)

//
// private functions
//

func importPost(config model.ImporterConfig, w http.ResponseWriter, r *http.Request) {
	importConfig := util.ReadJsonBody[model.ImportVolume](r)

	importConfig.DryRun = importConfig.DryRun || config.ForceDryRun

	if !util.DirectoryExists(importConfig.VolumePath) {
		w.WriteHeader(500)
	} else {
		action.Import(config, importConfig, func(_ *action.ImportQueueItem) {})
		w.WriteHeader(201)
	}

}
