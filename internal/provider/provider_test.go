package provider

import (
	"strings"
	"os"
	"testing"
)

func TestDetectProvider_Anthropic(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected ProviderType
	}{
		{"anthropic key", "sk-ant-api03-test", ProviderAnthropic},
		{"anthropic longer key", "sk-ant-api03-abc123def456ghi789jkl012mno345pqr", ProviderAnthropic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectProvider(tt.apiKey)
			if result != tt.expected {
				t.Errorf("DetectProvider(%q) = %v, want %v", tt.apiKey, result, tt.expected)
			}
		})
	}
}

func TestDetectProvider_OpenAI(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected ProviderType
	}{
		{"openai key", "sk-abcdefghijklmnopqrstuvwxyz123456789", ProviderOpenAI},
		{"openai longer", "sk-abc123def456ghi789jkl012mno345pqr678stu", ProviderOpenAI},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectProvider(tt.apiKey)
			if result != tt.expected {
				t.Errorf("DetectProvider(%q) = %v, want %v", tt.apiKey, result, tt.expected)
			}
		})
	}
}

func TestDetectProvider_Minimax(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected ProviderType
	}{
		{"minimax key", "mmx_yYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY", ProviderMinimax},
		{"long random key", "verylongkeythatdoesnotstartwithskorskant123456789", ProviderMinimax},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectProvider(tt.apiKey)
			if result != tt.expected {
				t.Errorf("DetectProvider(%q) = %v, want %v", tt.apiKey, result, tt.expected)
			}
		})
	}
}

func TestDetectProvider_ShortKey(t *testing.T) {
	// Short keys should return custom
	result := DetectProvider("sk")
	if result != ProviderCustom {
		t.Errorf("DetectProvider('sk') = %v, want %v", result, ProviderCustom)
	}
}

func TestDetectProvider_Vertex(t *testing.T) {
	// Vertex uses GOOGLE_APPLICATION_CREDENTIALS but the key format varies
	// Very long keys with specific patterns get detected as minimax
	result := DetectProvider("ya29.a0AfH6SMBx1234567890abcdefghijklmnopqrstuvwxyz1234567890")
	if result != ProviderMinimax {
		t.Errorf("DetectProvider() = %v, want %v", result, ProviderMinimax)
	}
}

func TestDetectProvider_AzureEdgeCase(t *testing.T) {
	// Azure keys are typically very long (80+ chars)
	result := DetectProvider("this-is-a-very-long-azure-key-with-lots-of-characters-that-makes-it-obviously-azure-123456789")
	if result != ProviderAzure {
		t.Errorf("DetectProvider() = %v, want %v", result, ProviderAzure)
	}
}

func TestDetectProvider_Empty(t *testing.T) {
	result := DetectProvider("")
	if result != ProviderCustom {
		t.Errorf("DetectProvider('') = %v, want %v", result, ProviderCustom)
	}
}

func TestGetDefaultModel(t *testing.T) {
	tests := []struct {
		provider ProviderType
		expected string
	}{
		{ProviderAnthropic, "claude-opus-4-6"},
		{ProviderOpenAI, "gpt-5"},
		{ProviderMinimax, "minimax-M2.5"},
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
		{ProviderMinimax, ""},
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

	// Test with OPENAI_API_KEY (should take precedence if set after)
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
			name:     "anthropic",
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
			name:     "minimax with endpoint",
			provider: ProviderMinimax,
			apiKey:   "mmx_test",
			model:    "minimax-M2.5",
			endpoint: "https://api.minimax.chat",
			expected: []string{
				"export MINIMAX_API_KEY=mmx_test",
				"export MINIMAX_MODEL=minimax-M2.5",
				"export MINIMAX_ENDPOINT=https://api.minimax.chat",
			},
		},
		{
			name:     "custom no model",
			provider: ProviderCustom,
			apiKey:   "test-key",
			model:    "",
			endpoint: "",
			expected: []string{
				"export CUSTOM_API_KEY=test-key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExportVars(tt.provider, tt.apiKey, tt.model, tt.endpoint)
			if len(result) != len(tt.expected) {
				t.Errorf("ExportVars() returned %d lines, want %d", len(result), len(tt.expected))
				return
			}
			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf("ExportVars()[%d] = %v, want %v", i, line, tt.expected[i])
				}
			}
		})
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

func TestDetectProvider_AdditionalCases(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected ProviderType
	}{
		{"very long key", strings.Repeat("x", 100), ProviderAzure},
		{"medium sk", "sk-abc123def456ghi789", ProviderOpenAI},
		{"non-sk long", "abcdefghijklmnopqrstuvwxyz12345678901234567890", ProviderMinimax},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectProvider(tt.apiKey)
			if result != tt.expected {
				t.Errorf("DetectProvider(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestDetectProvider_AllPatterns(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected ProviderType
	}{
		// Anthropic
		{"ant basic", "sk-ant-api01", ProviderAnthropic},
		{"ant long", "sk-ant-api03-abc123def456ghi789jkl012mno345pqr678stu901vwx", ProviderAnthropic},
		// OpenAI
		{"openai basic", "sk-abcdefghijklmnopqrstuvwxyz123456789", ProviderOpenAI},
		// Azure (very long)
		{"azure long", strings.Repeat("z", 85), ProviderAzure},
		// Minimax
		{"minimax", "mmx_abc123def456ghi789jkl012mno345pqr678stu", ProviderMinimax},
		{"random long", "abc123def456ghi789jkl012mno345pqr678stu901vwx234", ProviderMinimax},
		// Custom
		{"empty", "", ProviderCustom},
		{"short", "abc", ProviderCustom},
		{"tiny sk", "sk", ProviderCustom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectProvider(tt.apiKey)
			if result != tt.expected {
				t.Errorf("DetectProvider(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
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
