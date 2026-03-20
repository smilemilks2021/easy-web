package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/smilemilks2021/easy-web/internal/config"
	"github.com/smilemilks2021/easy-web/internal/cookie"
)

func init() {
	cacheCmd := &cobra.Command{Use: "cache", Short: "Manage cookie cache"}
	var clearDomain string
	var clearAll bool
	clearCmd := &cobra.Command{
		Use: "clear", Short: "Clear cached cookies",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cookie.NewCache(config.CacheDir())
			if clearAll {
				return c.Clear()
			}
			if clearDomain != "" {
				return c.Delete(clearDomain)
			}
			return fmt.Errorf("specify -d <domain> or --all")
		},
	}
	clearCmd.Flags().StringVarP(&clearDomain, "domain", "d", "", "Domain to clear")
	clearCmd.Flags().BoolVar(&clearAll, "all", false, "Clear all")
	cacheCmd.AddCommand(
		&cobra.Command{
			Use: "list", Short: "List cached domains",
			RunE: func(cmd *cobra.Command, args []string) error {
				c := cookie.NewCache(config.CacheDir())
				domains, err := c.List()
				if err != nil {
					return err
				}
				if len(domains) == 0 {
					fmt.Println("No cached cookies.")
					return nil
				}
				for _, d := range domains {
					fmt.Println(d)
				}
				return nil
			}},
		clearCmd,
	)
	rootCmd.AddCommand(cacheCmd)
}
