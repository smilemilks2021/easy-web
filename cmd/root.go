package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/smilemilks2021/easy-web/internal/auth"
	"github.com/smilemilks2021/easy-web/internal/config"
	"github.com/smilemilks2021/easy-web/internal/cookie"
)

var appVersion, appCommit, appDate string

func SetVersion(v, c, d string) { appVersion, appCommit, appDate = v, c, d }

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "easy-web",
	Short: "Browser auth CLI — capture cookies, automate API calls",
	Long: `easy-web captures browser login cookies and replays them for
authenticated API requests. Supports 5 auth modes, API capture,
token extraction, and local cache management.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize config before any command runs
		config.Init()
		// Dynamically register --<name> flags from capture_headers config
		for name, ch := range config.C.CaptureHeaders {
			flagName := name
			headerName := ch.Header
			if !cmd.Root().Flags().Changed(flagName) {
				cmd.Root().PersistentFlags().Bool(flagName, false,
					fmt.Sprintf("Capture %q header (from capture_headers config)", headerName))
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		targetURL, _ := cmd.Flags().GetString("url")
		if targetURL == "" {
			return cmd.Help()
		}
		mode, _ := cmd.Flags().GetString("mode")
		noReuse, _ := cmd.Flags().GetBool("no-reuse-profile")
		verbose, _ := cmd.Flags().GetBool("verbose-auth")
		useEmbedded, _ := cmd.Flags().GetBool("use-embedded-chromium")

		result, err := auth.Resolve(targetURL, auth.Options{
			Mode:                mode,
			NoReuseProfile:      noReuse,
			VerboseAuth:         verbose,
			UseEmbeddedChromium: useEmbedded,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Got %d cookies for %s\n", len(result.Cookies), targetURL)
		for _, c := range result.Cookies {
			v := c.Value
			if len(v) > 20 {
				v = v[:20] + "..."
			}
			fmt.Printf("  %s=%s\n", c.Name, v)
		}

		// Save to cache
		store := cookie.NewCache(config.CacheDir())
		return store.Save(parseHost(targetURL), result.Cookies)
	},
}

func init() {
	rootCmd.PersistentFlags().String("config", "", "config file (default: ~/.easy-web.yaml)")
	rootCmd.PersistentFlags().StringP("url", "u", "", "Target URL")
	rootCmd.PersistentFlags().StringP("mode", "m", "auto", "Auth mode: auto|chromedp|browser|chrome|remote")
	rootCmd.PersistentFlags().Bool("no-reuse-profile", false, "Don't reuse Chrome profile")
	rootCmd.PersistentFlags().Bool("no-auto-close", false, "Keep browser open after login")
	rootCmd.PersistentFlags().Bool("use-embedded-chromium", false, "Use embedded Chromium")
	rootCmd.PersistentFlags().Bool("verbose-auth", false, "Verbose auth debug output")
	rootCmd.PersistentFlags().Bool("auth-token", false, "Capture Authorization header")
	rootCmd.PersistentFlags().Bool("extract-token", false, "Extract token from localStorage/sessionStorage")
}

func parseHost(rawURL string) string {
	for _, prefix := range []string{"https://", "http://"} {
		if len(rawURL) > len(prefix) && rawURL[:len(prefix)] == prefix {
			rawURL = rawURL[len(prefix):]
			break
		}
	}
	for i, ch := range rawURL {
		if ch == '/' || ch == '?' || ch == ':' {
			return rawURL[:i]
		}
	}
	return rawURL
}
