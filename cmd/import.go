package cmd

import (
	"fmt"
	"log/slog"

	"github.com/hairlesshobo/go-import-media/action"
	"github.com/hairlesshobo/go-import-media/model"
	"github.com/spf13/cobra"
)

var (
	dryRun bool

	importCmd = &cobra.Command{
		Use:   "import [flags] volume_path",
		Short: "Import a specific volume",
		Long:  `This will allow you to perform an import on a single volume in interactive mode`,
		Args:  cobra.MinimumNArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			var importConfig model.ImportVolume
			importConfig.DryRun = dryRun
			importConfig.VolumePath = args[0]

			slog.Debug(fmt.Sprintf("%+v", importConfig))

			action.Import(importConfig)
		},
	}
)

func init() {
	importCmd.Flags().BoolVarP(&dryRun, "dry_run", "d", false, "Perform a dry-run import (don't copy anything)")

	rootCmd.AddCommand(importCmd)
}
