package cache

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MetricRecord represents a metric record to be written to PostgreSQL
type MetricRecord struct {
	NodeID         string
	ProbeID        string
	Timestamp      time.Time
	LatencyMs      float64
	PacketLossRate float64
	JitterMs       float64
	IsAggregated   bool
}

// BatchWriter handles async batch writing of metrics to PostgreSQL
type BatchWriter struct {
	buffer      chan *MetricRecord // Buffer channel (capacity 1000)
	flushTicker *time.Ticker       // 1-minute ticker for timeout flush
	db          *pgxpool.Pool      // PostgreSQL connection pool
	batchSize   int                // Batch size trigger (default 100)
	ctx         context.Context    // Context for cancellation
	cancel      context.CancelFunc // Cancel function
	wg          sync.WaitGroup     // Wait group for graceful shutdown
}

// NewBatchWriter creates a new batch writer
func NewBatchWriter(db *pgxpool.Pool, bufferSize, batchSize int) *BatchWriter {
	ctx, cancel := context.WithCancel(context.Background())

	return &BatchWriter{
		buffer:      make(chan *MetricRecord, bufferSize),
		flushTicker: time.NewTicker(1 * time.Minute),
		db:          db,
		batchSize:   batchSize,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start begins the background goroutine for batch writing
func (bw *BatchWriter) Start() {
	bw.wg.Add(1)
	go bw.processBatches()
}

// Stop gracefully stops the batch writer
func (bw *BatchWriter) Stop() {
	bw.cancel()
	bw.flushTicker.Stop()

	// Wait for processBatches goroutine to finish
	bw.wg.Wait()

	// Now safe to drain buffer and close
	bw.flush()
	close(bw.buffer)
}

// Write adds a metric record to the buffer (non-blocking)
func (bw *BatchWriter) Write(record *MetricRecord) error {
	if record == nil {
		return ErrNilMetricRecord
	}

	select {
	case bw.buffer <- record:
		return nil
	default:
		// Buffer is full, drop the metric
		return ErrBufferFull
	}
}

// processBatches runs in background goroutine, processing batch writes
func (bw *BatchWriter) processBatches() {
	defer bw.wg.Done()

	batch := make([]*MetricRecord, 0, bw.batchSize)

	for {
		select {
		case <-bw.ctx.Done():
			// Context cancelled, flush and exit
			if len(batch) > 0 {
				bw.writeBatchWithRetry(batch)
			}
			return

		case record := <-bw.buffer:
			batch = append(batch, record)

			// Trigger batch write when batch size is reached
			if len(batch) >= bw.batchSize {
				bw.writeBatchWithRetry(batch)
				batch = make([]*MetricRecord, 0, bw.batchSize)
			}

		case <-bw.flushTicker.C:
			// Timeout flush (1 minute)
			if len(batch) > 0 {
				bw.writeBatchWithRetry(batch)
				batch = make([]*MetricRecord, 0, bw.batchSize)
			}
		}
	}
}

// writeBatchWithRetry writes a batch to PostgreSQL with retry mechanism
func (bw *BatchWriter) writeBatchWithRetry(batch []*MetricRecord) {
	maxRetries := 3
	backoff := time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := bw.writeBatch(batch)
		if err == nil {
			// Success
			return
		}

		slog.Error("Failed to write batch to PostgreSQL",
			"attempt", attempt,
			"max_retries", maxRetries,
			"batch_size", len(batch),
			"error", err,
			"sample_data", fmt.Sprintf("node_id=%s, probe_id=%s, timestamp=%s",
				batch[0].NodeID, batch[0].ProbeID, batch[0].Timestamp))

		// Retry with exponential backoff
		if attempt < maxRetries {
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff: 1s, 2s, 4s
		} else {
			// Max retries exhausted
			slog.Error("Batch write failed after max retries",
				"batch_size", len(batch),
				"last_error", err)
		}
	}
}

// writeBatch performs batch insert using prepared statements with transaction
// Note: Using transactional batch INSERT for compatibility and simplicity
// Future optimization: Consider pgx.CopyFrom for higher throughput with large batches
func (bw *BatchWriter) writeBatch(batch []*MetricRecord) error {
	if len(batch) == 0 {
		return nil
	}

	if bw.db == nil {
		// No database configured, skip writing (for testing)
		return nil
	}

	// Use transactional batch INSERT for atomicity and consistency
	ctx, cancel := context.WithTimeout(bw.ctx, 30*time.Second)
	defer cancel()

	tx, err := bw.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Batch insert statement
	stmt := `
		INSERT INTO metrics (
			node_id, probe_id, timestamp,
			latency_ms, packet_loss_rate, jitter_ms,
			is_aggregated, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, NOW()
		)
	`

	// Execute insert for each record within transaction
	for _, record := range batch {
		_, err := tx.Exec(ctx, stmt,
			record.NodeID,
			record.ProbeID,
			record.Timestamp,
			record.LatencyMs,
			record.PacketLossRate,
			record.JitterMs,
			record.IsAggregated,
		)
		if err != nil {
			return fmt.Errorf("failed to insert record: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Debug("Successfully wrote batch to PostgreSQL",
		"batch_size", len(batch))

	return nil
}

// flush writes any remaining records in the buffer
func (bw *BatchWriter) flush() {
	// Drain the buffer and write remaining records
	remaining := make([]*MetricRecord, 0, len(bw.buffer))
	for {
		select {
		case record := <-bw.buffer:
			remaining = append(remaining, record)
		default:
			// Buffer is empty
			if len(remaining) > 0 {
				bw.writeBatchWithRetry(remaining)
			}
			return
		}
	}
}

// GetBufferSize returns the current buffer size
func (bw *BatchWriter) GetBufferSize() int {
	return len(bw.buffer)
}
