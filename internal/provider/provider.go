package provider

import (
	"encoding/json"
	"fmt"
	"os"
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
	model   string
	envKey  string
	modelEnv string
	endpoint string
}{
	ProviderAnthropic:   {model: "claude-opus-4-6", envKey: "ANTHROPIC_API_KEY", modelEnv: "ANTHROPIC_MODEL", endpoint: ""},
	ProviderOpenAI:      {model: "gpt-5", envKey: "OPENAI_API_KEY", modelEnv: "OPENAI_MODEL", endpoint: ""},
	ProviderMinimax:     {model: "minimax-M2.5", envKey: "MINIMAX_API_KEY", modelEnv: "MINIMAX_MODEL", endpoint: ""},
	ProviderGitHubCopilot: {model: "gpt-5-mini", envKey: "OPENAI_API_KEY", modelEnv: "OPENAI_MODEL", endpoint: ""},
	ProviderAzure:       {model: "gpt-4", envKey: "AZURE_OPENAI_API_KEY", modelEnv: "AZURE_OPENAI_MODEL", endpoint: "https://{resource}.openai.azure.com"},
	ProviderVertex:      {model: "gemini-2.0-flash", envKey: "GOOGLE_APPLICATION_CREDENTIALS", modelEnv: "VERTEX_MODEL", endpoint: ""},
}

// DetectProvider attempts to detect the provider from API key format
func DetectProvider(apiKey string) ProviderType {
	if apiKey == "" {
		return ProviderCustom
	}

	// Anthropic keys start with "sk-ant-"
	if strings.HasPrefix(apiKey, "sk-ant-") {
		return ProviderAnthropic
	}

	// Azure keys are typically very long (80+ chars) with mixed characters
	if len(apiKey) >= 80 {
		return ProviderAzure
	}

	// OpenAI keys are typically "sk-" followed by 20+ chars
	if strings.HasPrefix(apiKey, "sk-") && len(apiKey) >= 20 {
		return ProviderOpenAI
	}

	// Minimax keys typically don't have standard prefixes, check length and pattern
	if len(apiKey) > 20 && !strings.HasPrefix(apiKey, "sk-") {
		return ProviderMinimax
	}

	// Azure keys can be various formats, often longer
	if len(apiKey) > 40 {
		return ProviderAzure
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

// ExportVars returns the export commands for a provider
func ExportVars(p ProviderType, apiKey, model, endpoint string) []string {
	envKey := GetEnvKey(p)
	modelEnv := GetModelEnvKey(p)

	lines := []string{
		fmt.Sprintf("export %s=%s", envKey, apiKey),
	}

	if model != "" {
		lines = append(lines, fmt.Sprintf("export %s=%s", modelEnv, model))
	}

	if endpoint != "" {
		lines = append(lines, fmt.Sprintf("export %s_ENDPOINT=%s", strings.ToUpper(string(p)), endpoint))
	}

	return lines
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
