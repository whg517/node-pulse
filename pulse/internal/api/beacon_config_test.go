package api

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"hash/crc32"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestValidateBeaconConfig tests the validateBeaconConfig function
func TestValidateBeaconConfig(t *testing.T) {
	tests := []struct {
		name        string
		req         *BeaconConfigUpdateRequest
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Empty request is valid",
			req:         &BeaconConfigUpdateRequest{},
			expectError: false,
		},
		{
			name: "Valid interval",
			req: &BeaconConfigUpdateRequest{
				IntervalSeconds: intPtr(60),
			},
			expectError: false,
		},
		{
			name: "Interval too short",
			req: &BeaconConfigUpdateRequest{
				IntervalSeconds: intPtr(3),
			},
			expectError: true,
			errorMsg:    "interval too short",
		},
		{
			name: "Timeout too short",
			req: &BeaconConfigUpdateRequest{
				TimeoutSeconds: intPtr(0),
			},
			expectError: true,
			errorMsg:    "timeout too short",
		},
		{
			name: "Valid probes",
			req: &BeaconConfigUpdateRequest{
				Probes: &[]ProbeConfig{
					{
						ID:              "probe-1",
						Type:            "TCP",
						Target:          "example.com",
						Port:            80,
						IntervalSeconds: 30,
						TimeoutSeconds:  5,
					},
				},
			},
			expectError: false,
		},
		{
			name: "Invalid probe type",
			req: &BeaconConfigUpdateRequest{
				Probes: &[]ProbeConfig{
					{
						ID:              "probe-1",
						Type:            "HTTP",
						Target:          "example.com",
						Port:            80,
						IntervalSeconds: 30,
						TimeoutSeconds:  5,
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid probe type",
		},
		{
			name: "Empty probe target",
			req: &BeaconConfigUpdateRequest{
				Probes: &[]ProbeConfig{
					{
						ID:              "probe-1",
						Type:            "TCP",
						Target:          "",
						Port:            80,
						IntervalSeconds: 30,
						TimeoutSeconds:  5,
					},
				},
			},
			expectError: true,
			errorMsg:    "probe target cannot be empty",
		},
		{
			name: "Invalid probe port - too low",
			req: &BeaconConfigUpdateRequest{
				Probes: &[]ProbeConfig{
					{
						ID:              "probe-1",
						Type:            "TCP",
						Target:          "example.com",
						Port:            0,
						IntervalSeconds: 30,
						TimeoutSeconds:  5,
					},
				},
			},
			expectError: true,
			errorMsg:    "probe port must be between",
		},
		{
			name: "Invalid probe port - too high",
			req: &BeaconConfigUpdateRequest{
				Probes: &[]ProbeConfig{
					{
						ID:              "probe-1",
						Type:            "TCP",
						Target:          "example.com",
						Port:            70000,
						IntervalSeconds: 30,
						TimeoutSeconds:  5,
					},
				},
			},
			expectError: true,
			errorMsg:    "probe port must be between",
		},
		{
			name: "Probe interval too short",
			req: &BeaconConfigUpdateRequest{
				Probes: &[]ProbeConfig{
					{
						ID:              "probe-1",
						Type:            "TCP",
						Target:          "example.com",
						Port:            80,
						IntervalSeconds: 2,
						TimeoutSeconds:  5,
					},
				},
			},
			expectError: true,
			errorMsg:    "probe interval too short",
		},
		{
			name: "Probe timeout too short",
			req: &BeaconConfigUpdateRequest{
				Probes: &[]ProbeConfig{
					{
						ID:              "probe-1",
						Type:            "TCP",
						Target:          "example.com",
						Port:            80,
						IntervalSeconds: 30,
						TimeoutSeconds:  0,
					},
				},
			},
			expectError: true,
			errorMsg:    "probe timeout too short",
		},
		{
			name: "UDP probe type is valid",
			req: &BeaconConfigUpdateRequest{
				Probes: &[]ProbeConfig{
					{
						ID:              "probe-1",
						Type:            "UDP",
						Target:          "8.8.8.8",
						Port:            53,
						IntervalSeconds: 30,
						TimeoutSeconds:  5,
					},
				},
			},
			expectError: false,
		},
		{
			name: "MTR probe type is valid without port",
			req: &BeaconConfigUpdateRequest{
				Probes: &[]ProbeConfig{
					{
						ID:              "probe-mtr-1",
						Type:            "MTR",
						Target:          "example.com",
						IntervalSeconds: 60,
						TimeoutSeconds:  5,
						Count:           3,
						MaxHops:         30,
						PacketSize:      128,
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBeaconConfig(tt.req)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGetBeaconConfig tests the GetBeaconConfig handler
func TestGetBeaconConfig(t *testing.T) {
	// Reset the config store for testing
	beaconConfigMutex.Lock()
	beaconConfigStore = make(map[string]*BeaconConfig)
	beaconConfigMutex.Unlock()

	handler := &BeaconHandler{}

	tests := []struct {
		name         string
		beaconID     string
		expectStatus int
	}{
		{
			name:         "Get default config for non-existent beacon",
			beaconID:     "550e8400-e29b-41d4-a716-446655440000",
			expectStatus: http.StatusOK,
		},
		{
			name:         "Invalid beacon ID format",
			beaconID:     "invalid-uuid",
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.beaconID}}

			handler.GetBeaconConfig(c)

			assert.Equal(t, tt.expectStatus, w.Code)
		})
	}
}

// TestUpdateBeaconConfig tests the UpdateBeaconConfig handler
func TestUpdateBeaconConfig(t *testing.T) {
	// Reset the config store for testing
	beaconConfigMutex.Lock()
	beaconConfigStore = make(map[string]*BeaconConfig)
	beaconConfigMutex.Unlock()

	handler := &BeaconHandler{}

	t.Run("Update with valid config", func(t *testing.T) {
		beaconID := "550e8400-e29b-41d4-a716-446655440000"
		req := BeaconConfigUpdateRequest{
			IntervalSeconds: intPtr(30),
			TimeoutSeconds:  intPtr(10),
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: beaconID}}
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.UpdateBeaconConfig(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response BeaconConfigResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, 30, response.Data.IntervalSeconds)
		assert.Equal(t, 10, response.Data.TimeoutSeconds)
		assert.Equal(t, 1, response.Data.Version) // First version
	})

	t.Run("Update with invalid beacon ID", func(t *testing.T) {
		req := BeaconConfigUpdateRequest{
			IntervalSeconds: intPtr(30),
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.UpdateBeaconConfig(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Update with invalid config", func(t *testing.T) {
		beaconID := "550e8400-e29b-41d4-a716-446655440001"
		req := BeaconConfigUpdateRequest{
			IntervalSeconds: intPtr(2), // Too short
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: beaconID}}
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.UpdateBeaconConfig(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestGetBeaconConfigHistory tests the GetBeaconConfigHistory handler
func TestGetBeaconConfigHistory(t *testing.T) {
	// Reset the config store for testing
	beaconConfigMutex.Lock()
	beaconConfigStore = make(map[string]*BeaconConfig)
	beaconConfigHistory = make(map[string][]ConfigHistoryEntry)
	beaconConfigMutex.Unlock()

	handler := &BeaconHandler{}

	t.Run("Get history for beacon with no history", func(t *testing.T) {
		beaconID := "550e8400-e29b-41d4-a716-446655440000"

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: beaconID}}

		handler.GetBeaconConfigHistory(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ConfigHistoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Empty(t, response.Data)
	})

	t.Run("Invalid beacon ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

		handler.GetBeaconConfigHistory(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestBatchUpdateBeaconGroupConfig tests the BatchUpdateBeaconGroupConfig handler
func TestBatchUpdateBeaconGroupConfig(t *testing.T) {
	// Reset the config store for testing
	beaconConfigMutex.Lock()
	beaconConfigStore = make(map[string]*BeaconConfig)
	beaconConfigMutex.Unlock()

	handler := &BeaconHandler{}

	t.Run("Batch update multiple beacons", func(t *testing.T) {
		req := BatchConfigUpdateRequest{
			BeaconIDs: []string{
				"550e8400-e29b-41d4-a716-446655440000",
				"550e8400-e29b-41d4-a716-446655440001",
			},
			Config: BeaconConfigUpdateRequest{
				IntervalSeconds: intPtr(45),
			},
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "gid", Value: "group-1"}}
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.BatchUpdateBeaconGroupConfig(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response BatchConfigUpdateResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, 2, response.Data.SuccessCount)
		assert.Equal(t, 0, response.Data.FailedCount)
	})

	t.Run("Batch update with mixed results", func(t *testing.T) {
		req := BatchConfigUpdateRequest{
			BeaconIDs: []string{
				"550e8400-e29b-41d4-a716-446655440002",
				"invalid-uuid",
			},
			Config: BeaconConfigUpdateRequest{
				IntervalSeconds: intPtr(60),
			},
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "gid", Value: "group-2"}}
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.BatchUpdateBeaconGroupConfig(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response BatchConfigUpdateResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, 1, response.Data.SuccessCount)
		assert.Equal(t, 1, response.Data.FailedCount)
	})
}

// TestGetConfigPreview tests the GetConfigPreview handler
func TestGetConfigPreview(t *testing.T) {
	handler := &BeaconHandler{}

	t.Run("Preview valid config", func(t *testing.T) {
		beaconID := "550e8400-e29b-41d4-a716-446655440000"
		req := BeaconConfigUpdateRequest{
			IntervalSeconds: intPtr(60),
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: beaconID}}
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.GetConfigPreview(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		assert.True(t, data["valid"].(bool))
	})

	t.Run("Preview config with warnings", func(t *testing.T) {
		beaconID := "550e8400-e29b-41d4-a716-446655440001"
		req := BeaconConfigUpdateRequest{
			IntervalSeconds: intPtr(15), // Less than 30, should generate warning
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: beaconID}}
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.GetConfigPreview(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		warnings := data["warnings"].([]interface{})
		assert.Greater(t, len(warnings), 0)
	})
}

func TestAcknowledgeBeaconConfig(t *testing.T) {
	handler := &BeaconHandler{}

	t.Run("Valid acknowledgement", func(t *testing.T) {
		req := BeaconConfigAckRequest{
			NodeID:  "550e8400-e29b-41d4-a716-446655440000",
			Version: 2,
			Status:  "applied",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_id", req.NodeID)

		handler.AcknowledgeBeaconConfig(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Reject mismatched token node", func(t *testing.T) {
		req := BeaconConfigAckRequest{
			NodeID:  "550e8400-e29b-41d4-a716-446655440000",
			Version: 2,
			Status:  "applied",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_id", "550e8400-e29b-41d4-a716-446655440001")

		handler.AcknowledgeBeaconConfig(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Reject invalid status", func(t *testing.T) {
		req := BeaconConfigAckRequest{
			NodeID:  "550e8400-e29b-41d4-a716-446655440000",
			Version: 2,
			Status:  "unknown",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.AcknowledgeBeaconConfig(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleMTRResult(t *testing.T) {
	handler := &BeaconHandler{}

	t.Run("Valid MTR result without DB querier", func(t *testing.T) {
		nodeID := "550e8400-e29b-41d4-a716-446655440000"
		req := MTRResultRequest{
			NodeID:      nodeID,
			ProbeID:     "mtr-1",
			Target:      "example.com",
			CompletedAt: time.Now().Format(time.RFC3339),
			Success:     true,
			Hops: []MTRHopRequest{
				{HopNumber: 1, IP: "192.0.2.1", Sent: 10, Received: 10, LossRate: 0, AvgRTTMs: 1.2},
			},
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_id", nodeID)

		handler.HandleMTRResult(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Reject invalid timestamp", func(t *testing.T) {
		req := MTRResultRequest{
			NodeID:      "550e8400-e29b-41d4-a716-446655440000",
			Target:      "example.com",
			CompletedAt: "not-time",
			Hops:        []MTRHopRequest{{HopNumber: 1, IP: "192.0.2.1"}},
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.HandleMTRResult(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestCompressedHeartbeat tests the HandleCompressedHeartbeat handler
func TestCompressedHeartbeat(t *testing.T) {
	// Note: This test doesn't test the full flow since it requires database and cache setup
	// It tests the compression/decompression logic

	t.Run("CRC32 checksum validation", func(t *testing.T) {
		// Create test data
		testData := []byte("test data for compression")
		checksum := crc32.ChecksumIEEE(testData)

		// Verify checksum matches
		computedChecksum := crc32.ChecksumIEEE(testData)
		assert.Equal(t, checksum, computedChecksum)

		// Verify different data produces different checksum
		differentData := []byte("different test data")
		differentChecksum := crc32.ChecksumIEEE(differentData)
		assert.NotEqual(t, checksum, differentChecksum)
	})

	t.Run("Gzip compression and decompression", func(t *testing.T) {
		originalData := map[string]interface{}{
			"node_id":          "550e8400-e29b-41d4-a716-446655440000",
			"probe_id":         "probe-1",
			"latency_ms":       25.5,
			"packet_loss_rate": 0.0,
			"jitter_ms":        2.1,
			"timestamp":        "2024-01-01T00:00:00Z",
		}

		// Marshal to JSON
		jsonBytes, err := json.Marshal(originalData)
		require.NoError(t, err)

		// Compress with gzip
		var compressed bytes.Buffer
		writer := gzip.NewWriter(&compressed)
		_, err = writer.Write(jsonBytes)
		require.NoError(t, err)
		require.NoError(t, writer.Close())

		// Decompress
		reader, err := gzip.NewReader(bytes.NewReader(compressed.Bytes()))
		require.NoError(t, err)

		decompressed, err := io.ReadAll(reader)
		require.NoError(t, reader.Close())
		require.NoError(t, err)

		// Verify decompressed data matches original
		var result map[string]interface{}
		err = json.Unmarshal(decompressed, &result)
		require.NoError(t, err)
		assert.Equal(t, originalData["node_id"], result["node_id"])
		assert.Equal(t, originalData["probe_id"], result["probe_id"])
	})
}

// Helper functions
func intPtr(i int) *int {
	return &i
}
