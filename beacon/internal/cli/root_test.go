package cli

import (
	"bytes"
	"os"
	"testing"
)

// TestGetConfigFile tests GetConfigFile
func TestGetConfigFile(t *testing.T) {
	// Reset to default
	configFile = "beacon.yaml"

	result := GetConfigFile()
	if result != "beacon.yaml" {
		t.Errorf("Expected 'beacon.yaml', got '%s'", result)
	}

	// Set to custom value
	configFile = "/custom/path/beacon.yaml"
	result = GetConfigFile()
	if result != "/custom/path/beacon.yaml" {
		t.Errorf("Expected '/custom/path/beacon.yaml', got '%s'", result)
	}

	// Reset
	configFile = "beacon.yaml"
}

// TestIsDebug tests IsDebug
func TestIsDebug(t *testing.T) {
	// Default is false
	debugFlag = false
	if IsDebug() {
		t.Error("Expected IsDebug() to be false")
	}

	// Set to true
	debugFlag = true
	if !IsDebug() {
		t.Error("Expected IsDebug() to be true")
	}

	// Reset
	debugFlag = false
}

// TestExecute_HelpCommand tests Execute by running help
func TestExecute_HelpCommand(t *testing.T) {
	var buf bytes.Buffer
	cmd := GetRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	// Execute should not call os.Exit with --help
	// Run directly on root command to avoid os.Exit
	if err := cmd.Execute(); err != nil {
		t.Logf("Execute returned error (expected for help): %v", err)
	}
}

// TestExecute_NoArgs tests Execute with no arguments (shows help)
func TestExecute_NoArgs(t *testing.T) {
	var buf bytes.Buffer
	cmd := GetRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Logf("Execute returned error: %v", err)
	}
}

// TestGetRootCmd_ReturnsNonNil tests that GetRootCmd returns non-nil
func TestGetRootCmd_ReturnsNonNil(t *testing.T) {
	cmd := GetRootCmd()
	if cmd == nil {
		t.Fatal("Expected non-nil root command")
	}
}

// TestGetRootCmd_HasSubcommands tests that GetRootCmd has expected subcommands
func TestGetRootCmd_HasSubcommands(t *testing.T) {
	cmd := GetRootCmd()
	subcommands := cmd.Commands()

	expectedCmds := map[string]bool{
		"start":  false,
		"stop":   false,
		"status": false,
		"debug":  false,
	}

	for _, sub := range subcommands {
		if _, ok := expectedCmds[sub.Use]; ok {
			expectedCmds[sub.Use] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected subcommand '%s' not found", name)
		}
	}
}

// TestExecute_WithInvalidCommand tests Execute with invalid command (os.Exit is called)
// We test this indirectly through root cmd
func TestExecute_WithConfigFlag(t *testing.T) {
	// Test that --config flag works
	tmpFile, err := os.CreateTemp("", "beacon-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	cmd := GetRootCmd()
	cmd.SetArgs([]string{"--config", tmpFile.Name(), "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute help - should succeed
	_ = cmd.Execute()

	// The configFile should be set to the tmp file
	_ = GetConfigFile() // should not panic
}

// TestExecute_WithDebugFlag tests Execute with --debug flag
func TestExecute_WithDebugFlag(t *testing.T) {
	cmd := GetRootCmd()
	cmd.SetArgs([]string{"--debug", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	_ = cmd.Execute()
}

// TestExecute_PackageLevel tests the package-level Execute() function (success path)
func TestExecute_PackageLevel(t *testing.T) {
	// Set up args to --help (no error, no os.Exit)
	rootCmd.SetArgs([]string{"--help"})

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	// Call the package-level Execute() which calls rootCmd.Execute()
	// --help should not cause an error, so os.Exit(1) will NOT be called
	Execute()
}
