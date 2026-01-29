package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/kevin/node-pulse/pulse-api/internal/db"
	"github.com/kevin/node-pulse/pulse-api/internal/models"
)

// MockProbesQuerier is a mock for ProbesQuerier interface
type MockProbesQuerier struct {
	createProbeFunc    func(context.Context, uuid.UUID, uuid.UUID, string, string, int, int, int, int) error
	getProbesFunc      func(context.Context) ([]*models.Probe, error)
	getProbesByNodeFunc func(context.Context, uuid.UUID) ([]*models.Probe, error)
	getProbeByIDFunc   func(context.Context, uuid.UUID) (*models.Probe, error)
	updateProbeFunc    func(context.Context, uuid.UUID, map[string]interface{}) error
	deleteProbeFunc    func(context.Context, uuid.UUID) error
}

func (m *MockProbesQuerier) CreateProbe(ctx context.Context, probeID uuid.UUID, nodeID uuid.UUID, probeType string, target string, port int, intervalSeconds int, count int, timeoutSeconds int) error {
	if m.createProbeFunc != nil {
		return m.createProbeFunc(ctx, probeID, nodeID, probeType, target, port, intervalSeconds, count, timeoutSeconds)
	}
	return nil
}

func (m *MockProbesQuerier) GetProbes(ctx context.Context) ([]*models.Probe, error) {
	if m.getProbesFunc != nil {
		return m.getProbesFunc(ctx)
	}
	return nil, nil
}

func (m *MockProbesQuerier) GetProbesByNode(ctx context.Context, nodeID uuid.UUID) ([]*models.Probe, error) {
	if m.getProbesByNodeFunc != nil {
		return m.getProbesByNodeFunc(ctx, nodeID)
	}
	return nil, nil
}

func (m *MockProbesQuerier) GetProbeByID(ctx context.Context, probeID uuid.UUID) (*models.Probe, error) {
	if m.getProbeByIDFunc != nil {
		return m.getProbeByIDFunc(ctx, probeID)
	}
	return nil, nil
}

func (m *MockProbesQuerier) UpdateProbe(ctx context.Context, probeID uuid.UUID, updates map[string]interface{}) error {
	if m.updateProbeFunc != nil {
		return m.updateProbeFunc(ctx, probeID, updates)
	}
	return nil
}

func (m *MockProbesQuerier) DeleteProbe(ctx context.Context, probeID uuid.UUID) error {
	if m.deleteProbeFunc != nil {
		return m.deleteProbeFunc(ctx, probeID)
	}
	return nil
}

// TestCreateProbeHandler_ValidTCPProbe tests creating a valid TCP probe
func TestCreateProbeHandler_ValidTCPProbe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	nodeID := uuid.New()

	mockNodeQuerier := &MockNodesQuerier{
		getNodeByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Node, error) {
			return &models.Node{
				ID:   nodeID.String(),
				Name: "test-node",
				IP:   "192.168.1.1",
			}, nil
		},
	}

	mockProbeQuerier := &MockProbesQuerier{
		createProbeFunc: func(ctx context.Context, pid uuid.UUID, nid uuid.UUID, probeType string, target string, port int, intervalSeconds int, count int, timeoutSeconds int) error {
			assert.Equal(t, nodeID, nid)
			assert.Equal(t, "TCP", probeType)
			assert.Equal(t, "example.com", target)
			assert.Equal(t, 80, port)
			assert.Equal(t, 60, intervalSeconds)
			assert.Equal(t, 5, count)
			assert.Equal(t, 10, timeoutSeconds)
			return nil
		},
		getProbeByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Probe, error) {
			return &models.Probe{
				ID:              id.String(),
				NodeID:          nodeID.String(),
				Type:            "TCP",
				Target:          "example.com",
				Port:            80,
				IntervalSeconds: 60,
				Count:           5,
				TimeoutSeconds:  10,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}, nil
		},
	}

	handler := NewProbeHandler(mockProbeQuerier, mockNodeQuerier)

	reqBody := models.CreateProbeRequest{
		NodeID:          nodeID.String(),
		Type:            "TCP",
		Target:          "example.com",
		Port:            80,
		IntervalSeconds: 60,
		Count:           5,
		TimeoutSeconds:  10,
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/probes", strings.NewReader(string(body)))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateProbeHandler(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.CreateProbeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "探测配置创建成功", response.Message)
	assert.NotNil(t, response.Data.Probe)
	assert.Equal(t, "TCP", response.Data.Probe.Type)
}

// TestCreateProbeHandler_InvalidProbeType tests validation of invalid probe type
func TestCreateProbeHandler_InvalidProbeType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	nodeID := uuid.New()

	mockNodeQuerier := &MockNodesQuerier{
		getNodeByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Node, error) {
			return &models.Node{
				ID:   nodeID.String(),
				Name: "test-node",
				IP:   "192.168.1.1",
			}, nil
		},
	}

	mockProbeQuerier := &MockProbesQuerier{}
	handler := NewProbeHandler(mockProbeQuerier, mockNodeQuerier)

	reqBody := `{
		"node_id": "` + nodeID.String() + `",
		"type": "ICMP",
		"target": "example.com",
		"port": 80,
		"interval_seconds": 60,
		"count": 5,
		"timeout_seconds": 10
	}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/probes", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateProbeHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, ErrProbeTypeInvalid, response.Code)
	assert.Contains(t, response.Message, "探测类型无效")
}

// TestCreateProbeHandler_InvalidInterval tests interval range validation
func TestCreateProbeHandler_InvalidInterval(t *testing.T) {
	gin.SetMode(gin.TestMode)

	nodeID := uuid.New()
	mockNodeQuerier := &MockNodesQuerier{}
	mockProbeQuerier := &MockProbesQuerier{}
	handler := NewProbeHandler(mockProbeQuerier, mockNodeQuerier)

	reqBody := `{
		"node_id": "` + nodeID.String() + `",
		"type": "TCP",
		"target": "example.com",
		"port": 80,
		"interval_seconds": 30,
		"count": 5,
		"timeout_seconds": 10
	}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/probes", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateProbeHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Message, "请求参数无效")
}

// TestCreateProbeHandler_NodeNotFound tests validation of non-existent node
func TestCreateProbeHandler_NodeNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	nodeID := uuid.New()

	mockNodeQuerier := &MockNodesQuerier{
		getNodeByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Node, error) {
			return nil, db.ErrNodeNotFound
		},
	}

	mockProbeQuerier := &MockProbesQuerier{}
	handler := NewProbeHandler(mockProbeQuerier, mockNodeQuerier)

	reqBody := models.CreateProbeRequest{
		NodeID:          nodeID.String(),
		Type:            "TCP",
		Target:          "example.com",
		Port:            80,
		IntervalSeconds: 60,
		Count:           5,
		TimeoutSeconds:  10,
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/probes", strings.NewReader(string(body)))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateProbeHandler(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, ErrProbeNodeNotFound, response.Code)
}

// TestCreateProbeHandler_InvalidTarget tests validation of invalid target
func TestCreateProbeHandler_InvalidTarget(t *testing.T) {
	gin.SetMode(gin.TestMode)

	nodeID := uuid.New()

	mockNodeQuerier := &MockNodesQuerier{
		getNodeByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Node, error) {
			return &models.Node{
				ID:   nodeID.String(),
				Name: "test-node",
				IP:   "192.168.1.1",
			}, nil
		},
	}

	mockProbeQuerier := &MockProbesQuerier{}
	handler := NewProbeHandler(mockProbeQuerier, mockNodeQuerier)

	reqBody := models.CreateProbeRequest{
		NodeID:          nodeID.String(),
		Type:            "TCP",
		Target:          "-invalid-domain", // Invalid domain format
		Port:            80,
		IntervalSeconds: 60,
		Count:           5,
		TimeoutSeconds:  10,
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/probes", strings.NewReader(string(body)))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateProbeHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, ErrProbeTargetInvalid, response.Code)
}

// TestGetProbesHandler_Success tests getting all probes
func TestGetProbesHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	probeID := uuid.New()
	nodeID := uuid.New()

	mockProbes := []*models.Probe{
		{
			ID:              probeID.String(),
			NodeID:          nodeID.String(),
			Type:            "TCP",
			Target:          "example.com",
			Port:            80,
			IntervalSeconds: 60,
			Count:           5,
			TimeoutSeconds:  10,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	mockNodeQuerier := &MockNodesQuerier{}
	mockProbeQuerier := &MockProbesQuerier{
		getProbesFunc: func(ctx context.Context) ([]*models.Probe, error) {
			return mockProbes, nil
		},
	}

	handler := NewProbeHandler(mockProbeQuerier, mockNodeQuerier)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/probes", nil)

	handler.GetProbesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.GetProbesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "探测配置列表获取成功", response.Message)
	assert.Len(t, response.Data.Probes, 1)
	assert.Equal(t, "TCP", response.Data.Probes[0].Type)
}

// TestGetProbeByIDHandler_Success tests getting a probe by ID
func TestGetProbeByIDHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	probeID := uuid.New()
	nodeID := uuid.New()

	mockNodeQuerier := &MockNodesQuerier{}
	mockProbeQuerier := &MockProbesQuerier{
		getProbeByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Probe, error) {
			return &models.Probe{
				ID:              probeID.String(),
				NodeID:          nodeID.String(),
				Type:            "UDP",
				Target:          "192.168.1.1",
				Port:            53,
				IntervalSeconds: 120,
				Count:           10,
				TimeoutSeconds:  5,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}, nil
		},
	}

	handler := NewProbeHandler(mockProbeQuerier, mockNodeQuerier)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/probes/"+probeID.String(), nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: probeID.String()}}

	handler.GetProbeByIDHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.GetProbeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "探测配置查询成功", response.Message)
	assert.Equal(t, "UDP", response.Data.Probe.Type)
}

// TestGetProbeByIDHandler_NotFound tests getting a non-existent probe
func TestGetProbeByIDHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	probeID := uuid.New()

	mockNodeQuerier := &MockNodesQuerier{}
	mockProbeQuerier := &MockProbesQuerier{
		getProbeByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Probe, error) {
			return nil, db.ErrProbeNotFound
		},
	}

	handler := NewProbeHandler(mockProbeQuerier, mockNodeQuerier)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/probes/"+probeID.String(), nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: probeID.String()}}

	handler.GetProbeByIDHandler(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, ErrProbeNotFound, response.Code)
}

// TestUpdateProbeHandler_Success tests updating a probe
func TestUpdateProbeHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	probeID := uuid.New()
	nodeID := uuid.New()

	newInterval := 120
	newCount := 10

	mockNodeQuerier := &MockNodesQuerier{}
	mockProbeQuerier := &MockProbesQuerier{
		updateProbeFunc: func(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
			return nil
		},
		getProbeByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Probe, error) {
			return &models.Probe{
				ID:              probeID.String(),
				NodeID:          nodeID.String(),
				Type:            "TCP",
				Target:          "example.com",
				Port:            80,
				IntervalSeconds: newInterval,
				Count:           newCount,
				TimeoutSeconds:  10,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}, nil
		},
	}

	handler := NewProbeHandler(mockProbeQuerier, mockNodeQuerier)

	reqBody := models.UpdateProbeRequest{
		IntervalSeconds: &newInterval,
		Count:           &newCount,
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/v1/probes/"+probeID.String(), strings.NewReader(string(body)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: probeID.String()}}

	handler.UpdateProbeHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.UpdateProbeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "探测配置更新成功", response.Message)
	assert.Equal(t, newInterval, response.Data.Probe.IntervalSeconds)
}

// TestDeleteProbeHandler_Success tests deleting a probe
func TestDeleteProbeHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	probeID := uuid.New()

	mockNodeQuerier := &MockNodesQuerier{}
	mockProbeQuerier := &MockProbesQuerier{
		deleteProbeFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}

	handler := NewProbeHandler(mockProbeQuerier, mockNodeQuerier)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/api/v1/probes/"+probeID.String()+"?confirm=true", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: probeID.String()}}

	handler.DeleteProbeHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.DeleteProbeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "探测配置删除成功", response.Message)
}

// TestDeleteProbeHandler_NoConfirmation tests that deletion requires confirmation
func TestDeleteProbeHandler_NoConfirmation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	probeID := uuid.New()

	mockNodeQuerier := &MockNodesQuerier{}
	mockProbeQuerier := &MockProbesQuerier{}

	handler := NewProbeHandler(mockProbeQuerier, mockNodeQuerier)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/api/v1/probes/"+probeID.String(), nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: probeID.String()}}

	handler.DeleteProbeHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ERR_CONFIRMATION_REQUIRED", response.Code)
}

// TestCreateProbeHandler_CaseInsensitiveType tests case-insensitive type validation
func TestCreateProbeHandler_CaseInsensitiveType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	nodeID := uuid.New()
	testCases := []string{"tcp", "Tcp", "TCP", "udp", "Udp", "UDP"}

	for _, testType := range testCases {
		t.Run("type_"+testType, func(t *testing.T) {
			mockNodeQuerier := &MockNodesQuerier{
				getNodeByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Node, error) {
					return &models.Node{
						ID:   nodeID.String(),
						Name: "test-node",
						IP:   "192.168.1.1",
					}, nil
				},
			}

			normalizedType := strings.ToUpper(testType)
			mockProbeQuerier := &MockProbesQuerier{
				createProbeFunc: func(ctx context.Context, pid uuid.UUID, nid uuid.UUID, probeType string, target string, port int, intervalSeconds int, count int, timeoutSeconds int) error {
					assert.Equal(t, normalizedType, probeType, "Type should be normalized to uppercase")
					return nil
				},
				getProbeByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Probe, error) {
					return &models.Probe{
						ID:              id.String(),
						NodeID:          nodeID.String(),
						Type:            normalizedType,
						Target:          "example.com",
						Port:            80,
						IntervalSeconds: 60,
						Count:           5,
						TimeoutSeconds:  10,
						CreatedAt:       time.Now(),
						UpdatedAt:       time.Now(),
					}, nil
				},
			}

			handler := NewProbeHandler(mockProbeQuerier, mockNodeQuerier)

			reqBody := models.CreateProbeRequest{
				NodeID:          nodeID.String(),
				Type:            testType,
				Target:          "example.com",
				Port:            80,
				IntervalSeconds: 60,
				Count:           5,
				TimeoutSeconds:  10,
			}

			body, _ := json.Marshal(reqBody)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/api/v1/probes", strings.NewReader(string(body)))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreateProbeHandler(c)

			assert.Equal(t, http.StatusCreated, w.Code, "Should accept "+testType+" and normalize to "+normalizedType)
		})
	}
}

// TestCreateProbeHandler_PortBoundaryValues tests port boundary validation
func TestCreateProbeHandler_PortBoundaryValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	nodeID := uuid.New()

	testCases := []struct {
		name        string
		port        int
		shouldPass  bool
	}{
		{"port_min_valid", 1, true},
		{"port_max_valid", 65535, true},
		{"port_too_low", 0, false},
		{"port_too_high", 65536, false},
		{"port_negative", -1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockNodeQuerier := &MockNodesQuerier{
				getNodeByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Node, error) {
					return &models.Node{
						ID:   nodeID.String(),
						Name: "test-node",
						IP:   "192.168.1.1",
					}, nil
				},
			}

			mockProbeQuerier := &MockProbesQuerier{}
			handler := NewProbeHandler(mockProbeQuerier, mockNodeQuerier)

			reqBody := models.CreateProbeRequest{
				NodeID:          nodeID.String(),
				Type:            "TCP",
				Target:          "example.com",
				Port:            tc.port,
				IntervalSeconds: 60,
				Count:           5,
				TimeoutSeconds:  10,
			}

			body, _ := json.Marshal(reqBody)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/api/v1/probes", strings.NewReader(string(body)))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreateProbeHandler(c)

			if tc.shouldPass {
				assert.Equal(t, http.StatusCreated, w.Code, "Port %d should be valid", tc.port)
			} else {
				assert.Equal(t, http.StatusBadRequest, w.Code, "Port %d should be invalid", tc.port)
			}
		})
	}
}

// TestCreateProbeHandler_IntervalBoundaryValues tests interval boundary validation
func TestCreateProbeHandler_IntervalBoundaryValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	nodeID := uuid.New()

	testCases := []struct {
		name        string
		interval    int
		shouldPass  bool
	}{
		{"interval_min_valid", 60, true},
		{"interval_max_valid", 300, true},
		{"interval_too_low", 59, false},
		{"interval_too_high", 301, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockNodeQuerier := &MockNodesQuerier{
				getNodeByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Node, error) {
					return &models.Node{
						ID:   nodeID.String(),
						Name: "test-node",
						IP:   "192.168.1.1",
					}, nil
				},
			}

			mockProbeQuerier := &MockProbesQuerier{}
			handler := NewProbeHandler(mockProbeQuerier, mockNodeQuerier)

			reqBody := models.CreateProbeRequest{
				NodeID:          nodeID.String(),
				Type:            "TCP",
				Target:          "example.com",
				Port:            80,
				IntervalSeconds: tc.interval,
				Count:           5,
				TimeoutSeconds:  10,
			}

			body, _ := json.Marshal(reqBody)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/api/v1/probes", strings.NewReader(string(body)))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreateProbeHandler(c)

			if tc.shouldPass {
				assert.Equal(t, http.StatusCreated, w.Code, "Interval %d should be valid", tc.interval)
			} else {
				assert.Equal(t, http.StatusBadRequest, w.Code, "Interval %d should be invalid", tc.interval)
			}
		})
	}
}

// TestCreateProbeHandler_CountBoundaryValues tests count boundary validation
func TestCreateProbeHandler_CountBoundaryValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	nodeID := uuid.New()

	testCases := []struct {
		name        string
		count       int
		shouldPass  bool
	}{
		{"count_min_valid", 1, true},
		{"count_max_valid", 100, true},
		{"count_zero", 0, false},
		{"count_too_high", 101, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockNodeQuerier := &MockNodesQuerier{
				getNodeByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Node, error) {
					return &models.Node{
						ID:   nodeID.String(),
						Name: "test-node",
						IP:   "192.168.1.1",
					}, nil
				},
			}

			mockProbeQuerier := &MockProbesQuerier{}
			handler := NewProbeHandler(mockProbeQuerier, mockNodeQuerier)

			reqBody := models.CreateProbeRequest{
				NodeID:          nodeID.String(),
				Type:            "TCP",
				Target:          "example.com",
				Port:            80,
				IntervalSeconds: 60,
				Count:           tc.count,
				TimeoutSeconds:  10,
			}

			body, _ := json.Marshal(reqBody)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/api/v1/probes", strings.NewReader(string(body)))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreateProbeHandler(c)

			if tc.shouldPass {
				assert.Equal(t, http.StatusCreated, w.Code, "Count %d should be valid", tc.count)
			} else {
				assert.Equal(t, http.StatusBadRequest, w.Code, "Count %d should be invalid", tc.count)
			}
		})
	}
}

// TestCreateProbeHandler_TimeoutBoundaryValues tests timeout boundary validation
func TestCreateProbeHandler_TimeoutBoundaryValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	nodeID := uuid.New()

	testCases := []struct {
		name        string
		timeout     int
		shouldPass  bool
	}{
		{"timeout_min_valid", 1, true},
		{"timeout_max_valid", 30, true},
		{"timeout_zero", 0, false},
		{"timeout_too_high", 31, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockNodeQuerier := &MockNodesQuerier{
				getNodeByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Node, error) {
					return &models.Node{
						ID:   nodeID.String(),
						Name: "test-node",
						IP:   "192.168.1.1",
					}, nil
				},
			}

			mockProbeQuerier := &MockProbesQuerier{}
			handler := NewProbeHandler(mockProbeQuerier, mockNodeQuerier)

			reqBody := models.CreateProbeRequest{
				NodeID:          nodeID.String(),
				Type:            "TCP",
				Target:          "example.com",
				Port:            80,
				IntervalSeconds: 60,
				Count:           5,
				TimeoutSeconds:  tc.timeout,
			}

			body, _ := json.Marshal(reqBody)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/api/v1/probes", strings.NewReader(string(body)))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreateProbeHandler(c)

			if tc.shouldPass {
				assert.Equal(t, http.StatusCreated, w.Code, "Timeout %d should be valid", tc.timeout)
			} else {
				assert.Equal(t, http.StatusBadRequest, w.Code, "Timeout %d should be invalid", tc.timeout)
			}
		})
	}
}
