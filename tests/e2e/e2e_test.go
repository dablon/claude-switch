package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const binaryName = "claude-switch"

func getProjectRoot() string {
	wd, _ := os.Getwd()
	if strings.HasSuffix(wd, "/tests/e2e") || strings.HasSuffix(wd, "\\tests\\e2e") {
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
	if !strings.Contains(output, "profile1") {
		t.Errorf("Failed to add profile1: %s", output)
	}

	// 2. Add second profile  
	output = runCmd(t, binaryPath, tmpDir, "add", "--name", "profile2", "--key", "mmx-key2", "--provider", "minimax")
	if !strings.Contains(output, "profile2") {
		t.Errorf("Failed to add profile2: %s", output)
	}

	// 3. List should show both
	output = runCmd(t, binaryPath, tmpDir, "list")
	if !strings.Contains(output, "profile1") || !strings.Contains(output, "profile2") {
		t.Errorf("Expected both profiles in list: %s", output)
	}

	// 4. Switch to profile2
	output = runCmd(t, binaryPath, tmpDir, "use", "--name", "profile2")
	if !strings.Contains(output, "profile2") {
		t.Errorf("Failed to switch: %s", output)
	}

	// 5. Show should have profile2
	output = runCmd(t, binaryPath, tmpDir, "show")
	if !strings.Contains(output, "profile2") {
		t.Errorf("Expected profile2 in show: %s", output)
	}

	// 6. Export should have minimax vars
	output = runCmd(t, binaryPath, tmpDir, "export")
	if !strings.Contains(output, "MINIMAX_API_KEY") {
		t.Errorf("Expected MINIMAX_API_KEY in export: %s", output)
	}

	// 7. Remove profile1
	output = runCmd(t, binaryPath, tmpDir, "remove", "--name", "profile1")

	// 8. List should only have profile2
	output = runCmd(t, binaryPath, tmpDir, "list")
	if strings.Contains(output, "profile1") {
		t.Errorf("profile1 should be removed: %s", output)
	}
	if !strings.Contains(output, "profile2") {
		t.Errorf("profile2 should remain: %s", output)
	}
}

func TestE2E_QuickAddAutoDetect(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := buildBinary(t)

	// Quick add with auto-detection
	output := runCmd(t, binaryPath, tmpDir, "quick", "--name", "auto-test", "--key", "sk-ant-api03-key123456789")

	// Should auto-detect anthropic
	if !strings.Contains(output, "anthropic") {
		t.Errorf("Expected anthropic detection: %s", output)
	}
}

func TestE2E_DetectProvider(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := buildBinary(t)

	// Test detect with anthropic key
	output := runCmd(t, binaryPath, tmpDir, "detect", "--key", "sk-ant-test123")
	if !strings.Contains(output, "anthropic") {
		t.Errorf("Expected anthropic detection: %s", output)
	}

	// Test detect with openai key
	output = runCmd(t, binaryPath, tmpDir, "detect", "--key", "sk-test12345678901234567890")
	if !strings.Contains(output, "openai") {
		t.Errorf("Expected openai detection: %s", output)
	}

	// Test detect with minimax key
	output = runCmd(t, binaryPath, tmpDir, "detect", "--key", "mmx_verylongkey123456789012345678901234567890")
	if !strings.Contains(output, "minimax") {
		t.Errorf("Expected minimax detection: %s", output)
	}
}

func TestE2E_ProvidersList(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := buildBinary(t)

	output := runCmd(t, binaryPath, tmpDir, "providers")

	if !strings.Contains(output, "anthropic") {
		t.Errorf("Expected anthropic in providers: %s", output)
	}
	if !strings.Contains(output, "minimax") {
		t.Errorf("Expected minimax in providers: %s", output)
	}
}

func TestE2E_ApplyCommand(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := buildBinary(t)

	// Add a profile
	runCmd(t, binaryPath, tmpDir, "add", "--name", "apply-test", "--key", "sk-ant-key", "--provider", "anthropic")

	// Apply it
	output := runCmd(t, binaryPath, tmpDir, "apply", "--name", "apply-test")
	if !strings.Contains(output, "apply-test") {
		t.Errorf("Expected apply-test in output: %s", output)
	}
}
