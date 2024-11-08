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

package model

import "time"

// SourceFile desribes a file that is identified to be imported by
// the importer tool
type SourceFile struct {
	FileName    string
	SourcePath  string
	Size        int64
	MediaType   string
	SourceName  string
	CaptureDate time.Time
	FileModTime time.Time
}

// SyncRequest describes a request to synchronize between client
// and manager
type SyncRequest struct {
	// ClientName is the hostname of the client machine requesting the sync
	ClientName string `json:"client_name"`

	// SyncType is what phase of the sync process does this describe.
	// Valid values:
	//   - request
	//   - plan
	// TODO: make this an enum?
	SyncType     string                `json:"sync_type"`
	Services     []string              `json:"services"`
	MediaTypes   []string              `json:"media_types"`
	ServiceFiles map[string][]SyncFile `json:"service_files"`
}

// SyncFile describes a file that is to be synchronized between the managet
// and the client
type SyncFile struct {
	FileName    string    `json:"file_name"`
	FilePath    string    `json:"file_path"`
	Directory   string    `json:"directory"`
	MediaType   string    `json:"media_type"` // Audio, Photo, Video, etc...
	Size        int64     `json:"size"`
	FileModTime time.Time `json:"mod_dtm"`
	Service     string    `json:"service"`

	// possible actions:
	//   none (file exists in both locations) - no transmission required
	//   update (file needs to be updated on the manager or the client side) - requires send on other side
	//   add (file doesn't exist and needs to be added on the manager or client side) - requires send on other side
	//   send (file needs to be sent from the manager or client side to the opposing side) - requires "add" or "update" on other side
	//   moved (file needs to be moved on either the manager or client side to a new location) - no transmission required
	//   delete (file needs to be deleted on either the manager or client side) - no transmission required
	// TODO: make these enums?
	ServerAction string `json:"manager_action"`
	ClientAction string `json:"client_action"`
}
