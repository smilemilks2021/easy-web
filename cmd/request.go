package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/smilemilks2021/easy-web/internal/auth"
	"github.com/smilemilks2021/easy-web/internal/config"
	"github.com/smilemilks2021/easy-web/internal/cookie"
	"github.com/smilemilks2021/easy-web/internal/request"
)

func init() {
	cmd := &cobra.Command{
		Use: "request", Short: "Make an authenticated HTTP request",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetURL, _ := cmd.Flags().GetString("url")
			if targetURL == "" {
				return fmt.Errorf("--url is required")
			}
			method, _ := cmd.Flags().GetString("method")
			body, _ := cmd.Flags().GetString("data")
			hdrs, _ := cmd.Flags().GetStringArray("header")

			store := cookie.NewCache(config.CacheDir())
			entries, err := store.Load(parseHost(targetURL))
			if err != nil {
				mode, _ := cmd.Flags().GetString("mode")
				result, err := auth.Resolve(targetURL, auth.Options{Mode: mode})
				if err != nil {
					return fmt.Errorf("auth: %w", err)
				}
				entries = result.Cookies
			}

			extra := map[string]string{}
			for _, h := range hdrs {
				k, v, ok := strings.Cut(h, ":")
				if ok {
					extra[strings.TrimSpace(k)] = strings.TrimSpace(v)
				}
			}

			c := request.NewClient(entries, extra)
			resp, err := c.Do(method, targetURL, body, nil)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			fmt.Printf("HTTP %d\n", resp.StatusCode)
			io.Copy(cmd.OutOrStdout(), resp.Body)
			fmt.Println()
			return nil
		},
	}
	cmd.Flags().StringP("method", "X", "GET", "HTTP method")
	cmd.Flags().StringP("data", "d", "", "Request body")
	cmd.Flags().StringArrayP("header", "H", nil, "Extra header (Key: Value)")
	rootCmd.AddCommand(cmd)
}
