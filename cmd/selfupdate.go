package cmd

import (
	"context"
	"fmt"

	selfupdate "github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use: "selfupdate", Short: "Update easy-web to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			updater, err := selfupdate.NewUpdater(selfupdate.Config{})
			if err != nil {
				return err
			}
			latest, found, err := updater.DetectLatest(context.Background(),
				selfupdate.ParseSlug("smilemilks2021/easy-web"))
			if err != nil {
				return fmt.Errorf("detect latest: %w", err)
			}
			if !found || latest.LessOrEqual(appVersion) {
				fmt.Printf("easy-web %s is already latest.\n", appVersion)
				return nil
			}
			fmt.Printf("Updating to %s...\n", latest.Version())
			exe, err := selfupdate.ExecutablePath()
			if err != nil {
				return err
			}
			if err := updater.UpdateTo(context.Background(), latest, exe); err != nil {
				return fmt.Errorf("update failed: %w", err)
			}
			fmt.Printf("Updated to %s.\n", latest.Version())
			return nil
		},
	})
}
