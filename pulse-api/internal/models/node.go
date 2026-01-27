package models

import "time"

// Node represents a monitoring node in the system
type Node struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	IP        string    `json:"ip" db:"ip"`
	Region    string    `json:"region" db:"region"`
	Tags      string    `json:"tags,omitempty" db:"tags"`      // JSONB stored as string
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateNodeRequest represents the request to create a new node
type CreateNodeRequest struct {
	Name   string                 `json:"name" binding:"required,max=255"`
	IP     string                 `json:"ip" binding:"required,max=45"`
	Region string                 `json:"region" binding:"required,max=100"`
	Tags   map[string]interface{} `json:"tags,omitempty"`
}

// UpdateNodeRequest represents the request to update a node
type UpdateNodeRequest struct {
	Name   *string                 `json:"name,omitempty" binding:"omitempty,max=255"`
	IP     *string                 `json:"ip,omitempty" binding:"omitempty,max=45"`
	Region *string                 `json:"region,omitempty" binding:"omitempty,max=100"`
	Tags   *map[string]interface{} `json:"tags,omitempty"`
}

// CreateNodeData represents node data in create response
type CreateNodeData struct {
	Node *Node `json:"node"`
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
	Message   string    `json:"message"`
	Timestamp string    `json:"timestamp"`
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
	Message   string    `json:"message"`
	Timestamp string `json:"timestamp"`
}

// DeleteNodeResponse represents successful node deletion response
type DeleteNodeResponse struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}
