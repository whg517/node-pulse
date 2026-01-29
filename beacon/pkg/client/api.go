package client

// APIClient handles communication with Pulse API
// This package will be implemented in future stories
type APIClient struct {
	// API client implementation will go here
}

// NewAPIClient creates a new API client
func NewAPIClient(serverURL string) *APIClient {
	return &APIClient{}
}

// RegisterNode registers this node with the Pulse server
func (c *APIClient) RegisterNode(data any) error {
	// Placeholder - will be implemented in future stories
	return nil
}
