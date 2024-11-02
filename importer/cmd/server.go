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

	"github.com/spf13/cobra"
)

var (
	server_listenAddress string
	server_listenPort    int32

	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Start the importer in daemon mode (start REST server)",
		Long: `The importer consists of two main components. The server (this) which 
    waits for media inserted notifications and clients which are to be invoked by 
    udev or other integrations to notify the server when media has been inserted.`,

		Run: func(cmd *cobra.Command, args []string) {
			go localsend.RunServer(model.Config.LocalSend, func(outputDir string) {
				slog.Info(fmt.Sprintf("Triggering import of '%s' directory", outputDir))

				var importConfig model.ImportVolume
				importConfig.VolumePath = outputDir

				// queue the import with the server intance
				// TODO: should be able to provide remote server address here in case i
				// TODO: want to split localsend and import functionality between two systems
				uri := fmt.Sprintf("http://%s:%d/trigger_import", model.Config.ListenAddress, model.Config.ListenPort)
				_, statusCode := util.CallServer(uri, importConfig)

				if statusCode != 201 {
					slog.Error("Unknown error occurred sending request")
					os.Exit(1)
				}
			})
			server.StartServer(server_listenAddress, server_listenPort)
		},
	}
)

func init() {
	serverCmd.Flags().StringVarP(&server_listenAddress, "listen", "l", model.Config.ListenAddress, "Local IP address to listen on")
	serverCmd.Flags().Int32VarP(&server_listenPort, "port", "p", model.Config.ListenPort, "Port to start the HTTP REST server on")

	rootCmd.AddCommand(serverCmd)
}
