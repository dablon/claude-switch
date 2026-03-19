package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
				Usage: "Quick add profile - auto-detects provider and API key from env",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Required: true, Usage: "Profile name (must contain provider: anthropic, openai, minimax, etc.)"},
					&cli.StringFlag{Name: "key", Aliases: []string{"k"}, Usage: "API key (auto-read from env if omitted)"},
				},
				Action: quickAdd,
			},
			{
				Name:   "list",
				Usage:  "List all profiles",
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
				Name:   "show",
				Usage:  "Show current profile",
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
				Name:   "export",
				Usage:  "Export current profile as environment variables",
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
				Name:   "providers",
				Usage:  "List supported providers",
				Action: listProviders,
			},
			{
				Name:  "edit",
				Usage: "Edit an existing profile",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Required: true, Usage: "Profile name to edit"},
					&cli.StringFlag{Name: "model", Aliases: []string{"m"}, Usage: "New model name"},
					&cli.StringFlag{Name: "provider", Aliases: []string{"p"}, Usage: "New provider"},
					&cli.StringFlag{Name: "key", Aliases: []string{"k"}, Usage: "New API key"},
					&cli.StringFlag{Name: "endpoint", Aliases: []string{"e"}, Usage: "New endpoint"},
				},
				Action: editProfile,
			},
			{
				Name:   "test",
				Usage:  "Test the current profile API key",
				Action: testProfile,
			},
			{
				Name:  "chat",
				Usage: "Chat with the current profile",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "message", Aliases: []string{"m"}, Usage: "Single message (or leave empty for interactive mode)"},
				},
				Action: chatProfile,
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

	// Auto-detect provider from profile name
	detected := provider.DetectProvider(name)
	if detected == provider.ProviderCustom {
		return fmt.Errorf("could not detect provider from name %q. Use a name containing a provider (anthropic, openai, minimax, azure, vertex) or use 'add' with --provider", name)
	}

	// If no key provided, read from the provider's env var
	if apiKey == "" {
		envKey := provider.GetEnvKey(detected)
		apiKey = os.Getenv(envKey)
		if apiKey == "" {
			return fmt.Errorf("no --key provided and %s is not set in environment", envKey)
		}
		color.Cyan("  Using API key from %s", envKey)
	}

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
	color.Cyan("  Provider: %s | Model: %s", p.Provider, p.Model)

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

func editProfile(c *cli.Context) error {
	name := c.String("name")
	model := c.String("model")
	providerStr := c.String("provider")
	apiKey := c.String("key")
	endpoint := c.String("endpoint")

	// Check that at least one field is being updated
	if model == "" && providerStr == "" && apiKey == "" && endpoint == "" {
		color.Yellow("No changes specified. Use --model, --provider, --key, or --endpoint")
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	profile := config.GetProfile(cfg, name)
	if profile == nil {
		return fmt.Errorf("profile not found: %s", name)
	}

	// Update fields if provided
	if model != "" {
		profile.Model = model
	}
	if apiKey != "" {
		profile.APIKey = apiKey
	}
	if endpoint != "" {
		profile.Endpoint = endpoint
	}
	if providerStr != "" {
		profile.Provider = provider.ProviderType(providerStr)
		// If model wasn't explicitly set, update it to provider default
		if model == "" {
			profile.Model = provider.GetDefaultModel(profile.Provider)
		}
	}

	// Save the updated profile
	if err := config.AddProfile(cfg, *profile); err != nil {
		return err
	}

	if err := config.Save(cfg); err != nil {
		return err
	}

	color.Green("✓ Updated profile: %s", name)
	color.Cyan("  Provider: %s | Model: %s", profile.Provider, profile.Model)
	if profile.Endpoint != "" {
		fmt.Printf("  Endpoint: %s\n", profile.Endpoint)
	}

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

	profile := config.GetCurrentProfile(cfg)

	// Set system-level env vars automatically
	if err := config.ApplyEnvVars(profile); err != nil {
		return fmt.Errorf("failed to apply env vars: %w", err)
	}

	color.Green("✓ Switched to: %s (%s | %s)", name, profile.Provider, profile.Model)
	color.Green("✓ System env vars applied — restart Claude Code to pick up the changes.")

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
	name := c.String("key")
	detected := provider.DetectProvider(name)

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

func testProfile(c *cli.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	current := config.GetCurrentProfile(cfg)
	if current == nil {
		color.Yellow("No profile selected. Use 'claude-switch use --name <profile>'")
		return nil
	}

	color.Cyan("Testing profile: %s", current.Name)
	color.Cyan("Provider: %s | Model: %s", current.Provider, current.Model)
	fmt.Println()

	client := provider.NewClient(current.APIKey, current.Provider, current.Model, current.Endpoint)

	if err := client.Test(); err != nil {
		color.Red("✗ Test failed: %v", err)
		os.Exit(1)
	}

	color.Green("✓ API key works!")
	return nil
}

func chatProfile(c *cli.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	current := config.GetCurrentProfile(cfg)
	if current == nil {
		color.Yellow("No profile selected. Use 'claude-switch use --name <profile>'")
		return nil
	}

	color.Cyan("Chatting with: %s (%s - %s)", current.Name, current.Provider, current.Model)
	color.Yellow("Type 'exit' or 'quit' to end the conversation")
	color.Yellow("Type 'clear' to clear conversation history")
	fmt.Println()

	client := provider.NewClient(current.APIKey, current.Provider, current.Model, current.Endpoint)
	messages := []provider.Message{}

	msg := c.String("message")
	if msg != "" {
		messages = append(messages, provider.Message{Role: "user", Content: msg})
		response, err := client.Chat(messages)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
		fmt.Println(provider.FormatMessage("assistant", response))
		return nil
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\033[32mYou:\033[0m ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if strings.ToLower(input) == "exit" || strings.ToLower(input) == "quit" {
			color.Cyan("Goodbye!")
			break
		}
		if strings.ToLower(input) == "clear" {
			messages = []provider.Message{}
			color.Cyan("Conversation cleared")
			continue
		}

		messages = append(messages, provider.Message{Role: "user", Content: input})

		response, err := client.Chat(messages)
		if err != nil {
			color.Red("Error: %v", err)
			continue
		}

		fmt.Println(provider.FormatMessage("assistant", response))
	}

	return nil
}
