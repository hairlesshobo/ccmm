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
	"fmt"
	"log/slog"
	"os"

	"ccmm/client/action"
	"ccmm/model"

	"github.com/spf13/cobra"
)

var (
	syncArgDryRun bool
	syncArgServer string
	syncArgDump   bool

	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync one or more services with the manager",
		Long:  `This action will perform a synchronization between client and server, ensuring that both sides match accordingly`,

		Run: func(cmd *cobra.Command, args []string) {
			config := cmd.Context().Value(model.ClientConfigContext).(model.ClientConfig)

			var syncConfig model.SyncConfig
			syncConfig.DryRun = syncArgDryRun
			syncConfig.Dump = syncArgDump
			syncConfig.Services = []string{
				"2024-11-03",
			}
			syncConfig.MediaTypes = []string{
				"Video",
			}

			slog.Debug(fmt.Sprintf("%+v", syncConfig))

			err := action.DoSync(config, syncConfig)

			if err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}
		},
	}
)

func init() {
	// TODO: add args for data types, services, etc
	syncCmd.Flags().BoolVarP(&syncArgDryRun, "dry_run", "n", false, "Perform a dry-run import (don't copy anything)")
	syncCmd.Flags().BoolVarP(&syncArgDump, "dump", "d", false, "If set, dump the list of scanned files to json and exit (for debugging only)")
	syncCmd.Flags().StringVarP(&syncArgServer, "server", "s", "localhost:7273", "<host>:<port> -- If specified, connect to the specified server instance to queue an import")

	rootCmd.AddCommand(syncCmd)
}
