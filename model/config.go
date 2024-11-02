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

// Global configuration object
//
// Deprecated: globals need to die
var Config ImporterConfig

type ImporterConfig struct {
	LiveDataDir           string          `yaml:"live_data_dir"`
	LogLevel              int8            `yaml:"log_level"`
	ListenAddress         string          `yaml:"listen_address"`
	ListenPort            int32           `yaml:"listen_port"`
	ForceDryRun           bool            `yaml:"force_dry_run"`
	DisableAutoProcessing bool            `yaml:"disable_auto_processing"`
	EnabledProcessors     []string        `yaml:"enabled_processors"`
	LocalSend             LocalSendConfig `yaml:"localsend"`
}

type LocalSendConfig struct {
	Alias               string   `yaml:"alias,omitempty"`
	StoragePath         string   `yaml:"storage_path"`
	AppendSenderAlias   bool     `yaml:"append_sender_alias"`
	ListenAddress       string   `yaml:"listen_address,omitempty"`
	ListenPort          int      `yaml:"listen_port,omitempty"`
	UdpBroadcastAddress string   `yaml:"udp_broadcast_address,omitempty"`
	UdpBroadcastPort    int      `yaml:"udp_broadcast_port,omitempty"`
	AllowedAliases      []string `yaml:"allowed_aliases"`
	RequirePassword     string   `yaml:"require_password"`
}

var DefaultImporterConfig = ImporterConfig{
	LiveDataDir:           "./uploads",
	LogLevel:              0,
	ListenAddress:         "127.0.0.1",
	ListenPort:            7273,
	ForceDryRun:           false,
	DisableAutoProcessing: false,
	EnabledProcessors:     []string{},
	LocalSend: LocalSendConfig{
		Alias:               "",
		StoragePath:         "./uloads",
		AppendSenderAlias:   true,
		ListenAddress:       "0.0.0.0",
		ListenPort:          53317,
		UdpBroadcastAddress: "224.0.0.167",
		UdpBroadcastPort:    53317,
		AllowedAliases:      []string{"__ALL__"},
		RequirePassword:     "",
	},
}

type ManagerConfig struct {
	DataDir       string `yaml:"data_dir"`
	LogLevel      int8   `yaml:"log_level"`
	ListenAddress string `yaml:"listen_address"`
	ListenPort    int32  `yaml:"listen_port"`
	ForceReadOnly bool   `yaml:"force_read_only"`
}

var DefaultManagerConfig = ManagerConfig{
	DataDir:       "./uploads",
	LogLevel:      -4,
	ListenAddress: "0.0.0.0",
	ListenPort:    7280,
	ForceReadOnly: false,
}
