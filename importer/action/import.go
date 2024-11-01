// =================================================================================
//
//		gim - https://www.foxhollow.cc/projects/gim/
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
package action

import (
	"fmt"
	"log/slog"

	"ccmm/importer/processor"
	"ccmm/model"
	"ccmm/util"
)

func Import(params model.ImportVolume) bool {
	slog.Info(fmt.Sprintf("Beginning import for volume '%s'", params.VolumePath))

	if !util.DirectoryExists(params.VolumePath) {
		slog.Error(fmt.Sprintf("Directory not found: %s", params.VolumePath))
		return false
	}

	processors := processor.FindProcessors(params.VolumePath)
	processor.ProcessSources(processors, params.DryRun, params.Dump)

	slog.Info(fmt.Sprintf("Finished import for volume '%s'", params.VolumePath))

	return true
}
