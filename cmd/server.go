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
package cmd

import (
	"gim/localsend"
	"gim/model"
	"gim/server"
	"gim/util"

	lsmodel "gim/localsend/model"

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
			startLocalsend("0.0.0.0", 53317)
			server.StartServer(server_listenAddress, server_listenPort)
		},
	}
)

func init() {
	serverCmd.Flags().StringVarP(&server_listenAddress, "listen", "l", model.Config.ListenAddress, "Local IP address to listen on")
	serverCmd.Flags().Int32VarP(&server_listenPort, "port", "p", model.Config.ListenPort, "Port to start the HTTP REST server on")

	rootCmd.AddCommand(serverCmd)
}

func startLocalsend(listenAddress string, port int) {
	config := lsmodel.NewConfig()

	config.Alias = util.GetHostname()
	config.ListenAddress = listenAddress
	config.ListenPort = port

	go localsend.RunServer(config)
}
