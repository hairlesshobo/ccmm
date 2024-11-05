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
	"fmt"
	"log/slog"
	"os"

	"ccmm/importer/action"
	"ccmm/model"
	"ccmm/util"

	"github.com/spf13/cobra"
)

var (
	deviceAttachedIndividual bool
	deviceAttachedDryRun     bool
	deviceAttachedServer     string
	deviceAttachedNoUnmount  bool
	deviceAttachedNoPoweroff bool

	deviceAttachedCmd = &cobra.Command{
		Use:   "device_attached [flags] device_path",
		Short: "Process a device that was attached to the system.",
		Long:  `This is the fully automatic import process that can be triggered by udev to mount, import, unmount and power down an attached device.`,
		Args:  cobra.MinimumNArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			config := cmd.Context().Value(model.ImportConfigContext).(model.ImporterConfig)

			var deviceAttachedConfig model.DeviceAttached
			deviceAttachedConfig.DryRun = deviceAttachedDryRun || config.ForceDryRun
			deviceAttachedConfig.DevicePath = args[0]
			deviceAttachedConfig.NoUnmount = deviceAttachedNoUnmount
			deviceAttachedConfig.NoPoweroff = deviceAttachedNoPoweroff

			slog.Debug(fmt.Sprintf("%+v", deviceAttachedConfig))

			if deviceAttachedIndividual {
				action.DeviceAttached(config, deviceAttachedConfig)
			} else {
				// queue the import with the server intance
				uri := fmt.Sprintf("http://%s/device_attached", deviceAttachedServer)
				_, statusCode := util.CallServer(uri, deviceAttachedConfig)

				if statusCode != 201 {
					slog.Error("Unknown error occurred sending request")
					os.Exit(1)
				}
			}
		},
	}
)

func init() {
	deviceAttachedCmd.Flags().BoolVarP(&deviceAttachedIndividual, "individual", "i", false, "Run a single import without connecting to the running server")
	deviceAttachedCmd.Flags().BoolVarP(&deviceAttachedDryRun, "dry_run", "n", false, "Perform a dry-run import (don't copy anything)")
	deviceAttachedCmd.Flags().BoolVarP(&deviceAttachedNoUnmount, "no_unmount", "u", false, "Prevents the job from automatically unmounting the device when finished processing")
	deviceAttachedCmd.Flags().BoolVarP(&deviceAttachedNoPoweroff, "no_poweroff", "p", false, "Prevents thje job from automatically powering off the device when finished processing")
	deviceAttachedCmd.Flags().StringVarP(&deviceAttachedServer, "server", "s", "localhost:7273", "<host>:<port> -- If specified, connect to the specified server instance to queue an import")

	rootCmd.AddCommand(deviceAttachedCmd)
}
