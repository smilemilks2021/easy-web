package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	cfgCmd := &cobra.Command{Use: "config", Short: "Manage configuration"}
	cfgCmd.AddCommand(&cobra.Command{
		Use: "edit", Short: "Open config in editor",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := os.UserHomeDir()
			cfgFile := filepath.Join(home, ".easy-web.yaml")
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = os.Getenv("VISUAL")
			}
			if editor == "" {
				if runtime.GOOS == "windows" {
					editor = "notepad.exe"
				} else {
					editor = "vi"
				}
			}
			c := exec.Command(editor, cfgFile)
			c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("editor %q: %w", editor, err)
			}
			return nil
		},
	})
	rootCmd.AddCommand(cfgCmd)
}
