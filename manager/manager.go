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

package main

import (
	"ccmm/manager/server"
	"ccmm/model"
	"ccmm/util"
	"log/slog"
	"os"
)

func main() {
	config := loadConfig()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.Level(config.LogLevel),
	}))
	slog.SetDefault(logger)

	if config.ForceReadOnly {
		logger = logger.With(slog.Bool("DryRun", config.ForceReadOnly))
		slog.SetDefault(logger)

		slog.Info("Force dry run is ENABLED via config")
	}

	slog.Info("Configured services data directory: " + config.DataDirs.Services)

	server.StartServer(config)
}

func loadConfig() model.ManagerConfig {
	config := model.DefaultManagerConfig

	util.ReadConfig(&config, true, false, "manager.yml")

	return config
}
