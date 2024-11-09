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
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get hostname, error: %s", err.Error()))
		return ""
	}

	return hostname
}

func GetServiceQuarter(serviceDate time.Time) string {
	quarter := 0
	year := serviceDate.Year()
	month := int16(serviceDate.Month())

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

func ReadJsonBody[T any](r *http.Request) T {
	var obj T

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read request body: " + err.Error())
	}

	if err = json.Unmarshal(body, &obj); err != nil {
		slog.Error("Failed to unmarshal JSON: " + err.Error())
	}

	return obj
}
