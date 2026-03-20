package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/smilemilks2021/easy-web/internal/config"
	"github.com/smilemilks2021/easy-web/internal/cookie"
	"github.com/smilemilks2021/easy-web/internal/workflow"
)

func init() {
	cmd := &cobra.Command{
		Use:   "run <workflow.yml>",
		Short: "Execute a workflow file (chained HTTP requests with variables)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wfPath := args[0]

			// Load workflow file.
			wf, err := workflow.Load(wfPath)
			if err != nil {
				return fmt.Errorf("load workflow: %w", err)
			}

			// Determine base domain for cookie lookup.
			// Use base_url if set; otherwise try the first step's URL.
			lookupURL := wf.BaseURL
			if lookupURL == "" && len(wf.Steps) > 0 {
				lookupURL = wf.Steps[0].URL
			}

			var cookies []*cookie.Entry
			if lookupURL != "" {
				domain := cookieDomain(lookupURL)
				if domain != "" {
					store := cookie.NewCache(config.CacheDir())
					cookies, err = store.Load(domain)
					if err != nil {
						// Non-fatal: proceed without cookies.
						fmt.Printf("warning: could not load cookies for %s: %v\n", domain, err)
					}
				}
			}

			runner := workflow.NewRunner(cookies)
			failures, err := runner.Run(wf)
			if err != nil {
				return err
			}
			if failures > 0 {
				return fmt.Errorf("%d assertion(s) failed", failures)
			}
			return nil
		},
	}
	rootCmd.AddCommand(cmd)
}

// cookieDomain extracts the hostname from a URL string for cookie cache lookup.
func cookieDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Hostname()
}
