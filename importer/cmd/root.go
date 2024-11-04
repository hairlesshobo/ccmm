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

package cmd

import (
	"log/slog"
	"os"

	"ccmm/importer/action"
	"ccmm/model"
	"ccmm/util"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands

var (
	rootCmd = &cobra.Command{
		Use:   "go-import-media",
		Short: "media importer",
		Long:  `Custom tool for identifying source media and importing to appropriate destination`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}

func init() {
	config := loadConfig()
	model.Config = config

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.Level(config.LogLevel),
	}))
	slog.SetDefault(logger)

	if config.ForceDryRun {
		logger = logger.With(slog.Bool("DryRun", config.ForceDryRun))
		slog.SetDefault(logger)

		slog.Info("Force dry run is ENABLED via config")
	}

	slog.Info("Configured data directory: " + config.LiveDataDir)

	// This starts the thread that actually processes the import queue
	// TODO: need to add a shutdown channel for clean termination
	// var (
	// shutdownChan    chan struct{}
	// )
	// shutdownChan = make(chan struct{})
	go action.ImportWorker()

	// TODO: add config entries to enable/disable importer server
	// TODO: add config entries to enable/disable localsend server
}

func loadConfig() model.ImporterConfig {
	config := model.DefaultImporterConfig
	config.LocalSend.Alias = util.GetHostname()

	util.ReadConfig(&config, true, false, "importer.yml")

	return config
}
