package reporter

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"sync"

	"beacon/internal/logger"
)

// CompressionStats tracks compression statistics
type CompressionStats struct {
	mu sync.RWMutex

	// Total bytes before compression
	TotalOriginalBytes int64

	// Total bytes after compression
	TotalCompressedBytes int64

	// Number of compression operations
	CompressionCount int64

	// Number of decompression operations
	DecompressionCount int64

	// Number of corruption detections
	CorruptionCount int64
}

// Compressor handles GZIP compression with CRC32 checksum (FR-4.1.5)
type Compressor struct {
	// Compression level (1-9)
	level int

	// Minimum size to trigger compression (bytes)
	minSize int

	// Statistics
	stats CompressionStats
}

// NewCompressor creates a new compressor with the specified level
func NewCompressor(level int, minSize int) *Compressor {
	// Validate level
	if level < 1 || level > 9 {
		level = 6 // Default level
	}

	// Validate min size
	if minSize < 0 {
		minSize = 1024 // Default 1KB
	}

	return &Compressor{
		level:   level,
		minSize: minSize,
	}
}

// CompressedData represents compressed data with metadata
type CompressedData struct {
	// Original data CRC32 checksum
	Checksum uint32 `json:"checksum"`

	// Original data size
	OriginalSize int64 `json:"original_size"`

	// Compressed data
	Data []byte `json:"data"`

	// Whether data is compressed
	IsCompressed bool `json:"is_compressed"`
}

// Compress compresses data with GZIP and adds CRC32 checksum
func (c *Compressor) Compress(data []byte) (*CompressedData, error) {
	if len(data) == 0 {
		return &CompressedData{
			Checksum:     0,
			OriginalSize: 0,
			Data:         []byte{},
			IsCompressed: false,
		}, nil
	}

	// Calculate CRC32 checksum of original data
	checksum := crc32.ChecksumIEEE(data)
	originalSize := int64(len(data))

	// Check if compression is beneficial
	if originalSize < int64(c.minSize) {
		c.stats.mu.Lock()
		c.stats.TotalOriginalBytes += originalSize
		c.stats.TotalCompressedBytes += originalSize
		c.stats.CompressionCount++
		c.stats.mu.Unlock()

		return &CompressedData{
			Checksum:     checksum,
			OriginalSize: originalSize,
			Data:         data,
			IsCompressed: false,
		}, nil
	}

	// Compress with GZIP
	var buf bytes.Buffer

	// Write checksum as first 4 bytes (for integrity check during decompression)
	checksumBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(checksumBytes, checksum)
	buf.Write(checksumBytes)

	// Create gzip writer with specified level
	writer, err := gzip.NewWriterLevel(&buf, c.level)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip writer: %w", err)
	}

	// Write compressed data
	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to compress data: %w", err)
	}

	// Close writer to flush data
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	compressedData := buf.Bytes()
	compressedSize := int64(len(compressedData))

	// Update statistics
	c.stats.mu.Lock()
	c.stats.TotalOriginalBytes += originalSize
	c.stats.TotalCompressedBytes += compressedSize
	c.stats.CompressionCount++
	c.stats.mu.Unlock()

	// Log compression ratio
	ratio := float64(0)
	if originalSize > 0 {
		ratio = float64(originalSize-compressedSize) / float64(originalSize) * 100
	}

	logger.WithFields(map[string]interface{}{
		"component":      "compression",
		"original_size":  originalSize,
		"compressed_size": compressedSize,
		"ratio_percent":  ratio,
		"level":          c.level,
	}).Debug("Data compressed")

	return &CompressedData{
		Checksum:     checksum,
		OriginalSize: originalSize,
		Data:         compressedData,
		IsCompressed: true,
	}, nil
}

// Decompress decompresses GZIP data and verifies CRC32 checksum
func (c *Compressor) Decompress(compressed *CompressedData) ([]byte, error) {
	if compressed == nil {
		return nil, fmt.Errorf("compressed data is nil")
	}

	if len(compressed.Data) == 0 {
		return []byte{}, nil
	}

	// If not compressed, return original data
	if !compressed.IsCompressed {
		// Verify checksum
		checksum := crc32.ChecksumIEEE(compressed.Data)
		if checksum != compressed.Checksum {
			c.stats.mu.Lock()
			c.stats.CorruptionCount++
			c.stats.mu.Unlock()

			return nil, fmt.Errorf("checksum mismatch: expected %d, got %d (data corruption detected)",
				compressed.Checksum, checksum)
		}

		c.stats.mu.Lock()
		c.stats.DecompressionCount++
		c.stats.mu.Unlock()

		return compressed.Data, nil
	}

	// Read checksum from first 4 bytes
	if len(compressed.Data) < 4 {
		return nil, fmt.Errorf("compressed data too short")
	}

	storedChecksum := binary.LittleEndian.Uint32(compressed.Data[:4])

	// Create gzip reader
	reader, err := gzip.NewReader(bytes.NewReader(compressed.Data[4:]))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	// Read decompressed data
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	// Verify checksum
	checksum := crc32.ChecksumIEEE(data)
	if checksum != storedChecksum {
		c.stats.mu.Lock()
		c.stats.CorruptionCount++
		c.stats.mu.Unlock()

		return nil, fmt.Errorf("checksum mismatch after decompression: expected %d, got %d (data corruption detected)",
			storedChecksum, checksum)
	}

	// Verify size
	if int64(len(data)) != compressed.OriginalSize {
		c.stats.mu.Lock()
		c.stats.CorruptionCount++
		c.stats.mu.Unlock()

		return nil, fmt.Errorf("size mismatch: expected %d, got %d (data corruption detected)",
			compressed.OriginalSize, len(data))
	}

	// Update statistics
	c.stats.mu.Lock()
	c.stats.DecompressionCount++
	c.stats.mu.Unlock()

	return data, nil
}

// GetCompressionRatio returns the average compression ratio
func (c *Compressor) GetCompressionRatio() float64 {
	c.stats.mu.RLock()
	defer c.stats.mu.RUnlock()

	if c.stats.TotalOriginalBytes == 0 {
		return 0
	}

	saved := c.stats.TotalOriginalBytes - c.stats.TotalCompressedBytes
	return float64(saved) / float64(c.stats.TotalOriginalBytes) * 100
}

// GetStats returns compression statistics
func (c *Compressor) GetStats() CompressionStats {
	c.stats.mu.RLock()
	defer c.stats.mu.RUnlock()

	return CompressionStats{
		TotalOriginalBytes:   c.stats.TotalOriginalBytes,
		TotalCompressedBytes: c.stats.TotalCompressedBytes,
		CompressionCount:     c.stats.CompressionCount,
		DecompressionCount:   c.stats.DecompressionCount,
		CorruptionCount:      c.stats.CorruptionCount,
	}
}

// ShouldCompress determines if data should be compressed based on size
func (c *Compressor) ShouldCompress(data []byte) bool {
	return len(data) >= c.minSize
}

// CalculateChecksum calculates CRC32 checksum of data
func CalculateChecksum(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

// VerifyChecksum verifies data against expected checksum
func VerifyChecksum(data []byte, expectedChecksum uint32) bool {
	actualChecksum := crc32.ChecksumIEEE(data)
	return actualChecksum == expectedChecksum
}

// CompressWithLevel compresses data with a specific level (utility function)
func CompressWithLevel(data []byte, level int) ([]byte, error) {
	var buf bytes.Buffer

	writer, err := gzip.NewWriterLevel(&buf, level)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip writer: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to compress data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Decompress decompresses GZIP data (utility function)
func Decompress(compressedData []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	return data, nil
}
