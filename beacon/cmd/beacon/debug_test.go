package beacon

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// parseJSONFromOutput extracts JSON from output that may contain warning lines
func parseJSONFromOutput(t *testing.T, output string) map[string]interface{} {
	// Find the start of JSON (first line with "{")
	lines := strings.Split(output, "\n")
	startIdx := -1
	braceCount := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "{") {
			startIdx = i
			braceCount = strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
			break
		}
	}

	if startIdx == -1 {
		t.Fatalf("No JSON output found in:\n%s", output)
	}

	// Collect all lines until we have balanced braces
	var jsonLines []string
	jsonLines = append(jsonLines, lines[startIdx])

	for i := startIdx + 1; i < len(lines); i++ {
		line := lines[i]
		jsonLines = append(jsonLines, line)
		braceCount += strings.Count(line, "{") - strings.Count(line, "}")
		if braceCount == 0 {
			break
		}
	}

	jsonStr := strings.Join(jsonLines, "\n")

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v\nJSON:\n%s", err, jsonStr)
	}

	return result
}

// TestDebugCommand_JSONOutput tests JSON format diagnostic output (AC #1, #2, #3)
func TestDebugCommand_JSONOutput(t *testing.T) {
	// Create test config with probes
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-01"
node_name: "Test Node"
log_level: "INFO"
debug_mode: false
probes:
  - type: "tcp_ping"
    target: "192.168.1.1"
    port: 80
    timeout_seconds: 5
    interval: 60
    count: 10
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Capture output
	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetErr(&buf)
	GetRootCmd().SetArgs([]string{"--config", configPath, "debug"})

	// Execute command
	err := GetRootCmd().Execute()
	if err != nil {
		t.Fatalf("Debug command failed: %v", err)
	}

	output := buf.String()

	// Verify JSON format (AC #3)
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("Output is not valid JSON: %v\nOutput:\n%s", err, output)
	}

	// Check required fields (AC #2)
	requiredFields := []string{"timestamp", "level", "message", "diagnostics"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Missing required field: %s", field)
		}
	}

	// Check diagnostics sections
	diagnostics, ok := result["diagnostics"].(map[string]interface{})
	if !ok {
		t.Fatal("diagnostics field is not a map")
	}

	requiredSections := []string{"network_status", "configuration", "connection_status", "resource_usage"}
	for _, section := range requiredSections {
		if _, ok := diagnostics[section]; !ok {
			t.Errorf("Missing diagnostics section: %s", section)
		}
	}
}

// TestDebugCommand_PrettyOutput tests human-readable format (AC #3)
func TestDebugCommand_PrettyOutput(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-01"
node_name: "Test Node"
log_level: "INFO"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetArgs([]string{"--config", configPath, "debug", "--pretty"})

	err := GetRootCmd().Execute()
	if err != nil {
		t.Fatalf("Debug command failed: %v", err)
	}

	output := buf.String()

	// Check for pretty format indicators
	expectedHeaders := []string{
		"Beacon Diagnostic Information",
		"Network Status",
		"Configuration",
		"Resource Usage",
	}
	for _, header := range expectedHeaders {
		if !strings.Contains(output, header) {
			t.Errorf("Pretty output missing header: %s", header)
		}
	}

	// Pretty output should NOT contain raw JSON braces
	if strings.Contains(output, "{") && strings.Contains(output, "\"timestamp\"") {
		t.Error("Pretty output should not contain raw JSON format")
	}
}

// TestDebugCommand_NetworkStatus tests network status diagnostic (AC #2)
func TestDebugCommand_NetworkStatus(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-01"
node_name: "Test Node"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetErr(&buf)
	GetRootCmd().SetArgs([]string{"--config", configPath, "debug"})

	err := GetRootCmd().Execute()
	if err != nil {
		t.Fatalf("Debug command failed: %v", err)
	}

	result := parseJSONFromOutput(t, buf.String())

	diagnostics := result["diagnostics"].(map[string]interface{})
	networkStatus := diagnostics["network_status"].(map[string]interface{})

	// Check required network status fields
	requiredFields := []string{
		"pulse_server_reachable",
		"pulse_server_address",
		"packet_loss_rate",
	}
	for _, field := range requiredFields {
		if _, ok := networkStatus[field]; !ok {
			t.Errorf("Missing network status field: %s", field)
		}
	}

	// Verify pulse_server_address matches config
	if networkStatus["pulse_server_address"] != "http://localhost:8080" {
		t.Errorf("pulse_server_address mismatch: got %v", networkStatus["pulse_server_address"])
	}
}

// TestDebugCommand_ConfigurationInfo tests configuration diagnostic (AC #2)
func TestDebugCommand_ConfigurationInfo(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-node-01"
node_name: "Test Config Node"
log_level: "DEBUG"
debug_mode: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetErr(&buf)
	GetRootCmd().SetArgs([]string{"--config", configPath, "debug"})

	err := GetRootCmd().Execute()
	if err != nil {
		t.Fatalf("Debug command failed: %v", err)
	}

	result := parseJSONFromOutput(t, buf.String())

	diagnostics := result["diagnostics"].(map[string]interface{})
	config := diagnostics["configuration"].(map[string]interface{})

	// Check required config fields
	if config["config_file"] == nil {
		t.Error("config_file field is missing")
	}
	if config["log_level"] != "DEBUG" {
		t.Errorf("log_level mismatch: expected DEBUG, got %v", config["log_level"])
	}
	if config["debug_mode"] != true {
		t.Errorf("debug_mode mismatch: expected true, got %v", config["debug_mode"])
	}
}

// TestDebugCommand_ConnectionStatus tests connection retry status (AC #2)
func TestDebugCommand_ConnectionStatus(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-01"
node_name: "Test Node"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetErr(&buf)
	GetRootCmd().SetArgs([]string{"--config", configPath, "debug"})

	err := GetRootCmd().Execute()
	if err != nil {
		t.Fatalf("Debug command failed: %v", err)
	}

	result := parseJSONFromOutput(t, buf.String())

	diagnostics := result["diagnostics"].(map[string]interface{})
	connectionStatus := diagnostics["connection_status"].(map[string]interface{})

	// Check required connection status fields
	requiredFields := []string{
		"status",
		"retry_count",
		"backoff_seconds",
		"queue_size",
	}
	for _, field := range requiredFields {
		if _, ok := connectionStatus[field]; !ok {
			t.Errorf("Missing connection status field: %s", field)
		}
	}
}

// TestDebugCommand_ResourceUsage tests resource usage diagnostic (AC #2)
func TestDebugCommand_ResourceUsage(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-01"
node_name: "Test Node"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetErr(&buf)
	GetRootCmd().SetArgs([]string{"--config", configPath, "debug"})

	err := GetRootCmd().Execute()
	if err != nil {
		t.Fatalf("Debug command failed: %v", err)
	}

	result := parseJSONFromOutput(t, buf.String())

	diagnostics := result["diagnostics"].(map[string]interface{})
	resourceUsage := diagnostics["resource_usage"].(map[string]interface{})

	// Check required resource fields
	requiredFields := []string{"cpu_percent", "memory_mb", "memory_percent"}
	for _, field := range requiredFields {
		if _, ok := resourceUsage[field]; !ok {
			t.Errorf("Missing resource usage field: %s", field)
		}
	}

	// Verify values are reasonable
	cpuPercent, ok := resourceUsage["cpu_percent"].(float64)
	if !ok || cpuPercent < 0 || cpuPercent > 100 {
		t.Errorf("Invalid cpu_percent: %v", resourceUsage["cpu_percent"])
	}

	memoryPercent, ok := resourceUsage["memory_percent"].(float64)
	if !ok || memoryPercent < 0 || memoryPercent > 100 {
		t.Errorf("Invalid memory_percent: %v", resourceUsage["memory_percent"])
	}
}

// TestDebugCommand_ProbeTasks tests probe tasks diagnostic (AC #2)
func TestDebugCommand_ProbeTasks(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-01"
node_name: "Test Node"
probes:
  - type: "tcp_ping"
    target: "192.168.1.1"
    port: 80
    timeout_seconds: 5
    interval: 60
    count: 10
  - type: "udp_ping"
    target: "192.168.1.2"
    port: 53
    timeout_seconds: 3
    interval: 120
    count: 5
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetErr(&buf)
	GetRootCmd().SetArgs([]string{"--config", configPath, "debug"})

	err := GetRootCmd().Execute()
	if err != nil {
		t.Fatalf("Debug command failed: %v", err)
	}

	result := parseJSONFromOutput(t, buf.String())

	diagnostics := result["diagnostics"].(map[string]interface{})
	probeTasks := diagnostics["probe_tasks"].(map[string]interface{})

	// Check probe counts
	totalTasks, ok := probeTasks["total_tasks"].(float64)
	if !ok || totalTasks != 2 {
		t.Errorf("Expected 2 probe tasks, got: %v", totalTasks)
	}

	// Check tasks array
	tasks, ok := probeTasks["tasks"].([]interface{})
	if !ok || len(tasks) != 2 {
		t.Fatalf("Expected 2 tasks in array, got: %v", len(tasks))
	}

	// Check first task fields
	task1 := tasks[0].(map[string]interface{})
	if task1["type"] != "tcp_ping" {
		t.Errorf("First task type mismatch: got %v", task1["type"])
	}
	if task1["target"] != "192.168.1.1:80" {
		t.Errorf("First task target mismatch: got %v", task1["target"])
	}
}

// TestDebugCommand_DebugMode tests debug_mode configuration integration (AC #2, #3)
func TestDebugCommand_DebugMode(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")

	// Test with debug_mode: true
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-01"
node_name: "Test Node"
debug_mode: true
log_level: "WARN"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetErr(&buf)
	GetRootCmd().SetArgs([]string{"--config", configPath, "debug"})

	err := GetRootCmd().Execute()
	if err != nil {
		t.Fatalf("Debug command failed: %v", err)
	}

	result := parseJSONFromOutput(t, buf.String())

	// When debug_mode is true, log_level should be DEBUG
	diagnostics := result["diagnostics"].(map[string]interface{})
	config := diagnostics["configuration"].(map[string]interface{})

	if config["log_level"] != "DEBUG" {
		t.Errorf("debug_mode=true should override log_level to DEBUG, got: %v", config["log_level"])
	}
	if config["debug_mode"] != true {
		t.Errorf("debug_mode should be true, got: %v", config["debug_mode"])
	}
}

// TestDebugCommand_InvalidConfig tests debug command with invalid config
func TestDebugCommand_InvalidConfig(t *testing.T) {
	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetErr(&buf)
	GetRootCmd().SetArgs([]string{"--config", "/nonexistent/config.yaml", "debug"})

	err := GetRootCmd().Execute()
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}

	output := buf.String()
	if !strings.Contains(output, "error") {
		t.Error("Expected error message in output")
	}
}

// TestDebugCommand_StructuredLogFormat verifies structured log format (AC #3)
func TestDebugCommand_StructuredLogFormat(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "beacon.yaml")
	configContent := `
pulse_server: "http://localhost:8080"
node_id: "test-01"
node_name: "Test Node"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	var buf bytes.Buffer
	GetRootCmd().SetOut(&buf)
	GetRootCmd().SetErr(&buf)
	GetRootCmd().SetArgs([]string{"--config", configPath, "debug"})

	err := GetRootCmd().Execute()
	if err != nil {
		t.Fatalf("Debug command failed: %v", err)
	}

	result := parseJSONFromOutput(t, buf.String())

	// Check structured log fields (AC #3)
	if result["timestamp"] == nil {
		t.Error("Missing timestamp in structured log")
	}
	if result["level"] != "DEBUG" {
		t.Errorf("Expected level DEBUG, got: %v", result["level"])
	}
	if result["message"] == nil {
		t.Error("Missing message in structured log")
	}

	// Verify timestamp format (ISO 8601)
	timestamp, ok := result["timestamp"].(string)
	if !ok || !strings.Contains(timestamp, "T") {
		t.Errorf("Invalid timestamp format: %v", timestamp)
	}
}
