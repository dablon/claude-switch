package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type ProviderType string

const (
	ProviderAnthropic   ProviderType = "anthropic"
	ProviderOpenAI      ProviderType = "openai"
	ProviderMinimax     ProviderType = "minimax"
	ProviderGitHubCopilot ProviderType = "github-copilot"
	ProviderAzure       ProviderType = "azure"
	ProviderVertex      ProviderType = "vertex"
	ProviderCustom      ProviderType = "custom"
)

var providerDefaults = map[ProviderType]struct {
	model    string
	envKey   string
	modelEnv string
	endpoint string
}{
	ProviderAnthropic:     {model: "claude-opus-4-6", envKey: "ANTHROPIC_API_KEY", modelEnv: "ANTHROPIC_MODEL", endpoint: ""},
	ProviderOpenAI:        {model: "gpt-5", envKey: "OPENAI_API_KEY", modelEnv: "OPENAI_MODEL", endpoint: ""},
	ProviderMinimax:       {model: "MiniMax-M2.5", envKey: "MINIMAX_API_KEY", modelEnv: "MINIMAX_MODEL", endpoint: "https://api.minimax.io/anthropic"},
	ProviderGitHubCopilot: {model: "gpt-5-mini", envKey: "OPENAI_API_KEY", modelEnv: "OPENAI_MODEL", endpoint: ""},
	ProviderAzure:         {model: "gpt-4", envKey: "AZURE_OPENAI_API_KEY", modelEnv: "AZURE_OPENAI_MODEL", endpoint: "https://{resource}.openai.azure.com"},
	ProviderVertex:        {model: "gemini-2.0-flash", envKey: "GOOGLE_APPLICATION_CREDENTIALS", modelEnv: "VERTEX_MODEL", endpoint: ""},
}

// claudeVarsForProvider returns the ANTHROPIC_* vars Claude Code needs for a given provider.
func claudeVarsForProvider(p ProviderType, apiKey, model, endpoint string) map[string]string {
	vars := make(map[string]string)

	// Use profile endpoint or fall back to provider default
	if endpoint == "" {
		endpoint = GetEndpoint(p)
	}

	switch p {
	case ProviderMinimax:
		// MiniMax uses ANTHROPIC_AUTH_TOKEN and needs all model aliases set
		vars["ANTHROPIC_AUTH_TOKEN"] = apiKey
		vars["ANTHROPIC_BASE_URL"] = endpoint
		vars["ANTHROPIC_MODEL"] = model
		vars["ANTHROPIC_SMALL_FAST_MODEL"] = model
		vars["ANTHROPIC_DEFAULT_SONNET_MODEL"] = model
		vars["ANTHROPIC_DEFAULT_OPUS_MODEL"] = model
		vars["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = model
		vars["API_TIMEOUT_MS"] = "3000000"
		// Clear the standard key so Claude Code doesn't try to auth with it
		vars["ANTHROPIC_API_KEY"] = ""
	default:
		// Standard Anthropic-compatible providers
		vars["ANTHROPIC_API_KEY"] = apiKey
		if model != "" {
			vars["ANTHROPIC_MODEL"] = model
		}
		if endpoint != "" {
			vars["ANTHROPIC_BASE_URL"] = endpoint
		}
		// Clear MiniMax-specific vars
		vars["ANTHROPIC_AUTH_TOKEN"] = ""
		vars["ANTHROPIC_SMALL_FAST_MODEL"] = ""
		vars["ANTHROPIC_DEFAULT_SONNET_MODEL"] = ""
		vars["ANTHROPIC_DEFAULT_OPUS_MODEL"] = ""
		vars["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = ""
		vars["API_TIMEOUT_MS"] = ""
	}

	return vars
}

// DetectProvider attempts to detect the provider from the profile name
// by matching against known provider names and their common aliases.
func DetectProvider(name string) ProviderType {
	lower := strings.ToLower(name)

	for _, p := range AllProviders() {
		if strings.Contains(lower, string(p)) {
			return p
		}
	}

	return ProviderCustom
}

// GetDefaultModel returns the default model for a provider
func GetDefaultModel(p ProviderType) string {
	if def, ok := providerDefaults[p]; ok {
		return def.model
	}
	return "gpt-5"
}

// GetEnvKey returns the environment variable name for API key
func GetEnvKey(p ProviderType) string {
	if def, ok := providerDefaults[p]; ok {
		return def.envKey
	}
	return fmt.Sprintf("%s_API_KEY", strings.ToUpper(string(p)))
}

// GetModelEnvKey returns the environment variable name for model
func GetModelEnvKey(p ProviderType) string {
	if def, ok := providerDefaults[p]; ok {
		return def.modelEnv
	}
	return fmt.Sprintf("%s_MODEL", strings.ToUpper(string(p)))
}

// GetEndpoint returns the default endpoint for a provider
func GetEndpoint(p ProviderType) string {
	if def, ok := providerDefaults[p]; ok {
		return def.endpoint
	}
	return ""
}

// DetectFromEnv tries to detect provider from existing environment
func DetectFromEnv() ProviderType {
	envCheck := []string{
		"ANTHROPIC_API_KEY",
		"OPENAI_API_KEY",
		"MINIMAX_API_KEY",
		"AZURE_OPENAI_API_KEY",
		"GOOGLE_APPLICATION_CREDENTIALS",
	}

	for _, env := range envCheck {
		if _, exists := os.LookupEnv(env); exists {
			switch env {
			case "ANTHROPIC_API_KEY":
				return ProviderAnthropic
			case "OPENAI_API_KEY":
				return ProviderOpenAI
			case "MINIMAX_API_KEY":
				return ProviderMinimax
			case "AZURE_OPENAI_API_KEY":
				return ProviderAzure
			case "GOOGLE_APPLICATION_CREDENTIALS":
				return ProviderVertex
			}
		}
	}

	return ProviderCustom
}

// ExportVars returns bash export commands for Claude Code to use any provider.
func ExportVars(p ProviderType, apiKey, model, endpoint string) []string {
	vars := claudeVarsForProvider(p, apiKey, model, endpoint)
	var lines []string
	for k, v := range vars {
		if v != "" {
			lines = append(lines, fmt.Sprintf("export %s=%s", k, v))
		} else {
			lines = append(lines, fmt.Sprintf("unset %s", k))
		}
	}
	return lines
}

// ExportVarsPowerShell returns PowerShell $env: commands for Claude Code.
func ExportVarsPowerShell(p ProviderType, apiKey, model, endpoint string) []string {
	vars := claudeVarsForProvider(p, apiKey, model, endpoint)
	var lines []string
	for k, v := range vars {
		if v != "" {
			lines = append(lines, fmt.Sprintf("$env:%s=\"%s\"", k, v))
		} else {
			lines = append(lines, fmt.Sprintf("Remove-Item Env:%s -ErrorAction SilentlyContinue", k))
		}
	}
	return lines
}

// ApplyEnvVars sets ANTHROPIC_* environment variables at the system/user level
// so they persist across shell sessions and are picked up by Claude Code.
func ApplyEnvVars(p ProviderType, apiKey, model, endpoint string) error {
	vars := claudeVarsForProvider(p, apiKey, model, endpoint)
	for key, val := range vars {
		if err := setSystemEnv(key, val); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}
	return nil
}

// setSystemEnv sets or clears a persistent user-level environment variable.
// Empty value removes the var.
func setSystemEnv(key, value string) error {
	if runtime.GOOS == "windows" {
		if value == "" {
			// Delete from user env via registry
			cmd := exec.Command("reg", "delete", `HKCU\Environment`, "/v", key, "/f")
			cmd.Run() // ignore error if key doesn't exist
			return nil
		}
		cmd := exec.Command("setx", key, value)
		return cmd.Run()
	}
	return appendToEnvFile(key, value)
}

// appendToEnvFile writes env vars to ~/.claude-switch/env that can be sourced.
func appendToEnvFile(key, value string) error {
	homeDir, _ := os.UserHomeDir()
	envFile := homeDir + "/.claude-switch/env"

	// Read existing content to replace if key exists
	content, _ := os.ReadFile(envFile)
	lines := strings.Split(string(content), "\n")
	found := false
	prefix := "export " + key + "="
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			lines[i] = prefix + value
			found = true
		}
	}
	if !found {
		lines = append(lines, prefix+value)
	}

	// Remove empty lines
	var clean []string
	for _, l := range lines {
		if l != "" {
			clean = append(clean, l)
		}
	}

	return os.WriteFile(envFile, []byte(strings.Join(clean, "\n")+"\n"), 0600)
}

// AllProviders returns all supported provider types
func AllProviders() []ProviderType {
	return []ProviderType{
		ProviderAnthropic,
		ProviderOpenAI,
		ProviderMinimax,
		ProviderGitHubCopilot,
		ProviderAzure,
		ProviderVertex,
	}
}

// String returns the string representation
func (p ProviderType) String() string {
	return string(p)
}

// UnmarshalJSON implements custom JSON unmarshaling
func (p *ProviderType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*p = ProviderType(s)
	return nil
}

// MarshalJSON implements custom JSON marshaling
func (p ProviderType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(p))
}
