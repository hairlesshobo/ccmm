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
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func ReadConfig(cfg interface{}, readYaml bool, readEnv bool, configFileName string) {
	if readYaml {
		readFile(cfg, configFileName)
	}

	// TODO: re-add support for reading from environment variables
	// TODO: add support for watching config and reloading in realtime
}

func processError(err error) {
	slog.Error(err.Error())
	os.Exit(2)
}

func readFile(cfg interface{}, configFileName string) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	configPath := os.Getenv("CONFIG_FILE")

	if configPath == "" {
		binPath, _ := os.Executable()
		binDir := filepath.Dir(binPath)
		sidecarPath := path.Join(binDir, configFileName)

		if FileExists(sidecarPath) {
			configPath = sidecarPath
		} else {
			homeDir, _ := os.UserHomeDir()
			homeDotConfigPath := path.Join(homeDir, ".config", "ccmm", configFileName)

			if FileExists(homeDotConfigPath) {
				configPath = homeDotConfigPath
			}
		}
	}

	if configPath == "" {
		logger.Error("No config file found, using defaults")
		return
	}

	if !FileExists(configPath) {
		logger.Error("The specified config file does not exist: " + configPath)
		os.Exit(1)
	}

	logger.Info("Reading config from " + configPath)

	f, err := os.Open(configPath)
	if err != nil {
		processError(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		processError(err)
	}
}
