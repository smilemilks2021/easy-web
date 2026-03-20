package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/smilemilks2021/easy-web/internal/auth"
	"github.com/smilemilks2021/easy-web/internal/config"
)

func init() {
	var name string
	cmd := &cobra.Command{
		Use: "auth", Short: "Execute a multi_step_auth flow",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				if len(config.C.MultiStepAuth) == 1 {
					for k := range config.C.MultiStepAuth {
						name = k
					}
				} else if len(config.C.MultiStepAuth) == 0 {
					return fmt.Errorf("no multi_step_auth entries in ~/.easy-web.yaml")
				} else {
					fmt.Println("Available entries:")
					for k, v := range config.C.MultiStepAuth {
						fmt.Printf("  %s: %s\n", k, v.Description)
					}
					return fmt.Errorf("specify --name <entry>")
				}
			}
			return auth.RunMultiStep(name)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Name of multi_step_auth entry")
	rootCmd.AddCommand(cmd)
}
