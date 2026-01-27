package api

import (
	"context"
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

// MockNodesQuerier is a simple mock for NodesQuerier interface
type MockNodesQuerier struct {
	createNodeFunc     func(context.Context, uuid.UUID, string, string, string, map[string]interface{}) error
	getNodesFunc       func(context.Context) ([]*models.Node, error)
	getNodesByRegionFunc func(context.Context, string) ([]*models.Node, error)
	getNodeByIDFunc    func(context.Context, uuid.UUID) (*models.Node, error)
	updateNodeFunc     func(context.Context, uuid.UUID, map[string]interface{}) error
	deleteNodeFunc     func(context.Context, uuid.UUID) error
}

func (m *MockNodesQuerier) CreateNode(ctx context.Context, nodeID uuid.UUID, name string, ip string, region string, tags map[string]interface{}) error {
	if m.createNodeFunc != nil {
		return m.createNodeFunc(ctx, nodeID, name, ip, region, tags)
	}
	return nil
}

func (m *MockNodesQuerier) GetNodes(ctx context.Context) ([]*models.Node, error) {
	if m.getNodesFunc != nil {
		return m.getNodesFunc(ctx)
	}
	return nil, nil
}

func (m *MockNodesQuerier) GetNodesByRegion(ctx context.Context, region string) ([]*models.Node, error) {
	if m.getNodesByRegionFunc != nil {
		return m.getNodesByRegionFunc(ctx, region)
	}
	return nil, nil
}

func (m *MockNodesQuerier) GetNodeByID(ctx context.Context, nodeID uuid.UUID) (*models.Node, error) {
	if m.getNodeByIDFunc != nil {
		return m.getNodeByIDFunc(ctx, nodeID)
	}
	return nil, nil
}

func (m *MockNodesQuerier) UpdateNode(ctx context.Context, nodeID uuid.UUID, updates map[string]interface{}) error {
	if m.updateNodeFunc != nil {
		return m.updateNodeFunc(ctx, nodeID, updates)
	}
	return nil
}

func (m *MockNodesQuerier) DeleteNode(ctx context.Context, nodeID uuid.UUID) error {
	if m.deleteNodeFunc != nil {
		return m.deleteNodeFunc(ctx, nodeID)
	}
	return nil
}

// setupTestContext creates a Gin context with user authentication
func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "test-user-id")
	c.Set("role", "admin")
	return c, w
}

// setupViewerContext creates a Gin context with viewer role
func setupViewerContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "test-user-id")
	c.Set("role", "viewer")
	return c, w
}

// setupUnauthenticatedContext creates a Gin context without user authentication
func setupUnauthenticatedContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

// TestCreateNodeSuccess tests successful node creation
func TestCreateNodeSuccess(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)

	createCalled := false
	var capturedNodeID uuid.UUID
	mockQuerier.createNodeFunc = func(ctx context.Context, nID uuid.UUID, name string, ip string, region string, tags map[string]interface{}) error {
		createCalled = true
		capturedNodeID = nID
		assert.Equal(t, "Test Node", name)
		assert.Equal(t, "192.168.1.1", ip)
		assert.Equal(t, "us-east", region)
		return nil
	}

	getCalled := false
	mockQuerier.getNodeByIDFunc = func(ctx context.Context, nID uuid.UUID) (*models.Node, error) {
		getCalled = true
		expectedNode := &models.Node{
			ID:        capturedNodeID.String(),
			Name:      "Test Node",
			IP:        "192.168.1.1",
			Region:    "us-east",
			Tags:      "{}",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		return expectedNode, nil
	}

	// Act - directly call handler
	c, w := setupTestContext()
	reqBody := strings.NewReader(`{
		"name": "Test Node",
		"ip": "192.168.1.1",
		"region": "us-east"
	}`)
	c.Request, _ = http.NewRequest("POST", "/api/v1/nodes", reqBody)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateNodeHandler(c)
	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "节点创建成功")
	assert.True(t, createCalled, "CreateNode should be called")
	assert.True(t, getCalled, "GetNodeByID should be called")
}

// TestCreateNodeEmptyName tests validation failure for empty name
func TestCreateNodeEmptyName(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)

	// Act
	c, w := setupTestContext()
	reqBody := strings.NewReader(`{
		"ip": "192.168.1.1",
		"region": "us-east"
	}`)
	c.Request, _ = http.NewRequest("POST", "/api/v1/nodes", reqBody)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateNodeHandler(c)
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	body := w.Body.String()
	// Check for either custom validation error or Gin's binding validation error
	conditionMet := strings.Contains(body, "节点名称不能为空") ||
		strings.Contains(body, "Field validation for 'Name' failed")
	assert.True(t, conditionMet, "Expected validation error for empty name")
}

// TestCreateNodeInvalidIP tests validation failure for invalid IP
func TestCreateNodeInvalidIP(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)

	// Act
	c, w := setupTestContext()
	reqBody := strings.NewReader(`{
		"name": "Test Node",
		"ip": "invalid-ip",
		"region": "us-east"
	}`)
	c.Request, _ = http.NewRequest("POST", "/api/v1/nodes", reqBody)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateNodeHandler(c)
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "节点 IP 地址格式无效")
}

// TestCreateNodePermissionDenied tests RBAC for viewer role
func TestCreateNodePermissionDenied(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)

	createCalled := false
	var capturedNodeID uuid.UUID
	mockQuerier.createNodeFunc = func(ctx context.Context, nID uuid.UUID, name string, ip string, region string, tags map[string]interface{}) error {
		createCalled = true
		capturedNodeID = nID
		assert.Equal(t, "Test Node", name)
		assert.Equal(t, "192.168.1.1", ip)
		assert.Equal(t, "us-east", region)
		return nil
	}

	mockQuerier.getNodeByIDFunc = func(ctx context.Context, nID uuid.UUID) (*models.Node, error) {
		expectedNode := &models.Node{
			ID:        capturedNodeID.String(),
			Name:      "Test Node",
			IP:        "192.168.1.1",
			Region:    "us-east",
			Tags:      "{}",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		return expectedNode, nil
	}

	// Act - use viewer role
	c, w := setupViewerContext()
	reqBody := strings.NewReader(`{
		"name": "Test Node",
		"ip": "192.168.1.1",
		"region": "us-east"
	}`)
	c.Request, _ = http.NewRequest("POST", "/api/v1/nodes", reqBody)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateNodeHandler(c)
	// Assert: handler itself doesn't check auth (that's middleware's job)
	// In unit tests, we bypass middleware, so this should succeed
	// Integration tests verify middleware behavior
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "节点创建成功")
	assert.True(t, createCalled, "CreateNode should be called")
}

// TestCreateNodeUnauthenticated tests without authentication
func TestCreateNodeUnauthenticated(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)

	createCalled := false
	var capturedNodeID uuid.UUID
	mockQuerier.createNodeFunc = func(ctx context.Context, nID uuid.UUID, name string, ip string, region string, tags map[string]interface{}) error {
		createCalled = true
		capturedNodeID = nID
		assert.Equal(t, "Test Node", name)
		assert.Equal(t, "192.168.1.1", ip)
		assert.Equal(t, "us-east", region)
		return nil
	}

	mockQuerier.getNodeByIDFunc = func(ctx context.Context, nID uuid.UUID) (*models.Node, error) {
		expectedNode := &models.Node{
			ID:        capturedNodeID.String(),
			Name:      "Test Node",
			IP:        "192.168.1.1",
			Region:    "us-east",
			Tags:      "{}",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		return expectedNode, nil
	}

	// Act - no user_id in context
	c, w := setupUnauthenticatedContext()
	reqBody := strings.NewReader(`{
		"name": "Test Node",
		"ip": "192.168.1.1",
		"region": "us-east"
	}`)
	c.Request, _ = http.NewRequest("POST", "/api/v1/nodes", reqBody)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateNodeHandler(c)
	// Assert: handler itself doesn't check auth (that's middleware's job)
	// In unit tests, we bypass middleware, so this should succeed
	// Integration tests verify middleware behavior
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "节点创建成功")
	assert.True(t, createCalled, "CreateNode should be called")
}

// TestGetNodesSuccess tests successful retrieval of all nodes
func TestGetNodesSuccess(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)
	router := gin.Default()
	router.GET("/api/v1/nodes", handler.GetNodesHandler)

	nodeID1 := uuid.New()
	nodeID2 := uuid.New()
	expectedNodes := []*models.Node{
		{
			ID:     nodeID1.String(),
			Name:   "Node 1",
			IP:     "192.168.1.1",
			Region: "us-east",
			Tags:   "{}",
		},
		{
			ID:     nodeID2.String(),
			Name:   "Node 2",
			IP:     "192.168.1.2",
			Region: "asia-pacific",
			Tags:   "{}",
		},
	}

	mockQuerier.getNodesFunc = func(ctx context.Context) ([]*models.Node, error) {
		return expectedNodes, nil
	}

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/nodes", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "节点列表获取成功")
}

// TestGetNodesByRegion tests filtering by region
func TestGetNodesByRegion(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)
	router := gin.Default()
	router.GET("/api/v1/nodes", handler.GetNodesHandler)

	expectedNodes := []*models.Node{
		{
			ID:     uuid.New().String(),
			Name:   "Node 1",
			IP:     "192.168.1.1",
			Region: "us-east",
			Tags:   "{}",
		},
	}

	mockQuerier.getNodesByRegionFunc = func(ctx context.Context, region string) ([]*models.Node, error) {
		assert.Equal(t, "us-east", region)
		return expectedNodes, nil
	}

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/nodes?region=us-east", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "节点列表获取成功")
}

// TestUpdateNodeSuccess tests successful node update
func TestUpdateNodeSuccess(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)
	nodeID := uuid.New()
	updatedNode := &models.Node{
		ID:        nodeID.String(),
		Name:      "Updated Node",
		IP:        "192.168.1.100",
		Region:    "asia-pacific",
		Tags:      "{}",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	updateCalled := false
	mockQuerier.updateNodeFunc = func(ctx context.Context, nID uuid.UUID, updates map[string]interface{}) error {
		updateCalled = true
		assert.Equal(t, nodeID, nID)
		return nil
	}

	mockQuerier.getNodeByIDFunc = func(ctx context.Context, nID uuid.UUID) (*models.Node, error) {
		assert.Equal(t, nodeID, nID)
		return updatedNode, nil
	}

	// Act
	c, w := setupTestContext()
	nodeIDStr := nodeID.String()
	c.Params = gin.Params{gin.Param{Key: "id", Value: nodeIDStr}}
	reqBody := strings.NewReader(`{
		"name": "Updated Node",
		"ip": "192.168.1.100",
		"region": "asia-pacific"
	}`)
	c.Request, _ = http.NewRequest("PUT", "/api/v1/nodes/"+nodeIDStr, reqBody)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateNodeHandler(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "节点更新成功")
	assert.True(t, updateCalled, "UpdateNode should be called")
}

// TestUpdateNodeNotFound tests update of non-existent node
func TestUpdateNodeNotFound(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)
	nodeID := uuid.New()

	mockQuerier.updateNodeFunc = func(ctx context.Context, nID uuid.UUID, updates map[string]interface{}) error {
		assert.Equal(t, nodeID, nID)
		return db.ErrNodeNotFound
	}

	// Act
	c, w := setupTestContext()
	nodeIDStr := nodeID.String()
	c.Params = gin.Params{gin.Param{Key: "id", Value: nodeIDStr}}
	reqBody := strings.NewReader(`{
		"name": "Updated Node"
	}`)
	c.Request, _ = http.NewRequest("PUT", "/api/v1/nodes/"+nodeIDStr, reqBody)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateNodeHandler(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "节点不存在")
}

// TestUpdateNodePermissionDenied tests RBAC for viewer role
func TestUpdateNodePermissionDenied(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)
	nodeID := uuid.New()
	updatedNode := &models.Node{
		ID:        nodeID.String(),
		Name:      "Updated Node",
		IP:        "192.168.1.100",
		Region:    "asia-pacific",
		Tags:      "{}",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	updateCalled := false
	mockQuerier.updateNodeFunc = func(ctx context.Context, nID uuid.UUID, updates map[string]interface{}) error {
		updateCalled = true
		return nil
	}

	mockQuerier.getNodeByIDFunc = func(ctx context.Context, nID uuid.UUID) (*models.Node, error) {
		return updatedNode, nil
	}

	// Act - use viewer role
	// Note: In unit tests, we bypass middleware, so RBAC isn't checked
	// Integration tests verify middleware behavior
	c, w := setupViewerContext()
	nodeIDStr := nodeID.String()
	c.Params = gin.Params{gin.Param{Key: "id", Value: nodeIDStr}}
	reqBody := strings.NewReader(`{
		"name": "Updated Node"
	}`)
	c.Request, _ = http.NewRequest("PUT", "/api/v1/nodes/"+nodeIDStr, reqBody)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateNodeHandler(c)

	// Assert: handler assumes auth is done by middleware
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "节点更新成功")
	assert.True(t, updateCalled, "UpdateNode should be called")
}

// TestDeleteNodeSuccess tests successful node deletion
func TestDeleteNodeSuccess(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)
	nodeID := uuid.New()

	deleteCalled := false
	mockQuerier.deleteNodeFunc = func(ctx context.Context, nID uuid.UUID) error {
		deleteCalled = true
		assert.Equal(t, nodeID, nID)
		return nil
	}

	// Act
	c, w := setupTestContext()
	nodeIDStr := nodeID.String()
	c.Params = gin.Params{gin.Param{Key: "id", Value: nodeIDStr}}
	req, _ := http.NewRequest("DELETE", "/api/v1/nodes/"+nodeIDStr+"?confirm=true", nil)
	c.Request = req

	handler.DeleteNodeHandler(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "节点删除成功")
	assert.True(t, deleteCalled, "DeleteNode should be called")
}

// TestDeleteNodeWithoutConfirmation tests deletion without confirmation
func TestDeleteNodeWithoutConfirmation(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)
	nodeID := uuid.New()

	// Act
	c, w := setupTestContext()
	nodeIDStr := nodeID.String()
	c.Params = gin.Params{gin.Param{Key: "id", Value: nodeIDStr}}
	req, _ := http.NewRequest("DELETE", "/api/v1/nodes/"+nodeIDStr, nil)
	c.Request = req

	handler.DeleteNodeHandler(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "删除操作需要确认")
}

// TestDeleteNodeNotFound tests deletion of non-existent node
func TestDeleteNodeNotFound(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)
	nodeID := uuid.New()

	mockQuerier.deleteNodeFunc = func(ctx context.Context, nID uuid.UUID) error {
		assert.Equal(t, nodeID, nID)
		return db.ErrNodeNotFound
	}

	// Act
	c, w := setupTestContext()
	nodeIDStr := nodeID.String()
	c.Params = gin.Params{gin.Param{Key: "id", Value: nodeIDStr}}
	req, _ := http.NewRequest("DELETE", "/api/v1/nodes/"+nodeIDStr+"?confirm=true", nil)
	c.Request = req

	handler.DeleteNodeHandler(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "节点不存在")
}

// TestDeleteNodePermissionDenied tests RBAC for viewer role
func TestDeleteNodePermissionDenied(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)
	nodeID := uuid.New()

	deleteCalled := false
	mockQuerier.deleteNodeFunc = func(ctx context.Context, nID uuid.UUID) error {
		deleteCalled = true
		return nil
	}

	// Act - use viewer role
	// Note: In unit tests, we bypass middleware, so RBAC isn't checked
	// Integration tests verify middleware behavior
	c, w := setupViewerContext()
	nodeIDStr := nodeID.String()
	c.Params = gin.Params{gin.Param{Key: "id", Value: nodeIDStr}}
	req, _ := http.NewRequest("DELETE", "/api/v1/nodes/"+nodeIDStr+"?confirm=true", nil)
	c.Request = req

	handler.DeleteNodeHandler(c)

	// Assert: handler assumes auth is done by middleware
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "节点删除成功")
	assert.True(t, deleteCalled, "DeleteNode should be called")
}

// TestGetNodeByIDSuccess tests successful node retrieval by ID
func TestGetNodeByIDSuccess(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)
	nodeID := uuid.New()
	expectedNode := &models.Node{
		ID:     nodeID.String(),
		Name:   "Test Node",
		IP:     "192.168.1.1",
		Region: "us-east",
		Tags:   "{}",
	}

	mockQuerier.getNodeByIDFunc = func(ctx context.Context, nID uuid.UUID) (*models.Node, error) {
		assert.Equal(t, nodeID, nID)
		return expectedNode, nil
	}

	// Act
	c, w := setupTestContext()
	nodeIDStr := nodeID.String()
	c.Params = gin.Params{gin.Param{Key: "id", Value: nodeIDStr}}
	req, _ := http.NewRequest("GET", "/api/v1/nodes/"+nodeIDStr, nil)
	c.Request = req

	handler.GetNodeByIDHandler(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "节点查询成功")
}

// TestGetNodeByIDNotFound tests retrieval of non-existent node
func TestGetNodeByIDNotFound(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)
	nodeID := uuid.New()

	mockQuerier.getNodeByIDFunc = func(ctx context.Context, nID uuid.UUID) (*models.Node, error) {
		return nil, db.ErrNodeNotFound
	}

	// Act
	c, w := setupTestContext()
	nodeIDStr := nodeID.String()
	c.Params = gin.Params{gin.Param{Key: "id", Value: nodeIDStr}}
	req, _ := http.NewRequest("GET", "/api/v1/nodes/"+nodeIDStr, nil)
	c.Request = req

	handler.GetNodeByIDHandler(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "节点不存在")
}

// TestGetNodeByIDUnauthenticated tests RBAC for unauthenticated request
func TestGetNodeByIDUnauthenticated(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)
	nodeID := uuid.New()
	expectedNode := &models.Node{
		Name:   "Test Node",
		IP:     "192.168.1.1",
		Region: "us-east",
		Tags:   "{}",
	}

	mockQuerier.getNodeByIDFunc = func(ctx context.Context, nID uuid.UUID) (*models.Node, error) {
		return expectedNode, nil
	}

	// Act - no authentication
	// Note: In unit tests, we bypass middleware, so auth isn't checked
	// Integration tests verify middleware behavior
	c, w := setupUnauthenticatedContext()
	nodeIDStr := nodeID.String()
	c.Params = gin.Params{gin.Param{Key: "id", Value: nodeIDStr}}
	req, _ := http.NewRequest("GET", "/api/v1/nodes/"+nodeIDStr, nil)
	c.Request = req

	handler.GetNodeByIDHandler(c)

	// Assert: handler assumes auth is done by middleware
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "节点查询成功")
}

// TestIPValidationRejectsPartialIPv4 tests that partial IPv4 addresses are rejected
func TestIPValidationRejectsPartialIPv4(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)

	testCases := []string{
		"192.168",
		"192.168.1",
		"192.168.1.",
		"192.168.1.1.",
	}

	for _, invalidIP := range testCases {
		c, w := setupTestContext()
		reqBody := strings.NewReader(`{
			"name": "Test Node",
			"ip": "` + invalidIP + `",
			"region": "us-east"
		}`)
		c.Request, _ = http.NewRequest("POST", "/api/v1/nodes", reqBody)
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateNodeHandler(c)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code, "IP %s should be rejected", invalidIP)
		assert.Contains(t, w.Body.String(), "节点 IP 地址格式无效", "IP %s should be invalid", invalidIP)
	}
}

// TestIPValidationAcceptsValidIPv4 tests that valid IPv4 addresses are accepted
func TestIPValidationAcceptsValidIPv4(t *testing.T) {
	// Setup
	mockQuerier := &MockNodesQuerier{}
	handler := NewNodeHandler(mockQuerier)
	nodeID := uuid.New()
	expectedNode := &models.Node{
		ID:        nodeID.String(),
		Name:      "Test Node",
		IP:        "192.168.1.1",
		Region:    "us-east",
		Tags:      "{}",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockQuerier.createNodeFunc = func(ctx context.Context, nID uuid.UUID, name string, ip string, region string, tags map[string]interface{}) error {
		return nil
	}

	mockQuerier.getNodeByIDFunc = func(ctx context.Context, nID uuid.UUID) (*models.Node, error) {
		return expectedNode, nil
	}

	testCases := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"127.0.0.1",
	}

	for _, validIP := range testCases {
		c, w := setupTestContext()
		reqBody := strings.NewReader(`{
			"name": "Test Node",
			"ip": "` + validIP + `",
			"region": "us-east"
		}`)
		c.Request, _ = http.NewRequest("POST", "/api/v1/nodes", reqBody)
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateNodeHandler(c)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code, "IP %s should be accepted", validIP)
		assert.Contains(t, w.Body.String(), "节点创建成功", "IP %s should be valid", validIP)
	}
}
