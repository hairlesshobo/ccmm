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
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"runtime"
	"strings"
)

func platformNotSupported(target any) {
	funcName := getFunctionName(target)
	slog.Error(fmt.Sprintf("[%s] Function not supported on platform '%s'", funcName, runtime.GOOS))
	os.Exit(1)
}

func getFunctionName(i interface{}) string {
	parts := strings.Split(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name(), "/")
	return strings.Join(parts[3:], "/")
}
