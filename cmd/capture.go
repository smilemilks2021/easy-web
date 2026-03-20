package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/smilemilks2021/easy-web/internal/browser"
	"github.com/smilemilks2021/easy-web/internal/config"
	"github.com/smilemilks2021/easy-web/internal/skill"
)

func init() {
	cmd := &cobra.Command{
		Use:   "capture",
		Short: "Record API requests from a website",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetURL, _ := cmd.Flags().GetString("url")
			if targetURL == "" {
				return fmt.Errorf("--url is required")
			}
			patterns, _ := cmd.Flags().GetStringArray("pattern")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			autoSave, _ := cmd.Flags().GetBool("auto-save")
			skillName, _ := cmd.Flags().GetString("skill-name")

			reqs, err := browser.CaptureRequests(targetURL, browser.CaptureOptions{
				Patterns:     patterns,
				Timeout:      timeout,
				ReuseProfile: true,
				ProfileDir:   config.ProfileDir(),
			})
			if err != nil {
				return err
			}

			fmt.Printf("\nCaptured %d requests:\n", len(reqs))
			for i, r := range reqs {
				fmt.Printf("  [%d] %s %s\n", i+1, r.Method, r.URL)
			}
			if autoSave {
				fmt.Printf("Auto-saved %d API configurations.\n", len(reqs))
			}

			// Auto-generate Claude Code Skill after capture
			if len(reqs) > 0 {
				if err := skill.Generate(reqs, targetURL, skillName); err != nil {
					fmt.Printf("[skill] warning: could not generate skill: %v\n", err)
				}
			}

			return nil
		},
	}
	cmd.Flags().StringArrayP("pattern", "p", nil, "URL filter pattern (OR)")
	cmd.Flags().DurationP("timeout", "t", 5*time.Minute, "Capture timeout")
	cmd.Flags().Bool("auto-save", false, "Auto-save without confirmation")
	cmd.Flags().Bool("interactive", false, "Interactive API selection")
	cmd.Flags().String("skill-name", "", "Override generated skill name (default: derived from domain)")
	rootCmd.AddCommand(cmd)
}
