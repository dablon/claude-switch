package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

type Profile struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	APIKey   string `json:"api_key,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

type Config struct {
	Profiles       []Profile `json:"profiles"`
	CurrentProfile string    `json:"current_profile"`
}

var (
	configDir  = filepath.Join(os.Getenv("HOME"), ".claude-switch")
	configFile = filepath.Join(configDir, "config.json")
)

func main() {
	app := &cli.App{
		Name:  "claude-switch",
		Usage: "Switch between Claude Code profiles easily",
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a new profile",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Required: true, Usage: "Profile name"},
					&cli.StringFlag{Name: "provider", Required: true, Usage: "Provider (openai, anthropic, minimax, etc.)"},
					&cli.StringFlag{Name: "model", Required: true, Usage: "Model name"},
					&cli.StringFlag{Name: "key", Required: true, Usage: "API key"},
					&cli.StringFlag{Name: "endpoint", Usage: "Custom endpoint (optional)"},
				},
				Action: addProfile,
			},
			{
				Name:  "list",
				Usage: "List all profiles",
				Action: listProfiles,
			},
			{
				Name:  "use",
				Usage: "Switch to a profile",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Required: true, Usage: "Profile name to switch to"},
				},
				Action: useProfile,
			},
			{
				Name:  "show",
				Usage: "Show current profile",
				Action: showCurrent,
			},
			{
				Name:  "remove",
				Usage: "Remove a profile",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Required: true, Usage: "Profile name to remove"},
				},
				Action: removeProfile,
			},
			{
				Name:  "export",
				Usage: "Export current profile as environment variables",
				Action: exportEnv,
			},
			{
				Name:  "apply",
				Usage: "Apply profile to Claude Code config",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Required: true, Usage: "Profile name to apply"},
				},
				Action: applyProfile,
			},
		},
	}

	app.Run(os.Args)
}

func loadConfig() (*Config, error) {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return &Config{Profiles: []Profile{}}, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func saveConfig(config *Config) error {
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0600)
}

func addProfile(c *cli.Context) error {
	profile := Profile{
		Name:     c.String("name"),
		Provider: c.String("provider"),
		Model:    c.String("model"),
		APIKey:   c.String("key"),
		Endpoint: c.String("endpoint"),
	}

	config, err := loadConfig()
	if err != nil {
		return err
	}

	// Check if profile exists
	for i, p := range config.Profiles {
		if p.Name == profile.Name {
			config.Profiles[i] = profile
			color.Yellow("Updated existing profile: %s", profile.Name)
			return saveConfig(config)
		}
	}

	config.Profiles = append(config.Profiles, profile)

	// If first profile, set as current
	if config.CurrentProfile == "" {
		config.CurrentProfile = profile.Name
	}

	color.Green("Added profile: %s (%s/%s)", profile.Name, profile.Provider, profile.Model)
	return saveConfig(config)
}

func listProfiles(c *cli.Context) error {
	config, err := loadConfig()
	if err != nil {
		return err
	}

	if len(config.Profiles) == 0 {
		color.Yellow("No profiles configured. Use 'claude-switch add' to add one.")
		return nil
	}

	color.Cyan("📋 Configured Profiles:\n")

	// Sort profiles
	sort.Slice(config.Profiles, func(i, j int) bool {
		return config.Profiles[i].Name < config.Profiles[j].Name
	})

	for _, p := range config.Profiles {
		if p.Name == config.CurrentProfile {
			color.Green("  ✅ %s", p.Name)
		} else {
			fmt.Printf("    %s", p.Name)
		}
		fmt.Printf("     Provider: %s | Model: %s\n", p.Provider, p.Model)
	}

	fmt.Println()
	color.Cyan("💡 Use 'claude-switch use --name <profile>' to switch")

	return nil
}

func useProfile(c *cli.Context) error {
	name := c.String("name")

	config, err := loadConfig()
	if err != nil {
		return err
	}

	found := false
	for _, p := range config.Profiles {
		if p.Name == name {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("profile not found: %s", name)
	}

	config.CurrentProfile = name
	color.Green("Switched to profile: %s", name)
	return saveConfig(config)
}

func showCurrent(c *cli.Context) error {
	config, err := loadConfig()
	if err != nil {
		return err
	}

	if config.CurrentProfile == "" {
		color.Yellow("No profile selected. Use 'claude-switch use --name <profile>'")
		return nil
	}

	for _, p := range config.Profiles {
		if p.Name == config.CurrentProfile {
			color.Green("👤 Current Profile: %s", p.Name)
			fmt.Printf("   Provider: %s\n", p.Provider)
			fmt.Printf("   Model:    %s\n", p.Model)
			if p.Endpoint != "" {
				fmt.Printf("   Endpoint: %s\n", p.Endpoint)
			}
			return nil
		}
	}

	return nil
}

func removeProfile(c *cli.Context) error {
	name := c.String("name")

	config, err := loadConfig()
	if err != nil {
		return err
	}

	newProfiles := []Profile{}
	found := false
	for _, p := range config.Profiles {
		if p.Name != name {
			newProfiles = append(newProfiles, p)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("profile not found: %s", name)
	}

	config.Profiles = newProfiles

	// Reset current if removed
	if config.CurrentProfile == name {
		if len(config.Profiles) > 0 {
			config.CurrentProfile = config.Profiles[0].Name
		} else {
			config.CurrentProfile = ""
		}
	}

	color.Red("Removed profile: %s", name)
	return saveConfig(config)
}

func exportEnv(c *cli.Context) error {
	config, err := loadConfig()
	if err != nil {
		return err
	}

	if config.CurrentProfile == "" {
		return fmt.Errorf("no profile selected")
	}

	for _, p := range config.Profiles {
		if p.Name == config.CurrentProfile {
			switch p.Provider {
			case "openai", "github-copilot":
				fmt.Printf("export OPENAI_API_KEY=%s\n", p.APIKey)
				if p.Model != "" {
					fmt.Printf("export OPENAI_MODEL=%s\n", p.Model)
				}
			case "anthropic", "claude":
				fmt.Printf("export ANTHROPIC_API_KEY=%s\n", p.APIKey)
				if p.Model != "" {
					fmt.Printf("export ANTHROPIC_MODEL=%s\n", p.Model)
				}
			case "minimax":
				fmt.Printf("export MINIMAX_API_KEY=%s\n", p.APIKey)
				if p.Model != "" {
					fmt.Printf("export MINIMAX_MODEL=%s\n", p.Model)
				}
			default:
				// Generic export
				fmt.Printf("export %s_API_KEY=%s\n", strings.ToUpper(p.Provider), p.APIKey)
				if p.Model != "" {
					fmt.Printf("export %s_MODEL=%s\n", strings.ToUpper(p.Provider), p.Model)
				}
			}

			if p.Endpoint != "" {
				fmt.Printf("export %s_ENDPOINT=%s\n", strings.ToUpper(p.Provider), p.Endpoint)
			}
			return nil
		}
	}

	return fmt.Errorf("profile not found")
}

func applyProfile(c *cli.Context) error {
	name := c.String("name")

	config, err := loadConfig()
	if err != nil {
		return err
	}

	var profile *Profile
	for i := range config.Profiles {
		if config.Profiles[i].Name == name {
			profile = &config.Profiles[i]
			break
		}
	}

	if profile == nil {
		return fmt.Errorf("profile not found: %s", name)
	}

	config.CurrentProfile = name
	saveConfig(config)

	color.Green("✅ Switched to profile: %s", name)
	fmt.Println()
	color.Cyan("📝 Add these to your environment or Claude Code config:")
	fmt.Println()

	// Determine how to apply based on provider
	switch profile.Provider {
	case "openai", "github-copilot":
		color.Yellow("# For OpenAI/Claude Code:")
		fmt.Printf("  OPENAI_API_KEY=%s\n", profile.APIKey)
		if profile.Model != "" {
			fmt.Printf("  OPENAI_MODEL=%s\n", profile.Model)
		}
	case "anthropic", "claude":
		color.Yellow("# For Anthropic Claude:")
		fmt.Printf("  ANTHROPIC_API_KEY=%s\n", profile.APIKey)
		if profile.Model != "" {
			fmt.Printf("  ANTHROPIC_MODEL=%s\n", profile.Model)
		}
	case "minimax":
		color.Yellow("# For Minimax:")
		fmt.Printf("  MINIMAX_API_KEY=%s\n", profile.APIKey)
		if profile.Model != "" {
			fmt.Printf("  MINIMAX_MODEL=%s\n", profile.Model)
		}
	default:
		color.Yellow("# For %s:", profile.Provider)
		fmt.Printf("  %s_API_KEY=%s\n", strings.ToUpper(profile.Provider), profile.APIKey)
		if profile.Model != "" {
			fmt.Printf("  %s_MODEL=%s\n", strings.ToUpper(profile.Provider), profile.Model)
		}
	}

	fmt.Println()
	color.Cyan("💡 Quick apply command:")
	fmt.Printf("  eval $(claude-switch export)\n")

	return nil
}
