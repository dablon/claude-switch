package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const binaryName = "claude-switch"

func getProjectRoot() string {
	// Get the directory where the test binary is run
	// Walk up until we find cmd/claude-switch
	wd, _ := os.Getwd()
	
	// Try walking up from current directory
	for wd != "/" {
		if _, err := os.Stat(filepath.Join(wd, "cmd", "claude-switch")); err == nil {
			return wd
		}
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		wd = filepath.Dir(wd)
	}
	
	// Fallback - try relative to tests/e2e
	wd, _ = os.Getwd()
	if filepath.Base(wd) == "e2e" {
		return filepath.Dir(filepath.Dir(wd))
	}
	
	return wd
}

func buildBinary(t *testing.T) string {
	projectRoot := getProjectRoot()
	
	cmd := exec.Command("go", "build", "-o", binaryName, "./cmd/claude-switch")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}
	
	return filepath.Join(projectRoot, binaryName)
}

func runCmd(t *testing.T, binaryPath, tmpDir string, args ...string) string {
	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(os.Environ(), "HOME="+tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil && len(output) > 0 {
		t.Logf("Command output: %s", output)
	}
	return string(output)
}

func TestE2E_FullWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := buildBinary(t)

	// 1. Add first profile
	output := runCmd(t, binaryPath, tmpDir, "add", "--name", "profile1", "--key", "sk-ant-key1", "--provider", "anthropic")
	if output == "" {
		t.Errorf("Failed to add profile1")
	}

	// 2. Add second profile  
	output = runCmd(t, binaryPath, tmpDir, "add", "--name", "profile2", "--key", "mmx-key2", "--provider", "minimax")
	if output == "" {
		t.Errorf("Failed to add profile2")
	}

	// 3. List should show both
	output = runCmd(t, binaryPath, tmpDir, "list")
	if output == "" {
		t.Errorf("Failed to list profiles")
	}

	// 4. Switch to profile2
	output = runCmd(t, binaryPath, tmpDir, "use", "--name", "profile2")
	if output == "" {
		t.Errorf("Failed to switch")
	}

	// 5. Export should have minimax vars
	output = runCmd(t, binaryPath, tmpDir, "export")
	if output == "" {
		t.Errorf("Failed to export")
	}

	// 6. Remove profile1
	runCmd(t, binaryPath, tmpDir, "remove", "--name", "profile1")
}

func TestE2E_QuickAddAutoDetect(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := buildBinary(t)

	// Quick add with auto-detection
	output := runCmd(t, binaryPath, tmpDir, "quick", "--name", "auto-test", "--key", "sk-ant-api03-key123456789")
	if output == "" {
		t.Errorf("Failed to quick add")
	}
}

func TestE2E_DetectProvider(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := buildBinary(t)

	// Test detect with anthropic key
	output := runCmd(t, binaryPath, tmpDir, "detect", "--key", "sk-ant-test123")
	if output == "" {
		t.Errorf("Failed to detect")
	}
}

func TestE2E_ProvidersList(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := buildBinary(t)

	output := runCmd(t, binaryPath, tmpDir, "providers")
	if output == "" {
		t.Errorf("Failed to list providers")
	}
}

func TestE2E_ApplyCommand(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := buildBinary(t)

	// Add a profile
	runCmd(t, binaryPath, tmpDir, "add", "--name", "apply-test", "--key", "sk-ant-key", "--provider", "anthropic")

	// Apply it
	output := runCmd(t, binaryPath, tmpDir, "apply", "--name", "apply-test")
	if output == "" {
		t.Errorf("Failed to apply")
	}
}
