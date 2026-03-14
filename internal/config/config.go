package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"claude-switch/internal/provider"
)

type Profile struct {
	Name     string             `json:"name"`
	Provider provider.ProviderType `json:"provider"`
	Model    string             `json:"model"`
	APIKey   string             `json:"api_key,omitempty"`
	Endpoint string             `json:"endpoint,omitempty"`
}

type Config struct {
	Profiles       []Profile `json:"profiles"`
	CurrentProfile string    `json:"current_profile"`
}

var (
	configDir  = filepath.Join(os.Getenv("HOME"), ".claude-switch")
	configFile = filepath.Join(configDir, "config.json")
)

// Load reads the config from disk
func Load() (*Config, error) {
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

// Save writes the config to disk
func Save(c *Config) error {
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0600)
}

// AddProfile adds or updates a profile
func AddProfile(c *Config, p Profile) error {
	for i, existing := range c.Profiles {
		if existing.Name == p.Name {
			c.Profiles[i] = p
			return nil
		}
	}

	c.Profiles = append(c.Profiles, p)

	// If first profile, set as current
	if c.CurrentProfile == "" {
		c.CurrentProfile = p.Name
	}

	return nil
}

// RemoveProfile removes a profile by name
func RemoveProfile(c *Config, name string) error {
	newProfiles := []Profile{}
	found := false

	for _, p := range c.Profiles {
		if p.Name != name {
			newProfiles = append(newProfiles, p)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("profile not found: %s", name)
	}

	c.Profiles = newProfiles

	// Reset current if removed
	if c.CurrentProfile == name {
		if len(c.Profiles) > 0 {
			c.CurrentProfile = c.Profiles[0].Name
		} else {
			c.CurrentProfile = ""
		}
	}

	return nil
}

// GetCurrentProfile returns the current profile
func GetCurrentProfile(c *Config) *Profile {
	for i := range c.Profiles {
		if c.Profiles[i].Name == c.CurrentProfile {
			return &c.Profiles[i]
		}
	}
	return nil
}

// GetProfile returns a profile by name
func GetProfile(c *Config, name string) *Profile {
	for i := range c.Profiles {
		if c.Profiles[i].Name == name {
			return &c.Profiles[i]
		}
	}
	return nil
}

// SetCurrent sets the current profile
func SetCurrent(c *Config, name string) error {
	for _, p := range c.Profiles {
		if p.Name == name {
			c.CurrentProfile = name
			return nil
		}
	}
	return fmt.Errorf("profile not found: %s", name)
}

// SortedProfiles returns profiles sorted by name
func SortedProfiles(c *Config) []Profile {
	sorted := make([]Profile, len(c.Profiles))
	copy(sorted, c.Profiles)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	return sorted
}

// ExportEnvVars returns the environment variables for a profile
func ExportEnvVars(p *Profile) []string {
	return provider.ExportVars(p.Provider, p.APIKey, p.Model, p.Endpoint)
}

// DetectAndCreateProfile creates a profile with auto-detected provider
func DetectAndCreateProfile(name, apiKey, model, endpoint string) Profile {
	detectedProvider := provider.DetectProvider(apiKey)

	// If model not provided, use default for detected provider
	if model == "" {
		model = provider.GetDefaultModel(detectedProvider)
	}

	return Profile{
		Name:     name,
		Provider: detectedProvider,
		Model:    model,
		APIKey:   apiKey,
		Endpoint: endpoint,
	}
}
