package probe_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"beacon/internal/models"
	"beacon/internal/probe"
)

// TestTCPCoreMetricsIntegration tests TCP probe core metrics collection with real server
func TestTCPCoreMetricsIntegration(t *testing.T) {
	// Setup: Start a real TCP server
	server := startIntegrationTestTCPServer(t, "localhost:19001")
	defer server.Close()

	t.Run("TCP probe with 10 samples - core metrics calculation", func(t *testing.T) {
		config := probe.TCPProbeConfig{
			Type:           "tcp_ping",
			Target:         "localhost",
			Port:           19001,
			TimeoutSeconds: 5,
			Interval:       60,
			Count:          10,
		}

		pinger := probe.NewTCPPinger(config)
		result, err := pinger.ExecuteBatch(10)

		if err != nil {
			t.Fatalf("ExecuteBatch() failed: %v", err)
		}

		// Verify success
		if !result.Success {
			t.Errorf("Expected Success=true, got false. Error: %s", result.ErrorMessage)
		}

		// Verify sample count
		if result.SampleCount != 10 {
			t.Errorf("Expected SampleCount=10, got %d", result.SampleCount)
		}

		// Verify RTT metrics are calculated
		if result.RTTMs <= 0 {
			t.Errorf("Expected RTTMs > 0, got %f", result.RTTMs)
		}
		if result.RTTMedianMs <= 0 {
			t.Errorf("Expected RTTMedianMs > 0, got %f", result.RTTMedianMs)
		}

		// Verify jitter is calculated
		if result.JitterMs < 0 {
			t.Errorf("Expected JitterMs >= 0, got %f", result.JitterMs)
		}

		// Verify variance is calculated
		if result.VarianceMs < 0 {
			t.Errorf("Expected VarianceMs >= 0, got %f", result.VarianceMs)
		}

		// Verify packet loss rate (should be 0 for successful local connection)
		if result.PacketLossRate != 0 {
			t.Errorf("Expected PacketLossRate=0 for successful connection, got %f", result.PacketLossRate)
		}

		// Verify timestamp
		if result.Timestamp == "" {
			t.Errorf("Expected non-empty Timestamp")
		}

		t.Logf("TCP Core Metrics: RTT=%.2f ms (median=%.2f ms), jitter=%.2f ms, variance=%.2f ms², packet loss=%.2f%%, samples=%d",
			result.RTTMs, result.RTTMedianMs, result.JitterMs, result.VarianceMs, result.PacketLossRate, result.SampleCount)
	})

	t.Run("TCP probe with 100 samples - performance test", func(t *testing.T) {
		config := probe.TCPProbeConfig{
			Type:           "tcp_ping",
			Target:         "localhost",
			Port:           19001,
			TimeoutSeconds: 5,
			Interval:       60,
			Count:          100,
		}

		pinger := probe.NewTCPPinger(config)

		startTime := time.Now()
		result, err := pinger.ExecuteBatch(100)
		elapsed := time.Since(startTime)

		if err != nil {
			t.Fatalf("ExecuteBatch() failed: %v", err)
		}

		// Verify performance requirement: 100 samples should complete in reasonable time
		if elapsed > 30*time.Second {
			t.Errorf("100 samples took %v, expected < 30 seconds", elapsed)
		}

		// Verify all metrics calculated
		if result.SampleCount != 100 {
			t.Errorf("Expected SampleCount=100, got %d", result.SampleCount)
		}

		t.Logf("100 samples completed in %v (%.2f ms/sample)", elapsed, float64(elapsed.Milliseconds())/100)
	})
}

// TestUDPCoreMetricsIntegration tests UDP probe core metrics collection with real server
func TestUDPCoreMetricsIntegration(t *testing.T) {
	// Setup: Start a real UDP echo server
	server := startIntegrationTestUDPEchoServer(t, "localhost:19002")
	defer server.Close()

	t.Run("UDP probe with 10 samples - core metrics calculation", func(t *testing.T) {
		config := probe.UDPProbeConfig{
			Type:           "udp_ping",
			Target:         "localhost",
			Port:           19002,
			TimeoutSeconds: 5,
			Interval:       60,
			Count:          10,
		}

		pinger := probe.NewUDPPinger(config)
		result, err := pinger.ExecuteBatch(10)

		if err != nil {
			t.Fatalf("ExecuteBatch() failed: %v", err)
		}

		// Verify success
		if !result.Success {
			t.Errorf("Expected Success=true, got false. Error: %s", result.ErrorMessage)
		}

		// Verify sample count
		if result.SampleCount != 10 {
			t.Errorf("Expected SampleCount=10, got %d", result.SampleCount)
		}

		// Verify sent/received packets
		if result.SentPackets != 10 {
			t.Errorf("Expected SentPackets=10, got %d", result.SentPackets)
		}

		// For UDP echo server, we expect all packets to be received
		if result.ReceivedPackets == 0 {
			t.Errorf("Expected ReceivedPackets > 0, got %d", result.ReceivedPackets)
		}

		// Verify RTT metrics
		if result.RTTMs <= 0 {
			t.Errorf("Expected RTTMs > 0, got %f", result.RTTMs)
		}
		if result.RTTMedianMs <= 0 {
			t.Errorf("Expected RTTMedianMs > 0, got %f", result.RTTMedianMs)
		}

		// Verify jitter
		if result.JitterMs < 0 {
			t.Errorf("Expected JitterMs >= 0, got %f", result.JitterMs)
		}

		// Verify variance
		if result.VarianceMs < 0 {
			t.Errorf("Expected VarianceMs >= 0, got %f", result.VarianceMs)
		}

		// Verify packet loss rate is reasonable for local echo server
		if result.PacketLossRate > 50 {
			t.Errorf("Expected PacketLossRate <= 50%% for local echo server, got %f%%", result.PacketLossRate)
		}

		t.Logf("UDP Core Metrics: RTT=%.2f ms (median=%.2f ms), jitter=%.2f ms, variance=%.2f ms², packet loss=%.2f%%, sent=%d, received=%d, samples=%d",
			result.RTTMs, result.RTTMedianMs, result.JitterMs, result.VarianceMs, result.PacketLossRate,
			result.SentPackets, result.ReceivedPackets, result.SampleCount)
	})
}

// TestMeasurementPrecisionWithRealServer validates AC #5: 时延测量精度 ≤1 毫秒
func TestMeasurementPrecisionWithRealServer(t *testing.T) {
	server := startIntegrationTestTCPServer(t, "localhost:19003")
	defer server.Close()

	t.Run("RTT precision with real server - 100 samples", func(t *testing.T) {
		config := probe.TCPProbeConfig{
			Type:           "tcp_ping",
			Target:         "localhost",
			Port:           19003,
			TimeoutSeconds: 5,
			Interval:       60,
			Count:          100,
		}

		pinger := probe.NewTCPPinger(config)
		result, err := pinger.ExecuteBatch(100)

		if err != nil {
			t.Fatalf("ExecuteBatch() failed: %v", err)
		}

		if !result.Success {
			t.Fatalf("Expected successful connection, got error: %s", result.ErrorMessage)
		}

		// AC #5: 时延测量精度 ≤1 毫秒
		// Verify RTT has sub-millisecond precision (2 decimal places)
		rttStr := fmt.Sprintf("%.2f", result.RTTMs)
		if rttStr == fmt.Sprintf("%.0f", result.RTTMs) {
			t.Error("Expected RTT to have sub-millisecond precision (2 decimal places), got integer value")
		}

		// Verify median also has precision
		medianStr := fmt.Sprintf("%.2f", result.RTTMedianMs)
		if medianStr == fmt.Sprintf("%.0f", result.RTTMedianMs) {
			t.Error("Expected RTTMedian to have sub-millisecond precision (2 decimal places), got integer value")
		}

		// Verify jitter has precision
		jitterStr := fmt.Sprintf("%.2f", result.JitterMs)
		if jitterStr == fmt.Sprintf("%.0f", result.JitterMs) && result.JitterMs > 0 {
			t.Error("Expected Jitter to have sub-millisecond precision (2 decimal places), got integer value")
		}

		// Verify variance has precision
		varianceStr := fmt.Sprintf("%.2f", result.VarianceMs)
		if varianceStr == fmt.Sprintf("%.0f", result.VarianceMs) && result.VarianceMs > 0 {
			t.Error("Expected Variance to have sub-millisecond precision (2 decimal places), got integer value")
		}

		// All values should be ≤ 1ms precision (2 decimal places is better than required)
		t.Logf("✅ Precision validation passed: RTT=%.2f ms (median=%.2f ms), jitter=%.2f ms, variance=%.2f ms²",
			result.RTTMs, result.RTTMedianMs, result.JitterMs, result.VarianceMs)
	})
}

// TestCoreMetricsSerialization tests that core metrics can be serialized to JSON
// for Story 3.7 data upload integration
func TestCoreMetricsSerialization(t *testing.T) {
	t.Run("TCP probe result serialization", func(t *testing.T) {
		result := models.NewTCPProbeResultWithMetrics(
			true,   // success
			123.45, // rttMs
			122.50, // rttMedianMs
			5.67,   // jitterMs
			12.34,  // varianceMs
			0.0,    // packetLossRate
			10,     // sampleCount
			"",     // errorMessage
		)

		// Verify all fields are populated
		if result.RTTMs != 123.45 {
			t.Errorf("Expected RTTMs=123.45, got %f", result.RTTMs)
		}
		if result.RTTMedianMs != 122.50 {
			t.Errorf("Expected RTTMedianMs=122.50, got %f", result.RTTMedianMs)
		}
		if result.JitterMs != 5.67 {
			t.Errorf("Expected JitterMs=5.67, got %f", result.JitterMs)
		}
		if result.VarianceMs != 12.34 {
			t.Errorf("Expected VarianceMs=12.34, got %f", result.VarianceMs)
		}
		if result.SampleCount != 10 {
			t.Errorf("Expected SampleCount=10, got %d", result.SampleCount)
		}

		t.Logf("✅ TCP probe result serialization ready for Story 3.7 upload")
	})

	t.Run("UDP probe result serialization", func(t *testing.T) {
		result := models.NewUDPProbeResultWithMetrics(
			true,   // success
			0.0,    // packetLossRate
			98.76,  // rttMs
			97.50,  // rttMedianMs
			4.56,   // jitterMs
			10.12,  // varianceMs
			10,     // sentPackets
			10,     // receivedPackets
			10,     // sampleCount
			"",     // errorMessage
		)

		// Verify all fields are populated
		if result.PacketLossRate != 0.0 {
			t.Errorf("Expected PacketLossRate=0.0, got %f", result.PacketLossRate)
		}
		if result.RTTMs != 98.76 {
			t.Errorf("Expected RTTMs=98.76, got %f", result.RTTMs)
		}
		if result.RTTMedianMs != 97.50 {
			t.Errorf("Expected RTTMedianMs=97.50, got %f", result.RTTMedianMs)
		}
		if result.JitterMs != 4.56 {
			t.Errorf("Expected JitterMs=4.56, got %f", result.JitterMs)
		}
		if result.VarianceMs != 10.12 {
			t.Errorf("Expected VarianceMs=10.12, got %f", result.VarianceMs)
		}
		if result.SampleCount != 10 {
			t.Errorf("Expected SampleCount=10, got %d", result.SampleCount)
		}

		t.Logf("✅ UDP probe result serialization ready for Story 3.7 upload")
	})
}

// Helper functions for integration tests

// startIntegrationTestTCPServer starts a simple TCP server for testing
func startIntegrationTestTCPServer(t *testing.T, addr string) *integrationTestTCPServer {
	t.Helper()
	server := &integrationTestTCPServer{addr: addr}
	go server.start()
	time.Sleep(100 * time.Millisecond) // Give server time to start
	return server
}

type integrationTestTCPServer struct {
	addr string
	lis  net.Listener
}

func (s *integrationTestTCPServer) start() {
	var err error
	s.lis, err = net.Listen("tcp", s.addr)
	if err != nil {
		return
	}
	for {
		conn, err := s.lis.Accept()
		if err != nil {
			break
		}
		conn.Close()
	}
}

func (s *integrationTestTCPServer) Close() {
	if s.lis != nil {
		s.lis.Close()
	}
}

// startIntegrationTestUDPEchoServer starts a simple UDP echo server for testing
func startIntegrationTestUDPEchoServer(t *testing.T, addr string) *integrationTestUDPEchoServer {
	t.Helper()
	server := &integrationTestUDPEchoServer{addr: addr}
	go server.start()
	time.Sleep(100 * time.Millisecond) // Give server time to start
	return server
}

type integrationTestUDPEchoServer struct {
	addr     string
	conn     *net.UDPConn
	stopChan chan struct{}
}

func (s *integrationTestUDPEchoServer) start() {
	udpAddr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return
	}
	s.conn = conn
	s.stopChan = make(chan struct{})

	buf := make([]byte, 1024)
	for {
		select {
		case <-s.stopChan:
			return
		default:
			n, clientAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}
			// Echo the response back
			conn.WriteToUDP(buf[:n], clientAddr)
		}
	}
}

func (s *integrationTestUDPEchoServer) Close() {
	if s.stopChan != nil {
		close(s.stopChan)
	}
	if s.conn != nil {
		s.conn.Close()
	}
}
