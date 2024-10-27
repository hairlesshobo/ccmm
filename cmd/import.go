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
	importArg_dryRun     bool
	importArg_server     string
	importArg_individual bool

	importCmd = &cobra.Command{
		Use:   "import [flags] volume_path",
		Short: "Import a specific volume",
		Long:  `This will allow you to perform an import on a single volume in interactive mode`,
		Args:  cobra.MinimumNArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			var importConfig model.ImportVolume
			importConfig.DryRun = importArg_dryRun
			importConfig.VolumePath = args[0]

			slog.Debug(fmt.Sprintf("%+v", importConfig))

			if importArg_individual {
				action.Import(importConfig)
			} else {
				// queue the import with the server intance
				uri := fmt.Sprintf("http://%s/trigger_import", importArg_server)
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
	importCmd.Flags().BoolVarP(&importArg_dryRun, "dry_run", "d", false, "Perform a dry-run import (don't copy anything)")
	importCmd.Flags().BoolVarP(&importArg_individual, "individual", "i", false, "Run a single import without connecting to the running server")
	importCmd.Flags().StringVarP(&importArg_server, "server", "s", "localhost:7273", "<host>:<port> -- If specified, connect to the specified server instance to queue an import")

	rootCmd.AddCommand(importCmd)
}
