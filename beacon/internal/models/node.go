package models

// Node represents a monitoring node
type Node struct {
	ID          string `json:"node_id"`
	Name        string `json:"node_name"`
	Status      string `json:"status"`      // online, offline, connecting
	LastHeartbeat string `json:"last_heartbeat"`
	ConfigVersion string `json:"config_version"`
}

// NodeRegistration represents node registration data
type NodeRegistration struct {
	NodeID      string `json:"node_id"`
	NodeName    string `json:"node_name"`
	IPAddress   string `json:"ip_address"`
	Region      string `json:"region"`
	Labels      map[string]string `json:"labels,omitempty"`
}
