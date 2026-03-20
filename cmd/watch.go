package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/smilemilks2021/easy-web/internal/auth"
	"github.com/smilemilks2021/easy-web/internal/config"
	"github.com/smilemilks2021/easy-web/internal/cookie"
	"github.com/smilemilks2021/easy-web/internal/request"
)

func init() {
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Poll a URL at regular intervals with authentication",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetURL, _ := cmd.Flags().GetString("url")
			if targetURL == "" {
				return fmt.Errorf("--url is required")
			}
			interval, _ := cmd.Flags().GetDuration("interval")
			diff, _ := cmd.Flags().GetBool("diff")

			store := cookie.NewCache(config.CacheDir())
			entries, err := store.Load(parseHost(targetURL))
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

			client := request.NewClient(entries, nil)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigCh
				fmt.Fprintln(os.Stderr, "\nStopping watch...")
				cancel()
			}()

			var lastBody string
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			// Poll immediately on start, then on each tick.
			poll := func() {
				ts := time.Now().Format(time.RFC3339)
				resp, err := client.Do("GET", targetURL, "", nil)
				if err != nil {
					fmt.Printf("[%s] ERROR: %v\n", ts, err)
					return
				}
				defer resp.Body.Close()
				raw, _ := io.ReadAll(resp.Body)
				body := string(raw)

				if diff {
					if body == lastBody {
						// No change — skip output.
						return
					}
					fmt.Printf("[%s] HTTP %d (changed)\n", ts, resp.StatusCode)
					printDiff(lastBody, body)
				} else {
					fmt.Printf("[%s] HTTP %d\n%s\n", ts, resp.StatusCode, body)
				}
				lastBody = body
			}

			poll()
			for {
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
					poll()
				}
			}
		},
	}
	cmd.Flags().DurationP("interval", "i", 10*time.Second, "Poll interval (e.g. 5s, 1m)")
	cmd.Flags().Bool("diff", false, "Only print output when response changes")
	rootCmd.AddCommand(cmd)
}

// printDiff performs a simple line-by-line diff between old and new text,
// printing lines prefixed with "- " (removed) and "+ " (added).
func printDiff(oldText, newText string) {
	oldLines := splitLines(oldText)
	newLines := splitLines(newText)

	oldSet := make(map[string]bool, len(oldLines))
	newSet := make(map[string]bool, len(newLines))
	for _, l := range oldLines {
		oldSet[l] = true
	}
	for _, l := range newLines {
		newSet[l] = true
	}

	for _, l := range oldLines {
		if !newSet[l] {
			fmt.Println("- " + l)
		}
	}
	for _, l := range newLines {
		if !oldSet[l] {
			fmt.Println("+ " + l)
		}
	}
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimRight(s, "\n"), "\n")
}
