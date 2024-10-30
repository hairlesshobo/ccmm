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
package model

var Config ConfigModel

type ConfigModel struct {
	LiveDataDir           string `yaml:"live_data_dir" envconfig:"LIVE_DATA_DIR"`
	LogLevel              int8   `yaml:"log_level" envconfig:"LOG_LEVEL"`
	UseSudo               string `yaml:"use_sudo" envconfig:"USE_SUDO"`
	ListenAddress         string `yaml:"listen_address" envconfig:"LISTEN_ADDRESS"`
	ListenPort            int32  `yaml:"listen_port" envconfig:"LISTEN_PORT"`
	ForceDryRun           bool   `yaml:"force_dry_run" envconfig:"FORCE_DRY_RUN"`
	DisableAutoProcessing bool   `yaml:"disable_auto_processing" envconfig:"DISABLE_AUTO_PROCESSING"`
}
