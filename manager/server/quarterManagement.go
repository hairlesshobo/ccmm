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

package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

//
// private functions
//

func getQuarters(w http.ResponseWriter, r *http.Request) {
	config := getConfig(r)

	entries, err := os.ReadDir(config.DataDirs.Services)

	if err != nil {
		slog.Error(fmt.Sprintf("failed to read serice directory: %s", err.Error()))
		http.Error(w, "failed to read serices directory", http.StatusInternalServerError)
		return
	}

	fmt.Printf("%+v\n", entries)
	res, err := json.Marshal(entries)
	if err != nil {
		slog.Error(fmt.Sprintf("json convert failed: %s", err.Error()))
		http.Error(w, "json convert failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write(res)
}
