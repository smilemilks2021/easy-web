package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("easy-web %s\n  Build time: %s\n  Git commit: %s\n  Platform:   %s/%s\n",
				appVersion, appDate, appCommit, runtime.GOOS, runtime.GOARCH)
		},
	})
}
