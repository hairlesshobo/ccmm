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

	"ccmm/importer/action"
	"ccmm/model"
	"ccmm/util"

	"github.com/spf13/cobra"
)

var (
	importArgIndividual bool
	importArgDryRun     bool
	importArgServer     string
	importArgDump       bool

	importCmd = &cobra.Command{
		Use:   "import [flags] volume_path",
		Short: "Import a specific volume",
		Long:  `This will allow you to perform an import on a single volume in interactive mode`,
		Args:  cobra.MinimumNArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			config := cmd.Context().Value(model.ImportConfigContext).(model.ImporterConfig)

			var importConfig model.ImportVolume
			importConfig.DryRun = importArgDryRun || config.ForceDryRun
			importConfig.VolumePath = args[0]
			importConfig.Dump = importArgDump

			slog.Debug(fmt.Sprintf("%+v", importConfig))

			if importArgIndividual {
				action.Import(config, importConfig, func(_ *action.ImportQueueItem) {})
			} else {
				// queue the import with the server intance
				uri := fmt.Sprintf("http://%s/trigger_import", importArgServer)
				_, statusCode := util.CallServer(uri, importConfig)

				if statusCode != 201 {
					slog.Error("Unknown error occurred sending request")
					os.Exit(1)
				}
			}
		},
	}
)

func init() {
	importCmd.Flags().BoolVarP(&importArgIndividual, "individual", "i", false, "Run a single import without connecting to the running server")
	importCmd.Flags().BoolVarP(&importArgDryRun, "dry_run", "n", false, "Perform a dry-run import (don't copy anything)")
	importCmd.Flags().BoolVarP(&importArgDump, "dump", "d", false, "If set, dump the list of scanned files to json and exit (for debugging only)")
	importCmd.Flags().StringVarP(&importArgServer, "server", "s", "localhost:7273", "<host>:<port> -- If specified, connect to the specified server instance to queue an import")

	rootCmd.AddCommand(importCmd)
}
