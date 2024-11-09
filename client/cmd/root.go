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

package cmd

import (
	"context"
	"log/slog"
	"os"

	"ccmm/model"
	"ccmm/util"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands

var (
	rootCmd = &cobra.Command{
		Use:   "ccmm_client",
		Short: "GUI client for interacting with CC Media Manager",
		Long:  `The default behavior of this app, if ran without any command specified, is to execute the GUI client`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	config := loadConfig()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.Level(config.LogLevel),
	}))
	slog.SetDefault(logger)

	// Add the config to the cobra context so that it can be accessed by all child commands
	ctx := context.WithValue(context.TODO(), model.ClientConfigContext, config)
	err := rootCmd.ExecuteContext(ctx)

	if err != nil {
		os.Exit(1)
	}
}

func loadConfig() model.ClientConfig {
	config := model.DefaultClientConfig

	if config.ClientName == "" {
		config.ClientName = util.GetHostname()
	}

	util.ReadConfig(&config, true, false, "client.yml")

	return config
}
