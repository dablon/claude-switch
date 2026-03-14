package main

import (
	"fmt"
	"os"

	"claude-switch/internal/config"
	"claude-switch/internal/provider"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "claude-switch",
		Usage: "Switch between Claude Code profiles easily",
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a new profile (auto-detects provider from API key)",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Required: true, Usage: "Profile name"},
					&cli.StringFlag{Name: "key", Aliases: []string{"k"}, Required: true, Usage: "API key"},
					&cli.StringFlag{Name: "model", Aliases: []string{"m"}, Usage: "Model name (auto-detected if omitted)"},
					&cli.StringFlag{Name: "provider", Aliases: []string{"p"}, Usage: "Provider (auto-detected from key if omitted)"},
					&cli.StringFlag{Name: "endpoint", Aliases: []string{"e"}, Usage: "Custom endpoint (optional)"},
				},
				Action: addProfile,
			},
			{
				Name:  "quick",
				Usage: "Quick add profile - auto-detects provider from API key",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Required: true, Usage: "Profile name"},
					&cli.StringFlag{Name: "key", Aliases: []string{"k"}, Required: true, Usage: "API key"},
				},
				Action: quickAdd,
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
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Required: true, Usage: "Profile name"},
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
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Required: true, Usage: "Profile name"},
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
				Usage: "Apply and show profile configuration",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Usage: "Profile name (uses current if omitted)"},
				},
				Action: applyProfile,
			},
			{
				Name:  "detect",
				Usage: "Detect provider from API key",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "key", Aliases: []string{"k"}, Required: true, Usage: "API key to detect"},
				},
				Action: detectProvider,
			},
			{
				Name:  "providers",
				Usage: "List supported providers",
				Action: listProviders,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
}

func addProfile(c *cli.Context) error {
	name := c.String("name")
	apiKey := c.String("key")
	model := c.String("model")
	providerStr := c.String("provider")
	endpoint := c.String("endpoint")

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var p config.Profile

	if providerStr != "" {
		// Use explicitly specified provider
		p = config.Profile{
			Name:     name,
			Provider: provider.ProviderType(providerStr),
			Model:    model,
			APIKey:   apiKey,
			Endpoint: endpoint,
		}
		// Use default model if not provided
		if p.Model == "" {
			p.Model = provider.GetDefaultModel(p.Provider)
		}
	} else {
		// Auto-detect provider
		p = config.DetectAndCreateProfile(name, apiKey, model, endpoint)
	}

	if err := config.AddProfile(cfg, p); err != nil {
		return err
	}

	if err := config.Save(cfg); err != nil {
		return err
	}

	color.Green("✓ Added profile: %s", p.Name)
	color.Cyan("  Provider: %s | Model: %s", p.Provider, p.Model)

	return nil
}

func quickAdd(c *cli.Context) error {
	name := c.String("name")
	apiKey := c.String("key")

	// Auto-detect provider
	detected := provider.DetectProvider(apiKey)
	model := provider.GetDefaultModel(detected)

	p := config.Profile{
		Name:     name,
		Provider: detected,
		Model:    model,
		APIKey:   apiKey,
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := config.AddProfile(cfg, p); err != nil {
		return err
	}

	if err := config.Save(cfg); err != nil {
		return err
	}

	color.Green("✓ Quick added profile: %s", p.Name)
	color.Cyan("  Auto-detected: %s with model %s", p.Provider, p.Model)

	return nil
}

func listProfiles(c *cli.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Profiles) == 0 {
		color.Yellow("No profiles. Use 'claude-switch quick --name myprofile --key YOUR_KEY' to add one.")
		return nil
	}

	color.Cyan("📋 Profiles:")
	fmt.Println()

	for _, p := range config.SortedProfiles(cfg) {
		if p.Name == cfg.CurrentProfile {
			color.Green("  ✅ %s", p.Name)
		} else {
			color.Cyan("    %s", p.Name)
		}
		fmt.Printf("     %s | %s\n", p.Provider, p.Model)
	}

	fmt.Println()
	color.Cyan("💡 Use: claude-switch use --name <profile>")

	return nil
}

func useProfile(c *cli.Context) error {
	name := c.String("name")

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := config.SetCurrent(cfg, name); err != nil {
		return err
	}

	if err := config.Save(cfg); err != nil {
		return err
	}

	color.Green("✓ Switched to: %s", name)

	return nil
}

func showCurrent(c *cli.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	current := config.GetCurrentProfile(cfg)
	if current == nil {
		color.Yellow("No profile selected. Use 'claude-switch use --name <profile>'")
		return nil
	}

	color.Green("👤 Current: %s", current.Name)
	fmt.Printf("   Provider: %s\n", current.Provider)
	fmt.Printf("   Model:    %s\n", current.Model)
	if current.Endpoint != "" {
		fmt.Printf("   Endpoint: %s\n", current.Endpoint)
	}

	return nil
}

func removeProfile(c *cli.Context) error {
	name := c.String("name")

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := config.RemoveProfile(cfg, name); err != nil {
		return err
	}

	if err := config.Save(cfg); err != nil {
		return err
	}

	color.Red("✗ Removed: %s", name)

	return nil
}

func exportEnv(c *cli.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	current := config.GetCurrentProfile(cfg)
	if current == nil {
		return fmt.Errorf("no profile selected")
	}

	for _, line := range config.ExportEnvVars(current) {
		fmt.Println(line)
	}

	return nil
}

func applyProfile(c *cli.Context) error {
	name := c.String("name")

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var current *config.Profile
	if name != "" {
		current = config.GetProfile(cfg, name)
		if current == nil {
			return fmt.Errorf("profile not found: %s", name)
		}
	} else {
		current = config.GetCurrentProfile(cfg)
		if current == nil {
			return fmt.Errorf("no profile selected")
		}
	}

	// Save as current
	cfg.CurrentProfile = current.Name
	config.Save(cfg)

	color.Green("✓ Active: %s", current.Name)
	fmt.Println()
	color.Cyan("📝 Export with:")
	fmt.Println()
	color.Yellow("  eval $(claude-switch export)")
	fmt.Println()

	return nil
}

func detectProvider(c *cli.Context) error {
	apiKey := c.String("key")
	detected := provider.DetectProvider(apiKey)

	color.Green("Detected provider: %s", detected)
	fmt.Printf("Default model: %s\n", provider.GetDefaultModel(detected))
	fmt.Printf("Env key: %s\n", provider.GetEnvKey(detected))

	return nil
}

func listProviders(c *cli.Context) error {
	color.Cyan("Supported providers:")
	fmt.Println()

	for _, p := range provider.AllProviders() {
		color.Green("  • %s", p)
		fmt.Printf("    Model: %s\n", provider.GetDefaultModel(p))
		fmt.Printf("    Env:   %s\n", provider.GetEnvKey(p))
	}

	return nil
}
