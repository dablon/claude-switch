package config

import (
	"os"
	"path/filepath"
	"testing"

	"claude-switch/internal/provider"
)

// TestConfig is a helper to create temporary config
type TestConfig struct {
	Dir  string
	File string
}

func NewTestConfig(t *testing.T) *TestConfig {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	// Temporarily override config paths
	oldDir := configDir
	oldFile := configFile
	configDir = tmpDir
	configFile = configFile
	defer func() {
		configDir = oldDir
		configFile = oldFile
	}()

	return &TestConfig{
		Dir:  tmpDir,
		File: configFile,
	}
}

func TestLoad_Empty(t *testing.T) {
	// Use temp dir
	tmpDir := t.TempDir()
	oldDir := configDir
	oldFile := configFile
	configDir = tmpDir
	configFile = filepath.Join(tmpDir, "config.json")
	defer func() {
		configDir = oldDir
		configFile = oldFile
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Profiles) != 0 {
		t.Errorf("Load() returned %d profiles, want 0", len(cfg.Profiles))
	}

	if cfg.CurrentProfile != "" {
		t.Errorf("Load() CurrentProfile = %v, want empty", cfg.CurrentProfile)
	}
}

func TestAddProfile_New(t *testing.T) {
	cfg := &Config{}

	p := Profile{
		Name:     "test",
		Provider: provider.ProviderAnthropic,
		Model:    "claude-opus-4-6",
		APIKey:   "sk-ant-test",
	}

	err := AddProfile(cfg, p)
	if err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	if len(cfg.Profiles) != 1 {
		t.Errorf("AddProfile() left %d profiles, want 1", len(cfg.Profiles))
	}

	if cfg.CurrentProfile != "test" {
		t.Errorf("AddProfile() set CurrentProfile = %v, want 'test'", cfg.CurrentProfile)
	}
}

func TestAddProfile_Update(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{Name: "test", Provider: provider.ProviderAnthropic, Model: "claude-opus-4-6"},
		},
		CurrentProfile: "test",
	}

	p := Profile{
		Name:     "test",
		Provider: provider.ProviderMinimax,
		Model:    "minimax-M2.5",
		APIKey:   "mmx-test",
	}

	err := AddProfile(cfg, p)
	if err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	if len(cfg.Profiles) != 1 {
		t.Errorf("AddProfile() left %d profiles, want 1", len(cfg.Profiles))
	}

	if cfg.Profiles[0].Provider != provider.ProviderMinimax {
		t.Errorf("AddProfile() didn't update provider = %v", cfg.Profiles[0].Provider)
	}
}

func TestAddProfile_SecondProfile(t *testing.T) {
	cfg := &Config{
		Profiles:       []Profile{{Name: "first", Provider: provider.ProviderAnthropic}},
		CurrentProfile: "first",
	}

	AddProfile(cfg, Profile{Name: "second", Provider: provider.ProviderMinimax})

	if len(cfg.Profiles) != 2 {
		t.Errorf("AddProfile() created %d profiles, want 2", len(cfg.Profiles))
	}

	// First profile should remain current
	if cfg.CurrentProfile != "first" {
		t.Errorf("AddProfile() changed CurrentProfile = %v, want 'first'", cfg.CurrentProfile)
	}
}

func TestRemoveProfile_Existing(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{Name: "test", Provider: provider.ProviderAnthropic},
			{Name: "other", Provider: provider.ProviderMinimax},
		},
		CurrentProfile: "test",
	}

	err := RemoveProfile(cfg, "test")
	if err != nil {
		t.Fatalf("RemoveProfile() error = %v", err)
	}

	if len(cfg.Profiles) != 1 {
		t.Errorf("RemoveProfile() left %d profiles, want 1", len(cfg.Profiles))
	}

	if cfg.Profiles[0].Name != "other" {
		t.Errorf("RemoveProfile() removed wrong profile")
	}
}

func TestRemoveProfile_Last(t *testing.T) {
	cfg := &Config{
		Profiles:       []Profile{{Name: "test", Provider: provider.ProviderAnthropic}},
		CurrentProfile: "test",
	}

	err := RemoveProfile(cfg, "test")
	if err != nil {
		t.Fatalf("RemoveProfile() error = %v", err)
	}

	if len(cfg.Profiles) != 0 {
		t.Errorf("RemoveProfile() left %d profiles, want 0", len(cfg.Profiles))
	}

	if cfg.CurrentProfile != "" {
		t.Errorf("RemoveProfile() didn't clear CurrentProfile")
	}
}

func TestRemoveProfile_NotFound(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{Name: "test", Provider: provider.ProviderAnthropic},
		},
	}

	err := RemoveProfile(cfg, "nonexistent")
	if err == nil {
		t.Error("RemoveProfile() expected error for nonexistent profile")
	}
}

func TestGetCurrentProfile_Exists(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{Name: "test", Provider: provider.ProviderAnthropic},
			{Name: "other", Provider: provider.ProviderMinimax},
		},
		CurrentProfile: "test",
	}

	p := GetCurrentProfile(cfg)
	if p == nil {
		t.Fatal("GetCurrentProfile() returned nil")
	}

	if p.Name != "test" {
		t.Errorf("GetCurrentProfile() = %v, want 'test'", p.Name)
	}
}

func TestGetProfile(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{Name: "test", Provider: provider.ProviderAnthropic},
			{Name: "other", Provider: provider.ProviderMinimax},
		},
	}

	// Test existing profile
	p := GetProfile(cfg, "test")
	if p == nil {
		t.Fatal("GetProfile() returned nil")
	}
	if p.Name != "test" {
		t.Errorf("GetProfile() = %v, want 'test'", p.Name)
	}

	// Test non-existent profile
	p = GetProfile(cfg, "nonexistent")
	if p != nil {
		t.Errorf("GetProfile() = %v, want nil", p)
	}
}

func TestSetCurrent_Exists(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{Name: "test", Provider: provider.ProviderAnthropic},
			{Name: "other", Provider: provider.ProviderMinimax},
		},
	}

	err := SetCurrent(cfg, "other")
	if err != nil {
		t.Fatalf("SetCurrent() error = %v", err)
	}

	if cfg.CurrentProfile != "other" {
		t.Errorf("SetCurrent() = %v, want 'other'", cfg.CurrentProfile)
	}
}

func TestSetCurrent_NotFound(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{Name: "test", Provider: provider.ProviderAnthropic},
		},
	}

	err := SetCurrent(cfg, "nonexistent")
	if err == nil {
		t.Error("SetCurrent() expected error for nonexistent profile")
	}
}

func TestSortedProfiles(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{Name: "zebra", Provider: provider.ProviderMinimax},
			{Name: "apple", Provider: provider.ProviderAnthropic},
			{Name: "banana", Provider: provider.ProviderOpenAI},
		},
	}

	sorted := SortedProfiles(cfg)

	if len(sorted) != 3 {
		t.Fatalf("SortedProfiles() returned %d, want 3", len(sorted))
	}

	if sorted[0].Name != "apple" {
		t.Errorf("SortedProfiles()[0] = %v, want 'apple'", sorted[0].Name)
	}

	if sorted[1].Name != "banana" {
		t.Errorf("SortedProfiles()[1] = %v, want 'banana'", sorted[1].Name)
	}

	if sorted[2].Name != "zebra" {
		t.Errorf("SortedProfiles()[2] = %v, want 'zebra'", sorted[2].Name)
	}
}

func TestExportEnvVars(t *testing.T) {
	p := &Profile{
		Name:     "test",
		Provider: provider.ProviderAnthropic,
		Model:    "claude-opus-4-6",
		APIKey:   "sk-ant-test",
	}

	lines := ExportEnvVars(p)

	if len(lines) != 2 {
		t.Errorf("ExportEnvVars() returned %d lines, want 2", len(lines))
	}

	if lines[0] != "export ANTHROPIC_API_KEY=sk-ant-test" {
		t.Errorf("ExportEnvVars()[0] = %v", lines[0])
	}

	if lines[1] != "export ANTHROPIC_MODEL=claude-opus-4-6" {
		t.Errorf("ExportEnvVars()[1] = %v", lines[1])
	}
}

func TestDetectAndCreateProfile_AutoDetect(t *testing.T) {
	p := DetectAndCreateProfile("test", "sk-ant-test", "", "")

	if p.Name != "test" {
		t.Errorf("Name = %v, want 'test'", p.Name)
	}

	if p.Provider != provider.ProviderAnthropic {
		t.Errorf("Provider = %v, want %v", p.Provider, provider.ProviderAnthropic)
	}

	// Should get default model
	if p.Model == "" {
		t.Error("Model should not be empty")
	}
}

func TestDetectAndCreateProfile_Explicit(t *testing.T) {
	p := DetectAndCreateProfile("test", "sk-ant-test", "claude-sonnet-4-20250514", "")

	if p.Model != "claude-sonnet-4-20250514" {
		t.Errorf("Model = %v, want 'claude-sonnet-4-20250514'", p.Model)
	}
}

func TestSave_AndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir := configDir
	oldFile := configFile
	configDir = tmpDir
	configFile = filepath.Join(tmpDir, "config.json")
	defer func() {
		configDir = oldDir
		configFile = oldFile
	}()

	cfg := &Config{
		Profiles: []Profile{
			{Name: "test", Provider: provider.ProviderAnthropic, Model: "claude-opus-4-6"},
		},
		CurrentProfile: "test",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded.Profiles) != 1 {
		t.Errorf("Load() returned %d profiles, want 1", len(loaded.Profiles))
	}

	if loaded.CurrentProfile != "test" {
		t.Errorf("Load() CurrentProfile = %v, want 'test'", loaded.CurrentProfile)
	}
}

func TestLoad_FileReadError(t *testing.T) {
	// Skip if running as root (root can read any file)
	if os.Geteuid() == 0 {
		t.Skip("Skipping test when running as root")
	}
	
	tmpDir := t.TempDir()
	oldDir := configDir
	oldFile := configFile
	configDir = tmpDir
	configFile = filepath.Join(tmpDir, "config.json")
	defer func() {
		configDir = oldDir
		configFile = oldFile
	}()

	// Create file with no read permissions
	os.WriteFile(configFile, []byte("{}"), 0000)

	_, err := Load()
	if err == nil {
		t.Error("Load() should fail with unreadable file")
	}
}

func TestLoad_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir := configDir
	oldFile := configFile
	configDir = tmpDir
	configFile = filepath.Join(tmpDir, "config.json")
	defer func() {
		configDir = oldDir
		configFile = oldFile
	}()

	// Create a valid config file
	cfg := &Config{
		Profiles: []Profile{
			{Name: "test", Provider: provider.ProviderAnthropic},
		},
		CurrentProfile: "test",
	}
	Save(cfg)

	// Load should succeed
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Profiles) != 1 {
		t.Errorf("Load() returned %d profiles, want 1", len(loaded.Profiles))
	}
}


func TestGetCurrentProfile_MissingInList(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{Name: "profile1", Provider: provider.ProviderAnthropic},
		},
		CurrentProfile: "nonexistent", // Points to a profile that doesn't exist
	}

	p := GetCurrentProfile(cfg)
	if p != nil {
		t.Error("GetCurrentProfile() should return nil when current profile doesn't exist in list")
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a subdir that doesn't exist
	tmpDir = filepath.Join(tmpDir, "subdir")

	oldDir := configDir
	oldFile := configFile
	configDir = tmpDir
	configFile = filepath.Join(tmpDir, "config.json")
	defer func() {
		configDir = oldDir
		configFile = oldFile
	}()

	cfg := &Config{}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Save() didn't create config file")
	}
}

func TestAddProfile_UpdateExisting(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{Name: "test", Provider: provider.ProviderAnthropic, Model: "claude-opus-4-6"},
		},
		CurrentProfile: "test",
	}

	// Update with new values
	p := Profile{
		Name:     "test",
		Provider: provider.ProviderMinimax,
		Model:    "minimax-M2.5",
		APIKey:   "mmx-key",
	}

	err := AddProfile(cfg, p)
	if err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	// Should only have one profile
	if len(cfg.Profiles) != 1 {
		t.Errorf("AddProfile() left %d profiles, want 1", len(cfg.Profiles))
	}

	// Should have updated values
	if cfg.Profiles[0].Provider != provider.ProviderMinimax {
		t.Errorf("Provider not updated: %v", cfg.Profiles[0].Provider)
	}
}

func TestLoad_WithExistingProfile(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir := configDir
	oldFile := configFile
	configDir = tmpDir
	configFile = filepath.Join(tmpDir, "config.json")
	defer func() {
		configDir = oldDir
		configFile = oldFile
	}()

	// Save a config first
	cfg := &Config{
		Profiles: []Profile{
			{Name: "p1", Provider: provider.ProviderAnthropic, Model: "claude-opus-4-6"},
			{Name: "p2", Provider: provider.ProviderMinimax, Model: "minimax-M2.5"},
		},
		CurrentProfile: "p2",
	}
	Save(cfg)

	// Load should return saved config
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Profiles) != 2 {
		t.Errorf("Load() returned %d profiles, want 2", len(loaded.Profiles))
	}
	if loaded.CurrentProfile != "p2" {
		t.Errorf("Load() CurrentProfile = %v, want 'p2'", loaded.CurrentProfile)
	}
}

func TestSave_AndLoadMultipleProfiles(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir := configDir
	oldFile := configFile
	configDir = tmpDir
	configFile = filepath.Join(tmpDir, "config.json")
	defer func() {
		configDir = oldDir
		configFile = oldFile
	}()

	cfg := &Config{
		Profiles: []Profile{
			{Name: "a", Provider: provider.ProviderAnthropic},
			{Name: "b", Provider: provider.ProviderOpenAI},
			{Name: "c", Provider: provider.ProviderMinimax},
		},
		CurrentProfile: "b",
	}
	Save(cfg)

	loaded, _ := Load()
	if len(loaded.Profiles) != 3 {
		t.Errorf("Expected 3 profiles, got %d", len(loaded.Profiles))
	}
}

// Mock filesystem for testing
type mockFS struct {
	statErr    error
	readErr    error
	mkdirErr  error
	writeErr  error
	statIsDir bool
}

func (m mockFS) Stat(path string) (os.FileInfo, error) {
	if m.statErr != nil {
		return nil, m.statErr
	}
	if m.statIsDir {
		return nil, os.ErrInvalid
	}
	return nil, nil
}

func (m mockFS) ReadFile(path string) ([]byte, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	return []byte("{}"), nil
}

func (m mockFS) MkdirAll(path string, perm os.FileMode) error {
	if m.mkdirErr != nil {
		return m.mkdirErr
	}
	return nil
}

func (m mockFS) WriteFile(path string, data []byte, perm os.FileMode) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	return nil
}

func TestLoad_ReadError(t *testing.T) {
	// Save original fs
	original := fs
	defer func() { fs = original }()

	// Set mock that returns error on read
	fs = mockFS{readErr: os.ErrPermission}
	defer func() { fs = original }()

	cfg, err := Load()
	if err == nil {
		t.Error("Load() should return error when ReadFile fails")
	}
	if cfg != nil {
		t.Error("Load() should return nil config on error")
	}
}

func TestLoad_StatError(t *testing.T) {
	original := fs
	defer func() { fs = original }()

	fs = mockFS{statErr: os.ErrPermission}
	defer func() { fs = original }()

	cfg, err := Load()
	if err == nil {
		t.Error("Load() should return error when Stat fails")
	}
	_ = cfg
}

func TestSave_MkdirError(t *testing.T) {
	original := fs
	defer func() { fs = original }()

	fs = mockFS{mkdirErr: os.ErrPermission}
	defer func() { fs = original }()

	cfg := &Config{}
	err := Save(cfg)
	if err == nil {
		t.Error("Save() should return error when MkdirAll fails")
	}
}

func TestSave_WriteError(t *testing.T) {
	original := fs
	defer func() { fs = original }()

	fs = mockFS{writeErr: os.ErrPermission}
	defer func() { fs = original }()

	cfg := &Config{}
	err := Save(cfg)
	if err == nil {
		t.Error("Save() should return error when WriteFile fails")
	}
}

func TestSetFileSystem(t *testing.T) {
	original := fs
	defer func() { fs = original }()
	
	SetFileSystem(mockFS{})
	// Just verify it doesn't panic
}

func TestLoad_StatErrorNotNotExist(t *testing.T) {
	original := fs
	defer func() { fs = original }()

	// Stat returns an error that's not "file not found"
	fs = mockFS{statErr: os.ErrPermission}
	
	cfg, err := Load()
	if err == nil {
		t.Error("Load() should return error when Stat fails")
	}
	if cfg != nil {
		t.Error("cfg should be nil on error")
	}
}

// Additional test for Load error path
type mockFSWithInvalidData struct {
	mockFS
}

func (m mockFSWithInvalidData) ReadFile(path string) ([]byte, error) {
	return []byte("not valid json{"), nil
}

func TestLoad_JSONUnmarshalError(t *testing.T) {
	original := fs
	defer func() { fs = original }()

	fs = mockFSWithInvalidData{}
	defer func() { fs = original }()

	cfg, err := Load()
	if err == nil {
		t.Error("Load() should return error for invalid JSON")
	}
	if cfg != nil {
		t.Error("cfg should be nil on error")
	}
}
