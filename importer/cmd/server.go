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
	"ccmm/importer/localsend"
	"ccmm/importer/server"
	"ccmm/model"
	"ccmm/util"
	"fmt"
	"log/slog"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	serverListenAddress string
	serverListenPort    int32

	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Start the importer in daemon mode (start REST server)",
		Long: `The importer consists of two main components. The server (this) which 
    waits for media inserted notifications and clients which are to be invoked by 
    udev or other integrations to notify the server when media has been inserted.`,

		Run: func(cmd *cobra.Command, _ []string) {
			config := cmd.Context().Value(model.ImportConfigContext).(model.ImporterConfig)

			go localsend.RunServer(config.LocalSend, func(outputDir string) {
				slog.Info(fmt.Sprintf("Triggering import of '%s' directory", outputDir))

				var importConfig model.ImportVolume
				importConfig.VolumePath = outputDir

				// queue the import with the server intance
				// TODO: should be able to provide remote server address here in case i
				// TODO: want to split localsend and import functionality between two systems
				uri := fmt.Sprintf("http://%s:%d/trigger_import", config.ListenAddress, config.ListenPort)
				_, statusCode := util.CallServer(uri, importConfig)

				if statusCode != 201 {
					slog.Error("Unknown error occurred sending request")
					os.Exit(1)
				}
			})

			// special handling for mac systems that will spin up a diskutil watcher to
			// trigger imports
			if runtime.GOOS == "darwin" {
				go util.WatchForDeviceAttached(func(devicePath, volumePath string) {
					deviceAttachedConfig := model.DeviceAttached{
						AlreadyMounted: true,
						DryRun:         config.ForceDryRun,
						DevicePath:     devicePath,
						MountPath:      volumePath,
					}

					// queue the import with the server intance
					uri := fmt.Sprintf("http://%s/device_attached", deviceAttachedServer)
					_, statusCode := util.CallServer(uri, deviceAttachedConfig)

					if statusCode != 201 {
						slog.Error("Unknown error occurred sending request")
						os.Exit(1)
					}
				})
			}

			server.StartServer(config, serverListenAddress, serverListenPort)
		},
	}
)

func init() {
	serverCmd.Flags().StringVarP(&serverListenAddress, "listen", "l", model.DefaultImporterConfig.ListenAddress, "Local IP address to listen on")
	serverCmd.Flags().Int32VarP(&serverListenPort, "port", "p", model.DefaultImporterConfig.ListenPort, "Port to start the HTTP REST server on")

	rootCmd.AddCommand(serverCmd)
}
