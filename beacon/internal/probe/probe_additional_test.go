package probe

import (
"strings"
"testing"
)

// TestTCPProbeConfig_Validate_LongHostname tests hostname validation with > 253 chars
func TestTCPProbeConfig_Validate_LongHostname(t *testing.T) {
// Create hostname longer than 253 characters
longHostname := strings.Repeat("a", 254) + ".com"
cfg := TCPProbeConfig{
Type:           "tcp_ping",
Target:         longHostname,
Port:           80,
TimeoutSeconds: 5,
Interval:       60,
Count:          10,
}

if err := cfg.Validate(); err == nil {
t.Error("Expected error for hostname > 253 chars")
}
}

// TestValidateHostname_Empty tests validateHostname with empty string
func TestValidateHostname_Empty(t *testing.T) {
if err := validateHostname(""); err == nil {
t.Error("Expected error for empty hostname")
}
}

// TestValidateHostname_TooLong tests validateHostname with too long hostname
func TestValidateHostname_TooLong(t *testing.T) {
longHostname := strings.Repeat("a", 254)
if err := validateHostname(longHostname); err == nil {
t.Error("Expected error for hostname > 253 chars")
}
}

// TestTCPProbeConfig_Validate_InvalidType tests invalid type
func TestTCPProbeConfig_Validate_InvalidType(t *testing.T) {
cfg := TCPProbeConfig{
Type:           "wrong_type",
Target:         "localhost",
Port:           80,
TimeoutSeconds: 5,
Interval:       60,
Count:          10,
}

if err := cfg.Validate(); err == nil {
t.Error("Expected error for invalid type")
}
}

// TestTCPProbeConfig_Validate_EmptyTarget tests empty target
func TestTCPProbeConfig_Validate_EmptyTarget(t *testing.T) {
cfg := TCPProbeConfig{
Type:           "tcp_ping",
Target:         "",
Port:           80,
TimeoutSeconds: 5,
Interval:       60,
Count:          10,
}

if err := cfg.Validate(); err == nil {
t.Error("Expected error for empty target")
}
}

// TestTCPProbeConfig_Validate_InvalidCountBounds tests count bounds
func TestTCPProbeConfig_Validate_InvalidCountBounds(t *testing.T) {
// Count 0
cfg := TCPProbeConfig{
Type: "tcp_ping", Target: "localhost", Port: 80,
TimeoutSeconds: 5, Interval: 60, Count: 0,
}
if err := cfg.Validate(); err == nil {
t.Error("Expected error for count 0")
}

// Count 101
cfg.Count = 101
if err := cfg.Validate(); err == nil {
t.Error("Expected error for count 101")
}
}

// TestTCPProbeConfig_Validate_InvalidIntervalBounds tests interval bounds
func TestTCPProbeConfig_Validate_InvalidIntervalBounds(t *testing.T) {
// Interval too low
cfg := TCPProbeConfig{
Type: "tcp_ping", Target: "localhost", Port: 80,
TimeoutSeconds: 5, Interval: 59, Count: 10,
}
if err := cfg.Validate(); err == nil {
t.Error("Expected error for interval < 60")
}

// Interval too high
cfg.Interval = 301
if err := cfg.Validate(); err == nil {
t.Error("Expected error for interval > 300")
}
}

// TestTCPPinger_ExecuteBatch_InvalidCount tests ExecuteBatch with invalid count
func TestTCPPinger_ExecuteBatch_InvalidCount(t *testing.T) {
cfg := TCPProbeConfig{
Type:           "tcp_ping",
Target:         "localhost",
Port:           18891,
TimeoutSeconds: 5,
Interval:       60,
Count:          1,
}

pinger := NewTCPPinger(cfg)

// Count 0
_, err := pinger.ExecuteBatch(0)
if err == nil {
t.Error("Expected error for count 0")
}

// Count 101
_, err = pinger.ExecuteBatch(101)
if err == nil {
t.Error("Expected error for count 101")
}
}

// TestTCPPinger_ExecuteBatch_AllFailed tests ExecuteBatch when all connections fail
func TestTCPPinger_ExecuteBatch_AllFailed(t *testing.T) {
// Use a port that's definitely not listening
cfg := TCPProbeConfig{
Type:           "tcp_ping",
Target:         "localhost",
Port:           19998, // Definitely not listening
TimeoutSeconds: 1,
Interval:       60,
Count:          1,
}

pinger := NewTCPPinger(cfg)
result, err := pinger.ExecuteBatch(3)
if err != nil {
t.Fatalf("ExecuteBatch should not error when connections fail: %v", err)
}
if result == nil {
t.Fatal("Expected non-nil result even when connections fail")
}
if result.Success {
t.Error("Expected Success=false when all connections fail")
}
if result.PacketLossRate != 100.0 {
t.Errorf("Expected 100%% packet loss, got %f", result.PacketLossRate)
}
}

// TestUDPProbeConfig_Validate_InvalidType tests UDP validation for invalid type
func TestUDPProbeConfig_Validate_InvalidType(t *testing.T) {
cfg := UDPProbeConfig{
Type:           "wrong_type",
Target:         "localhost",
Port:           53,
TimeoutSeconds: 5,
Interval:       60,
Count:          10,
}
if err := cfg.Validate(); err == nil {
t.Error("Expected error for invalid UDP type")
}
}

// TestUDPPinger_Execute_FailedConnection tests Execute when connection fails
func TestUDPPinger_Execute_FailedConnection(t *testing.T) {
cfg := UDPProbeConfig{
Type:           "udp_ping",
Target:         "192.0.2.1", // RFC 5737 test address (unreachable)
Port:           12345,
TimeoutSeconds: 1,
Interval:       60,
Count:          10,
}

pinger := NewUDPPinger(cfg)
result, err := pinger.Execute()
if err != nil {
t.Fatalf("Execute should not error: %v", err)
}
if result == nil {
t.Fatal("Expected non-nil result")
}
// Should be failed/timeout
}

// TestUDPPinger_ExecuteBatch_AllFailed tests ExecuteBatch when all fail
func TestUDPPinger_ExecuteBatch_AllFailed(t *testing.T) {
cfg := UDPProbeConfig{
Type:           "udp_ping",
Target:         "192.0.2.1", // Unreachable
Port:           12345,
TimeoutSeconds: 1,
Interval:       60,
Count:          10,
}

pinger := NewUDPPinger(cfg)
result, err := pinger.ExecuteBatch(5)
if err != nil {
t.Fatalf("ExecuteBatch should not error: %v", err)
}
if result == nil {
t.Fatal("Expected non-nil result")
}
}

// TestUDPPinger_ExecuteBatch_InvalidCount tests ExecuteBatch with invalid count
func TestUDPPinger_ExecuteBatch_InvalidCount(t *testing.T) {
cfg := UDPProbeConfig{
Type:           "udp_ping",
Target:         "localhost",
Port:           53,
TimeoutSeconds: 1,
Interval:       60,
Count:          10,
}

pinger := NewUDPPinger(cfg)

// Count 0
_, err := pinger.ExecuteBatch(0)
if err == nil {
t.Error("Expected error for count 0")
}

// Count 101
_, err = pinger.ExecuteBatch(101)
if err == nil {
t.Error("Expected error for count 101")
}
}

// TestICMPPinger_Execute_NoRoot tests Execute without root (ICMP fails gracefully)
func TestICMPPinger_Execute_NoRoot(t *testing.T) {
cfg := ICMPProbeConfig{
Type:           "icmp_ping",
Target:         "127.0.0.1",
TimeoutSeconds: 1,
Interval:       60,
Count:          1,
}

pinger := NewICMPPinger(cfg)
result, err := pinger.Execute()

// Should not return an error - ICMP failures return a result with Success=false
if err != nil {
t.Fatalf("Execute should not return an error: %v", err)
}
if result == nil {
t.Fatal("Expected non-nil result")
}
t.Logf("ICMP Execute result: success=%v, error=%s", result.Success, result.ErrorMessage)
}

// TestICMPPinger_Execute_IPv6Target tests Execute with IPv6 target
func TestICMPPinger_Execute_IPv6Target(t *testing.T) {
cfg := ICMPProbeConfig{
Type:           "icmp_ping",
Target:         "::1",
TimeoutSeconds: 1,
Interval:       60,
Count:          1,
}

pinger := NewICMPPinger(cfg)
result, err := pinger.Execute()

if err != nil {
t.Fatalf("Execute should not return an error: %v", err)
}
if result == nil {
t.Fatal("Expected non-nil result")
}
t.Logf("IPv6 ICMP Execute result: success=%v, error=%s", result.Success, result.ErrorMessage)
}

// TestICMPPinger_Execute_InvalidConfig_Additional tests Execute with invalid config (additional)
func TestICMPPinger_Execute_InvalidConfig_Additional(t *testing.T) {
cfg := ICMPProbeConfig{
Type:           "wrong_type",
Target:         "127.0.0.1",
TimeoutSeconds: 1,
Interval:       60,
Count:          1,
}

pinger := NewICMPPinger(cfg)
_, err := pinger.Execute()

if err == nil {
t.Error("Expected error for invalid config type")
}
}

// TestICMPPinger_ExecuteBatch_NoRoot tests ExecuteBatch without root
func TestICMPPinger_ExecuteBatch_NoRoot(t *testing.T) {
cfg := ICMPProbeConfig{
Type:           "icmp_ping",
Target:         "127.0.0.1",
TimeoutSeconds: 1,
Interval:       60,
Count:          5,
}

pinger := NewICMPPinger(cfg)
result, err := pinger.ExecuteBatch(5)

if err != nil {
t.Fatalf("ExecuteBatch should not return an error: %v", err)
}
if result == nil {
t.Fatal("Expected non-nil result")
}
t.Logf("ICMP ExecuteBatch result: success=%v, error=%s", result.Success, result.ErrorMessage)
}

// TestICMPPinger_ExecuteBatch_InvalidCount_Additional tests invalid count (additional)
func TestICMPPinger_ExecuteBatch_InvalidCount_Additional(t *testing.T) {
cfg := ICMPProbeConfig{
Type:           "icmp_ping",
Target:         "127.0.0.1",
TimeoutSeconds: 1,
Interval:       60,
Count:          1,
}

pinger := NewICMPPinger(cfg)

// Count 0
_, err := pinger.ExecuteBatch(0)
if err == nil {
t.Error("Expected error for count 0")
}

// Count 101
_, err = pinger.ExecuteBatch(101)
if err == nil {
t.Error("Expected error for count 101")
}
}

// TestICMPPinger_ExecuteBatch_InvalidTarget tests ExecuteBatch with unresolvable target
func TestICMPPinger_ExecuteBatch_InvalidTarget(t *testing.T) {
cfg := ICMPProbeConfig{
Type:           "icmp_ping",
Target:         "this.is.not.a.valid.hostname.xyz.abc.def",
TimeoutSeconds: 1,
Interval:       60,
Count:          1,
}

pinger := NewICMPPinger(cfg)
result, err := pinger.ExecuteBatch(1)

if err != nil {
t.Fatalf("ExecuteBatch should not return an error: %v", err)
}
if result == nil {
t.Fatal("Expected non-nil result")
}
t.Logf("Invalid target result: success=%v, error=%s", result.Success, result.ErrorMessage)
}
