package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/smilemilks2021/easy-web/internal/browser"
	"github.com/smilemilks2021/easy-web/internal/config"
	"github.com/smilemilks2021/easy-web/internal/skill"
)

func init() {
	skillCmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage generated Claude Code Skills",
	}

	genCmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate or update a Claude Code Skill from captured APIs",
		Long: `Regenerate a Claude Code Skill for the given URL by re-capturing its APIs
and writing (or smart-merging) ~/.claude/skills/<name>/SKILL.md.

Example:
  easy-web skill gen -u https://s-power.sheincorp.cn
  easy-web skill gen -u https://grafana.internal.com --skill-name grafana`,
		RunE: func(cmd *cobra.Command, args []string) error {
			targetURL, _ := cmd.Flags().GetString("url")
			if targetURL == "" {
				return fmt.Errorf("--url is required")
			}
			skillName, _ := cmd.Flags().GetString("skill-name")
			patterns, _ := cmd.Flags().GetStringArray("pattern")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			// If --from-file is set, read a request list JSON instead of capturing
			fromFile, _ := cmd.Flags().GetString("from-file")
			if fromFile != "" {
				return genFromFile(fromFile, targetURL, skillName)
			}

			fmt.Printf("Capturing APIs from %s to generate skill...\n", targetURL)
			reqs, err := browser.CaptureRequests(targetURL, browser.CaptureOptions{
				Patterns:     patterns,
				Timeout:      timeout,
				ReuseProfile: true,
				ProfileDir:   config.ProfileDir(),
			})
			if err != nil {
				return err
			}

			if len(reqs) == 0 {
				return fmt.Errorf("no API requests captured — check --pattern filter or increase --timeout")
			}

			return skill.Generate(reqs, targetURL, skillName)
		},
	}
	genCmd.Flags().StringP("url", "u", "", "Target URL to capture APIs from")
	genCmd.Flags().String("skill-name", "", "Override skill name (default: derived from domain)")
	genCmd.Flags().StringArrayP("pattern", "p", nil, "URL filter pattern (OR)")
	genCmd.Flags().Duration("timeout", 0, "Capture timeout (0 = use last captured data)")
	genCmd.Flags().String("from-file", "", "Generate from a saved request list JSON file")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List generated Claude Code Skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := os.UserHomeDir()
			skillsDir := home + "/.claude/skills"
			entries, err := os.ReadDir(skillsDir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("No skills generated yet. Run: easy-web capture -u <URL>")
					return nil
				}
				return err
			}
			fmt.Printf("Generated skills in %s:\n", skillsDir)
			for _, e := range entries {
				if e.IsDir() {
					skillFile := skillsDir + "/" + e.Name() + "/SKILL.md"
					if _, err := os.Stat(skillFile); err == nil {
						fmt.Printf("  %s → %s\n", e.Name(), skillFile)
					}
				}
			}
			return nil
		},
	}

	skillCmd.AddCommand(genCmd)
	skillCmd.AddCommand(listCmd)
	rootCmd.AddCommand(skillCmd)
}

// genFromFile generates a skill from a previously saved request list (future use).
func genFromFile(path, baseURL, skillName string) error {
	return fmt.Errorf("--from-file not yet implemented (path: %s)", path)
}
