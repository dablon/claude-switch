package provider

import (
	"os"
	"testing"
)

func TestDetectProvider_ByName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ProviderType
	}{
		// Exact matches
		{"anthropic", "anthropic", ProviderAnthropic},
		{"openai", "openai", ProviderOpenAI},
		{"minimax", "minimax", ProviderMinimax},
		{"azure", "azure", ProviderAzure},
		{"vertex", "vertex", ProviderVertex},
		{"github-copilot", "github-copilot", ProviderGitHubCopilot},

		// Name contains provider
		{"my-anthropic-key", "my-anthropic-key", ProviderAnthropic},
		{"openai-prod", "openai-prod", ProviderOpenAI},
		{"minimax-test", "minimax-test", ProviderMinimax},
		{"azure-eastus", "azure-eastus", ProviderAzure},

		// Case insensitive
		{"Anthropic", "Anthropic", ProviderAnthropic},
		{"OPENAI", "OPENAI", ProviderOpenAI},
		{"MiniMax", "MiniMax", ProviderMinimax},

		// Unknown names fall back to custom
		{"my-custom-llm", "my-custom-llm", ProviderCustom},
		{"random-name", "random-name", ProviderCustom},
		{"empty", "", ProviderCustom},
		{"claude-pro", "claude-pro", ProviderCustom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectProvider(tt.input)
			if result != tt.expected {
				t.Errorf("DetectProvider(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetDefaultModel(t *testing.T) {
	tests := []struct {
		provider ProviderType
		expected string
	}{
		{ProviderAnthropic, "claude-opus-4-6"},
		{ProviderOpenAI, "gpt-5"},
		{ProviderMinimax, "MiniMax-M2.5"},
		{ProviderGitHubCopilot, "gpt-5-mini"},
		{ProviderCustom, "gpt-5"},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			result := GetDefaultModel(tt.provider)
			if result != tt.expected {
				t.Errorf("GetDefaultModel(%s) = %v, want %v", tt.provider, result, tt.expected)
			}
		})
	}
}

func TestGetEnvKey(t *testing.T) {
	tests := []struct {
		provider ProviderType
		expected string
	}{
		{ProviderAnthropic, "ANTHROPIC_API_KEY"},
		{ProviderOpenAI, "OPENAI_API_KEY"},
		{ProviderMinimax, "MINIMAX_API_KEY"},
		{ProviderAzure, "AZURE_OPENAI_API_KEY"},
		{ProviderCustom, "CUSTOM_API_KEY"},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			result := GetEnvKey(tt.provider)
			if result != tt.expected {
				t.Errorf("GetEnvKey(%s) = %v, want %v", tt.provider, result, tt.expected)
			}
		})
	}
}

func TestGetModelEnvKey(t *testing.T) {
	tests := []struct {
		provider ProviderType
		expected string
	}{
		{ProviderAnthropic, "ANTHROPIC_MODEL"},
		{ProviderOpenAI, "OPENAI_MODEL"},
		{ProviderMinimax, "MINIMAX_MODEL"},
		{ProviderAzure, "AZURE_OPENAI_MODEL"},
		{ProviderCustom, "CUSTOM_MODEL"},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			result := GetModelEnvKey(tt.provider)
			if result != tt.expected {
				t.Errorf("GetModelEnvKey(%s) = %v, want %v", tt.provider, result, tt.expected)
			}
		})
	}
}

func TestGetEndpoint(t *testing.T) {
	tests := []struct {
		provider ProviderType
		expected string
	}{
		{ProviderAzure, "https://{resource}.openai.azure.com"},
		{ProviderAnthropic, ""},
		{ProviderOpenAI, ""},
		{ProviderMinimax, "https://api.minimax.io/anthropic"},
		{ProviderCustom, ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			result := GetEndpoint(tt.provider)
			if result != tt.expected {
				t.Errorf("GetEndpoint(%s) = %v, want %v", tt.provider, result, tt.expected)
			}
		})
	}
}

func TestDetectFromEnv(t *testing.T) {
	// Save original env
	origAnthropic := os.Getenv("ANTHROPIC_API_KEY")
	origOpenAI := os.Getenv("OPENAI_API_KEY")
	origMinimax := os.Getenv("MINIMAX_API_KEY")
	origAzure := os.Getenv("AZURE_OPENAI_API_KEY")
	origGoogle := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	defer func() {
		setOrUnset("ANTHROPIC_API_KEY", origAnthropic)
		setOrUnset("OPENAI_API_KEY", origOpenAI)
		setOrUnset("MINIMAX_API_KEY", origMinimax)
		setOrUnset("AZURE_OPENAI_API_KEY", origAzure)
		setOrUnset("GOOGLE_APPLICATION_CREDENTIALS", origGoogle)
	}()

	// Clear all env vars
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("MINIMAX_API_KEY")
	os.Unsetenv("AZURE_OPENAI_API_KEY")
	os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")

	result := DetectFromEnv()
	if result != ProviderCustom {
		t.Errorf("DetectFromEnv() = %v, want %v with no env vars", result, ProviderCustom)
	}

	// Test with ANTHROPIC_API_KEY
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test")
	result = DetectFromEnv()
	if result != ProviderAnthropic {
		t.Errorf("DetectFromEnv() = %v, want %v with ANTHROPIC_API_KEY", result, ProviderAnthropic)
	}

	// Test with OPENAI_API_KEY
	os.Setenv("OPENAI_API_KEY", "sk-test")
	os.Unsetenv("ANTHROPIC_API_KEY")
	result = DetectFromEnv()
	if result != ProviderOpenAI {
		t.Errorf("DetectFromEnv() = %v, want %v with OPENAI_API_KEY", result, ProviderOpenAI)
	}

	// Test with MINIMAX_API_KEY
	os.Setenv("MINIMAX_API_KEY", "mmx-test")
	os.Unsetenv("OPENAI_API_KEY")
	result = DetectFromEnv()
	if result != ProviderMinimax {
		t.Errorf("DetectFromEnv() = %v, want %v with MINIMAX_API_KEY", result, ProviderMinimax)
	}

	// Test with GOOGLE_APPLICATION_CREDENTIALS
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/path/to/creds.json")
	os.Unsetenv("MINIMAX_API_KEY")
	result = DetectFromEnv()
	if result != ProviderVertex {
		t.Errorf("DetectFromEnv() = %v, want %v with GOOGLE_APPLICATION_CREDENTIALS", result, ProviderVertex)
	}
}

func setOrUnset(env, value string) {
	if value != "" {
		os.Setenv(env, value)
	} else {
		os.Unsetenv(env)
	}
}

func TestExportVars(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderType
		apiKey   string
		model    string
		endpoint string
		expected []string
	}{
		{
			name:     "anthropic maps to ANTHROPIC vars",
			provider: ProviderAnthropic,
			apiKey:   "sk-ant-test",
			model:    "claude-opus-4-6",
			endpoint: "",
			expected: []string{
				"export ANTHROPIC_API_KEY=sk-ant-test",
				"export ANTHROPIC_MODEL=claude-opus-4-6",
			},
		},
		{
			name:     "minimax uses AUTH_TOKEN and sets all model aliases",
			provider: ProviderMinimax,
			apiKey:   "mmx_test",
			model:    "MiniMax-M2.5",
			endpoint: "https://api.minimax.chat",
			expected: []string{
				"export ANTHROPIC_AUTH_TOKEN=mmx_test",
				"export ANTHROPIC_MODEL=MiniMax-M2.5",
				"export ANTHROPIC_BASE_URL=https://api.minimax.chat",
				"export ANTHROPIC_DEFAULT_SONNET_MODEL=MiniMax-M2.5",
				"export ANTHROPIC_DEFAULT_OPUS_MODEL=MiniMax-M2.5",
				"export ANTHROPIC_DEFAULT_HAIKU_MODEL=MiniMax-M2.5",
			},
		},
		{
			name:     "custom no model",
			provider: ProviderCustom,
			apiKey:   "test-key",
			model:    "",
			endpoint: "",
			expected: []string{
				"export ANTHROPIC_API_KEY=test-key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExportVars(tt.provider, tt.apiKey, tt.model, tt.endpoint)
			// Check that all expected lines are present (order-independent)
			resultSet := make(map[string]bool, len(result))
			for _, l := range result {
				resultSet[l] = true
			}
			for _, want := range tt.expected {
				if !resultSet[want] {
					t.Errorf("ExportVars() missing line %q\ngot: %v", want, result)
				}
			}
		})
	}
}

func TestExportVarsPowerShell(t *testing.T) {
	result := ExportVarsPowerShell(ProviderMinimax, "mmx_test", "MiniMax-M2.5", "https://api.minimax.chat")
	expected := []string{
		`$env:ANTHROPIC_AUTH_TOKEN="mmx_test"`,
		`$env:ANTHROPIC_MODEL="MiniMax-M2.5"`,
		`$env:ANTHROPIC_BASE_URL="https://api.minimax.chat"`,
		`$env:ANTHROPIC_DEFAULT_SONNET_MODEL="MiniMax-M2.5"`,
	}
	resultSet := make(map[string]bool, len(result))
	for _, l := range result {
		resultSet[l] = true
	}
	for _, want := range expected {
		if !resultSet[want] {
			t.Errorf("ExportVarsPowerShell() missing line %q\ngot: %v", want, result)
		}
	}
}

func TestAllProviders(t *testing.T) {
	providers := AllProviders()
	if len(providers) < 5 {
		t.Errorf("AllProviders() returned %d providers, want at least 5", len(providers))
	}
}

func TestProviderString(t *testing.T) {
	p := ProviderAnthropic
	if p.String() != "anthropic" {
		t.Errorf("ProviderType.String() = %v, want %v", p.String(), "anthropic")
	}
}

func TestProviderMarshalJSON(t *testing.T) {
	p := ProviderAnthropic
	data, err := p.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON() error = %v", err)
	}
	if string(data) != `"anthropic"` {
		t.Errorf("MarshalJSON() = %v, want %v", string(data), `"anthropic"`)
	}
}

func TestProviderUnmarshalJSON(t *testing.T) {
	var p ProviderType
	err := p.UnmarshalJSON([]byte(`"minimax"`))
	if err != nil {
		t.Errorf("UnmarshalJSON() error = %v", err)
	}
	if p != ProviderMinimax {
		t.Errorf("UnmarshalJSON() = %v, want %v", p, ProviderMinimax)
	}
}

func TestDetectFromEnv_Vertex(t *testing.T) {
	// Save original
	orig := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	defer func() {
		if orig != "" {
			os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", orig)
		} else {
			os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		}
	}()

	// Clear all env vars first
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("MINIMAX_API_KEY")
	os.Unsetenv("AZURE_OPENAI_API_KEY")

	// Set Vertex
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/path/to/creds.json")
	result := DetectFromEnv()
	if result != ProviderVertex {
		t.Errorf("DetectFromEnv() = %v, want %v", result, ProviderVertex)
	}
}

func TestUnmarshalJSON_Error(t *testing.T) {
	var p ProviderType
	err := p.UnmarshalJSON([]byte("{invalid json}"))
	if err == nil {
		t.Error("UnmarshalJSON should fail with invalid JSON")
	}
}
