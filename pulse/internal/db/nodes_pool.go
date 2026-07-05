package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// PoolQuerier implements NodesQuerier using pgxpool
type PoolQuerier struct {
	pool *pgxpool.Pool
}

// NewPoolQuerier creates a new PoolQuerier
func NewPoolQuerier(pool *pgxpool.Pool) *PoolQuerier {
	return &PoolQuerier{
		pool: pool,
	}
}

// CreateNode implements NodesQuerier
func (p *PoolQuerier) CreateNode(ctx context.Context, nodeID uuid.UUID, name string, ip string, region string, tags map[string]interface{}) error {
	return CreateNode(ctx, p.pool, nodeID, name, ip, region, tags)
}

// GetNodes implements NodesQuerier
func (p *PoolQuerier) GetNodes(ctx context.Context) ([]*models.Node, error) {
	return GetNodes(ctx, p.pool)
}

// GetNodesByRegion implements NodesQuerier
func (p *PoolQuerier) GetNodesByRegion(ctx context.Context, region string) ([]*models.Node, error) {
	return GetNodesByRegion(ctx, p.pool, region)
}

// GetNodeByID implements NodesQuerier
func (p *PoolQuerier) GetNodeByID(ctx context.Context, nodeID uuid.UUID) (*models.Node, error) {
	return GetNodeByID(ctx, p.pool, nodeID)
}

// GetNodeByNameAndIP implements NodesQuerier
func (p *PoolQuerier) GetNodeByNameAndIP(ctx context.Context, name string, ip string) (*models.Node, error) {
	return GetNodeByNameAndIP(ctx, p.pool, name, ip)
}

// GetNodeStatus implements NodesQuerier
func (p *PoolQuerier) GetNodeStatus(ctx context.Context, nodeID uuid.UUID) (*models.NodeStatus, error) {
	return GetNodeStatus(ctx, p.pool, nodeID)
}

// UpdateNode implements NodesQuerier
func (p *PoolQuerier) UpdateNode(ctx context.Context, nodeID uuid.UUID, updates map[string]interface{}) error {
	return UpdateNode(ctx, p.pool, nodeID, updates)
}

// DeleteNode implements NodesQuerier
func (p *PoolQuerier) DeleteNode(ctx context.Context, nodeID uuid.UUID) error {
	return DeleteNode(ctx, p.pool, nodeID)
}

// UpdateNodeHeartbeat implements NodesQuerier
func (p *PoolQuerier) UpdateNodeHeartbeat(ctx context.Context, nodeID uuid.UUID) (string, error) {
	return UpdateNodeHeartbeat(ctx, p.pool, nodeID)
}

// MarkNodeOffline implements NodesQuerier
func (p *PoolQuerier) MarkNodeOffline(ctx context.Context, nodeID uuid.UUID) error {
	return MarkNodeOffline(ctx, p.pool, nodeID)
}

// GetStaleNodes implements NodesQuerier
func (p *PoolQuerier) GetStaleNodes(ctx context.Context, timeout time.Duration) ([]StaleNode, error) {
	return GetStaleNodes(ctx, p.pool, timeout)
}

// CreateProbe implements ProbesQuerier
func (p *PoolQuerier) CreateProbe(ctx context.Context, probeID uuid.UUID, nodeID uuid.UUID, probeType string, target string, port int, intervalSeconds int, count int, timeoutSeconds int) error {
	return CreateProbe(ctx, p.pool, probeID, nodeID, probeType, target, port, intervalSeconds, count, timeoutSeconds)
}

// GetProbes implements ProbesQuerier
func (p *PoolQuerier) GetProbes(ctx context.Context) ([]*models.Probe, error) {
	return GetProbes(ctx, p.pool)
}

// GetProbesByNode implements ProbesQuerier
func (p *PoolQuerier) GetProbesByNode(ctx context.Context, nodeID uuid.UUID) ([]*models.Probe, error) {
	return GetProbesByNode(ctx, p.pool, nodeID)
}

// GetProbeByID implements ProbesQuerier
func (p *PoolQuerier) GetProbeByID(ctx context.Context, probeID uuid.UUID) (*models.Probe, error) {
	return GetProbeByID(ctx, p.pool, probeID)
}

// UpdateProbe implements ProbesQuerier
func (p *PoolQuerier) UpdateProbe(ctx context.Context, probeID uuid.UUID, updates map[string]interface{}) error {
	return UpdateProbe(ctx, p.pool, probeID, updates)
}

// DeleteProbe implements ProbesQuerier
func (p *PoolQuerier) DeleteProbe(ctx context.Context, probeID uuid.UUID) error {
	return DeleteProbe(ctx, p.pool, probeID)
}

// GetBeaconConfig implements BeaconConfigsQuerier.
func (p *PoolQuerier) GetBeaconConfig(ctx context.Context, beaconID uuid.UUID) (*BeaconConfig, error) {
	return GetBeaconConfig(ctx, p.pool, beaconID)
}

// UpsertBeaconConfig implements BeaconConfigsQuerier.
func (p *PoolQuerier) UpsertBeaconConfig(ctx context.Context, beaconID uuid.UUID, update BeaconConfigUpdate) (*BeaconConfig, error) {
	return UpsertBeaconConfig(ctx, p.pool, beaconID, update)
}

// GetBeaconConfigHistory implements BeaconConfigsQuerier.
func (p *PoolQuerier) GetBeaconConfigHistory(ctx context.Context, beaconID uuid.UUID, limit int) ([]BeaconConfigHistoryEntry, error) {
	return GetBeaconConfigHistory(ctx, p.pool, beaconID, limit)
}

// AcknowledgeBeaconConfig implements BeaconConfigsQuerier.
func (p *PoolQuerier) AcknowledgeBeaconConfig(ctx context.Context, beaconID uuid.UUID, version int, status string, errorMessage string) error {
	return AcknowledgeBeaconConfig(ctx, p.pool, beaconID, version, status, errorMessage)
}

// SaveMTRResult implements MTRResultsQuerier.
func (p *PoolQuerier) SaveMTRResult(ctx context.Context, input MTRResultInput) (*MTRResult, error) {
	return SaveMTRResult(ctx, p.pool, input)
}

// GetLatestMTRResult implements MTRResultsQuerier.
func (p *PoolQuerier) GetLatestMTRResult(ctx context.Context, nodeID uuid.UUID) (*MTRResult, error) {
	return GetLatestMTRResult(ctx, p.pool, nodeID)
}

// GetMTRResults implements MTRResultsQuerier.
func (p *PoolQuerier) GetMTRResults(ctx context.Context, query MTRResultQuery) ([]MTRResult, error) {
	return GetMTRResults(ctx, p.pool, query)
}
