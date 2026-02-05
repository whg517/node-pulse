package models

import "time"

// Node represents a monitoring node in system
type Node struct {
	ID             string     `json:"id" db:"id"`
	Name           string     `json:"name" db:"name"`
	IP             string     `json:"ip" db:"ip"`
	Region         string     `json:"region" db:"region"`
	Tags           string     `json:"tags,omitempty" db:"tags"`           // JSONB stored as string
	LastHeartbeat  *time.Time `json:"last_heartbeat,omitempty" db:"last_heartbeat"`   // Beacon heartbeat time
	LastReportTime *time.Time `json:"last_report_time,omitempty" db:"last_report_time"` // Data write time
	Status         string     `json:"status" db:"status"`                           // online/offline/connecting
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateNodeRequest represents request to create a new node
type CreateNodeRequest struct {
	Name   string                 `json:"name" binding:"required,max=255"`
	IP     string                 `json:"ip" binding:"required,max=45"`
	Region string                 `json:"region" binding:"required,max=100"`
	Tags   map[string]interface{} `json:"tags,omitempty"`
}

// UpdateNodeRequest represents request to update a node
type UpdateNodeRequest struct {
	Name   *string                 `json:"name,omitempty" binding:"omitempty,max=255"`
	IP     *string                 `json:"ip,omitempty" binding:"omitempty,max=45"`
	Region *string                 `json:"region,omitempty" binding:"omitempty,max=100"`
	Tags   *map[string]interface{} `json:"tags,omitempty"`
}

// CreateNodeData represents node data in create response
// Note: Changed from nested {data: {node: {...}}} to flat {data: {...}} format
// This allows Beacon client to directly access node fields (AC #2: "返回完整的节点信息（含自动生成的 UUID)")
type CreateNodeData struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	IP        string    `json:"ip"`
	Region    string    `json:"region"`
	Tags      string    `json:"tags,omitempty"`           // JSONB stored as string
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateNodeResponse represents successful node creation response
type CreateNodeResponse struct {
	Data CreateNodeData `json:"data"`
	Message   string    `json:"message"`
	Timestamp string `json:"timestamp"`
}

// GetNodesData represents nodes data in list response
type GetNodesData struct {
	Nodes []*Node `json:"nodes"`
}

// GetNodesResponse represents successful node list retrieval response
type GetNodesResponse struct {
	Data GetNodesData `json:"data"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// UpdateNodeResponse represents successful node update response
type UpdateNodeResponse struct {
	Data struct {
		Node *Node `json:"node"`
	} `json:"data"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// GetNodeData represents node data in get response
type GetNodeData struct {
	Node *Node `json:"node"`
}

// GetNodeResponse represents successful node retrieval response
type GetNodeResponse struct {
	Data GetNodeData `json:"data"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// DeleteNodeResponse represents successful node deletion response
type DeleteNodeResponse struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// NodeStatusData represents node status data in response
type NodeStatusData struct {
	Node *NodeStatus `json:"node"`
}

// NodeStatus represents node status information
type NodeStatus struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Status         string     `json:"status"`
	LastHeartbeat  *time.Time `json:"last_heartbeat,omitempty"`
	LastReportTime *time.Time `json:"last_report_time,omitempty"`
}

// GetNodeStatusResponse represents successful node status retrieval response
type GetNodeStatusResponse struct {
	Data    NodeStatusData `json:"data"`
	Message string       `json:"message"`
	Timestamp string     `json:"timestamp"`
}
