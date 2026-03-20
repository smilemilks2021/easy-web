package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Manage named environments (base URLs)",
	}

	envCmd.AddCommand(envListCmd())
	envCmd.AddCommand(envAddCmd())
	envCmd.AddCommand(envUseCmd())
	envCmd.AddCommand(envShowCmd())
	envCmd.AddCommand(envRmCmd())

	rootCmd.AddCommand(envCmd)
}

// envConfigPath returns the path to ~/.easy-web.yaml.
func envConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".easy-web.yaml"), nil
}

// envLoad reads environments and current_env from the config via viper.
func envLoad() (map[string]string, string) {
	envs := viper.GetStringMapString("environments")
	current := viper.GetString("current_env")
	if envs == nil {
		envs = map[string]string{}
	}
	return envs, current
}

// envSave writes updated environments and current_env back to ~/.easy-web.yaml
// using viper.WriteConfigAs so that all existing settings are preserved.
func envSave(envs map[string]string, current string) error {
	viper.Set("environments", envs)
	viper.Set("current_env", current)

	cfgPath, err := envConfigPath()
	if err != nil {
		return err
	}
	return viper.WriteConfigAs(cfgPath)
}

func envListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all environments",
		RunE: func(cmd *cobra.Command, args []string) error {
			envs, current := envLoad()
			if len(envs) == 0 {
				fmt.Println("No environments configured.")
				return nil
			}
			names := make([]string, 0, len(envs))
			for n := range envs {
				names = append(names, n)
			}
			sort.Strings(names)
			for _, n := range names {
				marker := "  "
				if n == current {
					marker = "* "
				}
				fmt.Printf("%s%s  %s\n", marker, n, envs[n])
			}
			return nil
		},
	}
}

func envAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Add or update a named environment",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, url := args[0], args[1]
			envs, current := envLoad()
			envs[name] = url
			if err := envSave(envs, current); err != nil {
				return err
			}
			fmt.Printf("Environment %q added: %s\n", name, url)
			return nil
		},
	}
}

func envUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "Switch the current environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			envs, _ := envLoad()
			if _, ok := envs[name]; !ok {
				return fmt.Errorf("environment %q not found; run 'easy-web env list' to see available environments", name)
			}
			if err := envSave(envs, name); err != nil {
				return err
			}
			fmt.Printf("Switched to environment %q (%s)\n", name, envs[name])
			return nil
		},
	}
}

func envShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show the current environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			envs, current := envLoad()
			if current == "" {
				fmt.Println("No current environment set.")
				return nil
			}
			url, ok := envs[current]
			if !ok {
				fmt.Printf("Current environment: %q (URL not found)\n", current)
				return nil
			}
			fmt.Printf("Current environment: %s  %s\n", current, url)
			return nil
		},
	}
}

func envRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove an environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			envs, current := envLoad()
			if _, ok := envs[name]; !ok {
				return fmt.Errorf("environment %q not found", name)
			}
			delete(envs, name)
			if current == name {
				current = ""
			}
			if err := envSave(envs, current); err != nil {
				return err
			}
			fmt.Printf("Environment %q removed.\n", name)
			return nil
		},
	}
}
