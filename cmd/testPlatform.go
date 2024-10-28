package cmd

import (
	"github.com/hairlesshobo/go-import-media/util"
	"github.com/spf13/cobra"
)

var (
	testPlatformCmd = &cobra.Command{
		Use:   "test_platform",
		Short: "Test the platform you are running on for support",
		Long:  `Quick check to see if the running platform is supported. Only used for debugging while building`,

		Run: func(cmd *cobra.Command, args []string) {
			util.TestPlatform()
		},
	}
)

func init() {

	rootCmd.AddCommand(testPlatformCmd)
}
