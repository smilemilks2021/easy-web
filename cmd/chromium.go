package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/smilemilks2021/easy-web/internal/browser/chromium"
	"github.com/smilemilks2021/easy-web/internal/config"
)

func init() {
	mgr := func() *chromium.Manager { return chromium.NewManager(config.ChromiumDir(), chromium.DefaultRevision) }
	cr := &cobra.Command{Use: "chromium", Short: "Manage embedded Chromium"}
	cr.AddCommand(
		&cobra.Command{
			Use: "list", Short: "List downloaded versions",
			RunE: func(cmd *cobra.Command, _ []string) error {
				revs, err := mgr().List()
				if err != nil {
					return err
				}
				for _, r := range revs {
					fmt.Println(r)
				}
				return nil
			}},
		&cobra.Command{
			Use: "info", Short: "Show Chromium config",
			Run: func(_ *cobra.Command, _ []string) { mgr().Info() }},
		&cobra.Command{
			Use:  "clean",
			Short: "Remove old versions",
			RunE: func(_ *cobra.Command, _ []string) error { return mgr().Clean() }},
		&cobra.Command{
			Use: "download [revision]", Short: "Download Chromium",
			RunE: func(_ *cobra.Command, args []string) error {
				rev := chromium.DefaultRevision
				if len(args) > 0 {
					rev = args[0]
				}
				_, err := chromium.Download(rev, config.ChromiumDir())
				return err
			}},
	)
	rootCmd.AddCommand(cr)
}
