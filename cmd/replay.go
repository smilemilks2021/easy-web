package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/smilemilks2021/easy-web/internal/auth"
	"github.com/smilemilks2021/easy-web/internal/browser"
	"github.com/smilemilks2021/easy-web/internal/config"
	"github.com/smilemilks2021/easy-web/internal/cookie"
	"github.com/smilemilks2021/easy-web/internal/request"
)

// replayDir returns the directory where captured request snapshots are stored.
func replayDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".easy-web", "replay")
}

// replayFile returns the path to the replay JSON file for the given domain.
func replayFile(domain string) string {
	safe := strings.ReplaceAll(strings.ReplaceAll(domain, ":", "_"), "/", "_")
	return filepath.Join(replayDir(), safe+".json")
}

// SaveReplayRequests persists captured requests to ~/.easy-web/replay/<domain>.json
// so that the replay command can read them later.
func SaveReplayRequests(domain string, reqs []*browser.CapturedRequest) error {
	if err := os.MkdirAll(replayDir(), 0700); err != nil {
		return fmt.Errorf("create replay dir: %w", err)
	}
	data, err := json.MarshalIndent(reqs, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal replay data: %w", err)
	}
	if err := os.WriteFile(replayFile(domain), data, 0600); err != nil {
		return fmt.Errorf("write replay file: %w", err)
	}
	return nil
}

func init() {
	cmd := &cobra.Command{
		Use:   "replay",
		Short: "Replay previously captured API requests for a domain",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetURL, _ := cmd.Flags().GetString("url")
			if targetURL == "" {
				return fmt.Errorf("--url is required")
			}
			filter, _ := cmd.Flags().GetString("filter")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			domain := parseHost(targetURL)
			path := replayFile(domain)

			data, err := os.ReadFile(path)
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("no captured requests found for %q; run 'easy-web capture -u %s' first", domain, targetURL)
				}
				return fmt.Errorf("read replay file: %w", err)
			}

			var reqs []*browser.CapturedRequest
			if err := json.Unmarshal(data, &reqs); err != nil {
				return fmt.Errorf("parse replay file: %w", err)
			}

			// Apply filter.
			var filtered []*browser.CapturedRequest
			for _, r := range reqs {
				if filter == "" || strings.Contains(r.URL, filter) {
					filtered = append(filtered, r)
				}
			}

			if len(filtered) == 0 {
				fmt.Println("No requests match the given filter.")
				return nil
			}

			if dryRun {
				fmt.Printf("Dry run — would replay %d requests:\n", len(filtered))
				for i, r := range filtered {
					fmt.Printf("  [%d] %s %s\n", i+1, r.Method, r.URL)
				}
				return nil
			}

			// Load cookies.
			store := cookie.NewCache(config.CacheDir())
			entries, err := store.Load(domain)
			if err != nil {
				return fmt.Errorf("reading cookie cache: %w", err)
			}
			if len(entries) == 0 {
				mode, _ := cmd.Flags().GetString("mode")
				result, err := auth.Resolve(targetURL, auth.Options{Mode: mode})
				if err != nil {
					return fmt.Errorf("auth: %w", err)
				}
				entries = result.Cookies
			}

			client, err := request.NewClient(entries, nil)
			if err != nil {
				return fmt.Errorf("create client: %w", err)
			}

			fmt.Printf("Replaying %d requests for %s\n", len(filtered), domain)
			for i, r := range filtered {
				resp, err := client.Do(r.Method, r.URL, r.Body, r.Headers)
				if err != nil {
					fmt.Printf("  [%d] %s %s  ERROR: %v\n", i+1, r.Method, r.URL, err)
					continue
				}
				raw, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				summary := string(raw)
				if len(summary) > 120 {
					summary = summary[:120] + "..."
				}
				// Replace newlines for compact single-line summary.
				summary = strings.ReplaceAll(strings.ReplaceAll(summary, "\r\n", " "), "\n", " ")
				fmt.Printf("  [%d] %s %s  HTTP %d  %s\n", i+1, r.Method, r.URL, resp.StatusCode, summary)
			}
			return nil
		},
	}
	cmd.Flags().String("filter", "", "Only replay URLs containing this substring")
	cmd.Flags().Bool("dry-run", false, "List requests without sending them")
	rootCmd.AddCommand(cmd)
}
