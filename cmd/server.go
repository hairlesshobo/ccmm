package cmd

import (
	"github.com/hairlesshobo/go-import-media/server"
	"github.com/spf13/cobra"
)

var (
	port int32

	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Start the importer in daemon mode (start REST server)",
		Long: `The importer consists of two main components. The server (this) which 
    waits for media inserted notifications and clients which are to be invoked by 
    udev or other integrations to notify the server when media has been inserted.`,

		Run: func(cmd *cobra.Command, args []string) {
			server.StartServer(port)
		},
	}
)

func init() {
	serverCmd.Flags().Int32VarP(&port, "port", "p", 7273, "Port to start the HTTP REST server on")

	rootCmd.AddCommand(serverCmd)
}
