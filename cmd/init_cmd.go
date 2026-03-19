package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	var force bool
	cmd := &cobra.Command{
		Use: "init", Short: "Initialize easy-web configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := os.UserHomeDir()
			dest := filepath.Join(home, ".easy-web.yaml")
			if _, err := os.Stat(dest); err == nil && !force {
				fmt.Printf("%s already exists. Use --force to overwrite.\n", dest)
				return nil
			}
			defaults := `mode: "chromedp"
port: 8080
debug_port: 9222
auto_close: true
no_reuse_profile: false
domains:
  jwt_cookies: [token, access_token, authorization]
  localStorage_keys: [token, accessToken, access_token]
capture_headers: {}
multi_step_auth: {}
`
			if err := os.WriteFile(dest, []byte(defaults), 0600); err != nil {
				return err
			}
			fmt.Printf("Configuration written to %s\n", dest)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing config")
	rootCmd.AddCommand(cmd)
}
