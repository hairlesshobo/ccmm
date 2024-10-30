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
package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func CallServer(uri string, body any) ([]byte, int) {
	slog.Debug(fmt.Sprintf("util.CallServer: Calling URL '%s'", uri))

	jsonStr, _ := json.Marshal(body)
	slog.Debug(fmt.Sprintf("util.CallServer: Sending JSON body: '%s'", string(jsonStr)))

	req, _ := http.NewRequest("POST", uri, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error(fmt.Sprintf("util.CallServer: Error occurred sending request: %s", err.Error()))
		panic(err)
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	// fmt.Println("response Headers:", resp.Header)
	slog.Debug(fmt.Sprintf("util.CallServer: Response status '%s'", resp.Status))
	slog.Debug(fmt.Sprintf("util.CallServer: Response body '%s'", string(responseBody)))

	return responseBody, resp.StatusCode
}
