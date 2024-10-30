package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hairlesshobo/go-import-media/action"
	"github.com/hairlesshobo/go-import-media/model"
	"github.com/hairlesshobo/go-import-media/util"
	"github.com/spf13/cobra"
)

var (
	deviceAttached_individual bool
	deviceAttached_dryRun     bool
	deviceAttached_server     string

	deviceAttachedCmd = &cobra.Command{
		Use:   "device_attached [flags] device_path",
		Short: "Process a device that was attached to the system.",
		Long:  `This is the fully automatic import process that can be triggered by udev to mount, import, unmount and power down an attached device.`,
		Args:  cobra.MinimumNArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			var deviceAttachedConfig model.DeviceAttached
			deviceAttachedConfig.DryRun = deviceAttached_dryRun || model.Config.ForceDryRun
			deviceAttachedConfig.DevicePath = args[0]

			slog.Debug(fmt.Sprintf("%+v", deviceAttachedConfig))

			if deviceAttached_individual {
				action.DeviceAttached(deviceAttachedConfig)
			} else {
				// queue the import with the server intance
				uri := fmt.Sprintf("http://%s/device_attached", deviceAttached_server)
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
	// TODO: add options for disabling auto dismount, auto poweroff, etc
	deviceAttachedCmd.Flags().BoolVarP(&deviceAttached_individual, "individual", "i", false, "Run a single import without connecting to the running server")
	deviceAttachedCmd.Flags().BoolVarP(&deviceAttached_dryRun, "dry_run", "n", false, "Perform a dry-run import (don't copy anything)")
	deviceAttachedCmd.Flags().StringVarP(&deviceAttached_server, "server", "s", "localhost:7273", "<host>:<port> -- If specified, connect to the specified server instance to queue an import")

	rootCmd.AddCommand(deviceAttachedCmd)
}
