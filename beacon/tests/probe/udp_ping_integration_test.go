package probe

import (
	"fmt"
	"net"
	"testing"
	"time"

	"beacon/internal/config"
	"beacon/internal/probe"
)

// TestUDPProbeIntegration tests the complete UDP probe workflow
func TestUDPProbeIntegration(t *testing.T) {
	// Start a simple UDP echo server for testing
	serverAddr, ready, err := startUDPEchoServer()
	if err != nil {
		t.Fatalf("Failed to start UDP test server: %v", err)
	}
	defer func() {
		// Server will auto-close after test duration
	}()

	// Wait for the server to be ready
	<-ready

	t.Run("UDP probe to echo server", func(t *testing.T) {
		host, port, _ := net.SplitHostPort(serverAddr)
		portNum := 0
		fmt.Sscanf(port, "%d", &portNum)

		udpConfig := probe.UDPProbeConfig{
			Type:           "udp_ping",
			Target:         host,
			Port:           portNum,
			TimeoutSeconds: 5,
			Interval:       60,
			Count:          10,
		}

		pinger := probe.NewUDPPinger(udpConfig)
		result, err := pinger.ExecuteBatch(10)

		if err != nil {
			t.Errorf("ExecuteBatch() error = %v", err)
		}

		if result == nil {
			t.Fatal("ExecuteBatch() result should not be nil")
		}

		// Check that we sent packets
		if result.SentPackets != 10 {
			t.Errorf("SentPackets = %v, want 10", result.SentPackets)
		}

		// With a real server, we should have received some packets
		if result.ReceivedPackets == 0 {
			t.Logf("WARNING: Received 0 packets from echo server - server may not be responding")
		}

		// Success should be true if we received at least one packet
		expectedSuccess := result.ReceivedPackets > 0
		if result.Success != expectedSuccess {
			t.Errorf("Success = %v, want %v (ReceivedPackets=%d)", result.Success, expectedSuccess, result.ReceivedPackets)
		}

		t.Logf("UDP probe result: sent=%d, received=%d, packet_loss=%.2f%%, avg_rtt=%.2fms, success=%v",
			result.SentPackets, result.ReceivedPackets, result.PacketLossRate, result.RTTMs, result.Success)
	})

	t.Run("UDP probe to invalid port", func(t *testing.T) {
		udpConfig := probe.UDPProbeConfig{
			Type:           "udp_ping",
			Target:         "127.0.0.1",
			Port:           9999, // Likely no server on this port
			TimeoutSeconds: 1,
			Interval:       60,
			Count:          10,
		}

		pinger := probe.NewUDPPinger(udpConfig)
		result, err := pinger.ExecuteBatch(10)

		if err != nil {
			t.Errorf("ExecuteBatch() error = %v", err)
		}

		if result == nil {
			t.Fatal("ExecuteBatch() result should not be nil")
		}

		if result.SentPackets != 10 {
			t.Errorf("SentPackets = %v, want 10", result.SentPackets)
		}

		// Should have 100% packet loss (no response)
		if result.PacketLossRate != 100.0 {
			t.Errorf("PacketLossRate = %v, want 100.0", result.PacketLossRate)
		}

		if result.Success != false {
			t.Errorf("Success = %v, want false", result.Success)
		}

		t.Logf("UDP probe to closed port: packet_loss=%.2f%%, error=%s", result.PacketLossRate, result.ErrorMessage)
	})
}

// TestSchedulerWithUDPProbes tests the scheduler with UDP probe configuration
func TestSchedulerWithUDPProbes(t *testing.T) {
	probeConfigs := []config.ProbeConfig{
		{
			Type:     "udp_ping",
			Target:   "127.0.0.1",
			Port:     12345,
			Interval: 60,
			Count:    10,
			TimeoutSeconds:  2,
		},
		{
			Type:     "udp_ping",
			Target:   "example.com",
			Port:     123,
			Interval: 60,
			Count:    10,
			TimeoutSeconds:  5,
		},
	}

	scheduler, err := probe.NewProbeScheduler(probeConfigs)
	if err != nil {
		t.Fatalf("NewProbeScheduler() error = %v", err)
	}

	if scheduler.GetProbeCount() != 2 {
		t.Errorf("GetProbeCount() = %v, want 2", scheduler.GetProbeCount())
	}

	// Test starting and stopping scheduler
	err = scheduler.Start()
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}

	// Wait a bit for probes to execute
	time.Sleep(3 * time.Second)

	scheduler.Stop()

	if scheduler.IsRunning() {
		t.Error("IsRunning() = true, want false after Stop()")
	}
}

// TestMixedProbeScheduler tests scheduler with both TCP and UDP probes
func TestMixedProbeScheduler(t *testing.T) {
	probeConfigs := []config.ProbeConfig{
		{
			Type:     "tcp_ping",
			Target:   "127.0.0.1",
			Port:     8080,
			Interval: 60,
			Count:    10,
			TimeoutSeconds:  2,
		},
		{
			Type:     "udp_ping",
			Target:   "127.0.0.1",
			Port:     8081,
			Interval: 60,
			Count:    10,
			TimeoutSeconds:  2,
		},
		{
			Type:     "tcp_ping",
			Target:   "example.com",
			Port:     80,
			Interval: 60,
			Count:    10,
			TimeoutSeconds:  3,
		},
	}

	scheduler, err := probe.NewProbeScheduler(probeConfigs)
	if err != nil {
		t.Fatalf("NewProbeScheduler() error = %v", err)
	}

	if scheduler.GetProbeCount() != 3 {
		t.Errorf("GetProbeCount() = %v, want 3 (2 TCP + 1 UDP)", scheduler.GetProbeCount())
	}

	// Verify scheduler started successfully
	err = scheduler.Start()
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}

	// Wait for probes to execute
	time.Sleep(2 * time.Second)

	// Stop scheduler
	scheduler.Stop()
}

// startUDPEchoServer starts a simple UDP echo server for testing
// Returns the server address (host:port) and a readiness channel that closes when the server is ready
func startUDPEchoServer() (string, <-chan struct{}, error) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return "", nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return "", nil, err
	}

	ready := make(chan struct{})

	// Start echo handler in background
	go func() {
		defer conn.Close()
		buf := make([]byte, 1024)

		// Signal that the server is ready to receive packets
		close(ready)

		for {
			n, clientAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				return // Server closed or error
			}

			// Echo the data back
			conn.WriteToUDP(buf[:n], clientAddr)
		}
	}()

	// Return the actual bound address and readiness channel
	return conn.LocalAddr().String(), ready, nil
}
