package reporter

import (
	"bytes"
	"testing"
)

func TestNewCompressor(t *testing.T) {
	compressor := NewCompressor(6, 1024)
	if compressor == nil {
		t.Fatal("Expected non-nil compressor")
	}
	if compressor.level != 6 {
		t.Errorf("Expected level 6, got %d", compressor.level)
	}
	if compressor.minSize != 1024 {
		t.Errorf("Expected minSize 1024, got %d", compressor.minSize)
	}
}

func TestNewCompressor_DefaultLevel(t *testing.T) {
	// Invalid level should default to 6
	compressor := NewCompressor(0, 0)
	if compressor.level != 6 {
		t.Errorf("Expected default level 6, got %d", compressor.level)
	}

	// Invalid level (too high)
	compressor = NewCompressor(15, 0)
	if compressor.level != 6 {
		t.Errorf("Expected default level 6 for invalid level, got %d", compressor.level)
	}
}

func TestCompressor_CompressDecompress(t *testing.T) {
	compressor := NewCompressor(6, 100)

	originalData := []byte("This is a test string that should compress well. " +
		"Repeated data compresses better. " +
		"This is a test string that should compress well. " +
		"Repeated data compresses better.")

	compressed, err := compressor.Compress(originalData)
	if err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}

	// Verify compression happened
	if !compressed.IsCompressed {
		t.Error("Expected data to be marked as compressed")
	}

	// Verify checksum was calculated
	if compressed.Checksum == 0 {
		t.Error("Expected checksum to be calculated")
	}

	// Decompress
	decompressed, err := compressor.Decompress(compressed)
	if err != nil {
		t.Fatalf("Failed to decompress: %v", err)
	}

	// Verify data integrity
	if !bytes.Equal(originalData, decompressed) {
		t.Error("Decompressed data does not match original")
	}
}

func TestCompressor_CompressionRatio(t *testing.T) {
	compressor := NewCompressor(6, 100)

	// Compress repetitive data
	data := bytes.Repeat([]byte("ABCDEFGH"), 100) // 800 bytes

	compressed, err := compressor.Compress(data)
	if err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}

	// Compression ratio should be significant
	ratio := compressor.GetCompressionRatio()
	if ratio < 50 {
		t.Errorf("Expected compression ratio >= 50%%, got %.2f%%", ratio)
	}

	t.Logf("Original: %d bytes, Compressed: %d bytes, Ratio: %.2f%%",
		compressed.OriginalSize, len(compressed.Data), ratio)
}

func TestCompressor_SmallDataNotCompressed(t *testing.T) {
	compressor := NewCompressor(6, 1024) // Min size 1KB

	// Data smaller than min size
	smallData := []byte("small")

	compressed, err := compressor.Compress(smallData)
	if err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}

	// Should not be compressed (too small)
	if compressed.IsCompressed {
		t.Error("Expected small data not to be compressed")
	}

	// Data should be unchanged
	if !bytes.Equal(smallData, compressed.Data) {
		t.Error("Expected uncompressed data to be unchanged")
	}
}

func TestCompressor_EmptyData(t *testing.T) {
	compressor := NewCompressor(6, 0)

	compressed, err := compressor.Compress([]byte{})
	if err != nil {
		t.Fatalf("Failed to compress empty data: %v", err)
	}

	if compressed.IsCompressed {
		t.Error("Empty data should not be marked as compressed")
	}

	if compressed.OriginalSize != 0 {
		t.Errorf("Expected original size 0, got %d", compressed.OriginalSize)
	}
}

func TestCompressor_DecompressNil(t *testing.T) {
	compressor := NewCompressor(6, 0)

	_, err := compressor.Decompress(nil)
	if err == nil {
		t.Error("Expected error when decompressing nil data")
	}
}

func TestCompressor_ChecksumVerification(t *testing.T) {
	compressor := NewCompressor(6, 100)

	data := []byte("test data for checksum verification")

	compressed, _ := compressor.Compress(data)

	// Corrupt the checksum
	compressed.Checksum = 99999

	// Decompression should fail due to checksum mismatch
	_, err := compressor.Decompress(compressed)
	if err == nil {
		t.Error("Expected error due to checksum mismatch")
	}

	// Verify corruption was tracked
	stats := compressor.GetStats()
	if stats.CorruptionCount == 0 {
		t.Error("Expected corruption to be tracked")
	}
}

func TestCompressor_ShouldCompress(t *testing.T) {
	compressor := NewCompressor(6, 100)

	// Small data
	if compressor.ShouldCompress([]byte("small")) {
		t.Error("Expected small data not to need compression")
	}

	// Large data
	largeData := bytes.Repeat([]byte("x"), 200)
	if !compressor.ShouldCompress(largeData) {
		t.Error("Expected large data to need compression")
	}
}

func TestCalculateChecksum(t *testing.T) {
	data := []byte("test data")

	checksum1 := CalculateChecksum(data)
	checksum2 := CalculateChecksum(data)

	// Same data should produce same checksum
	if checksum1 != checksum2 {
		t.Error("Expected same checksum for same data")
	}

	// Different data should produce different checksum
	differentData := []byte("different data")
	checksum3 := CalculateChecksum(differentData)
	if checksum1 == checksum3 {
		t.Error("Expected different checksum for different data")
	}
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("test data")
	checksum := CalculateChecksum(data)

	// Valid checksum
	if !VerifyChecksum(data, checksum) {
		t.Error("Expected checksum verification to pass")
	}

	// Invalid checksum
	if VerifyChecksum(data, 99999) {
		t.Error("Expected checksum verification to fail for invalid checksum")
	}

	// Modified data
	modifiedData := []byte("modified data")
	if VerifyChecksum(modifiedData, checksum) {
		t.Error("Expected checksum verification to fail for modified data")
	}
}

func TestCompressWithLevel(t *testing.T) {
	data := []byte("test data for compression level test")

	// Test all compression levels
	for level := 1; level <= 9; level++ {
		compressed, err := CompressWithLevel(data, level)
		if err != nil {
			t.Errorf("Failed to compress with level %d: %v", level, err)
		}

		if len(compressed) == 0 {
			t.Errorf("Expected non-empty compressed data for level %d", level)
		}
	}
}

func TestDecompress(t *testing.T) {
	data := []byte("test data for decompression test")

	compressed, err := CompressWithLevel(data, 6)
	if err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}

	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Failed to decompress: %v", err)
	}

	if !bytes.Equal(data, decompressed) {
		t.Error("Decompressed data does not match original")
	}
}

func TestCompressor_GetStats(t *testing.T) {
	compressor := NewCompressor(6, 0)

	// Perform some operations
	data := bytes.Repeat([]byte("test"), 100)
	compressed, _ := compressor.Compress(data)
	_, _ = compressor.Decompress(compressed)

	stats := compressor.GetStats()

	if stats.CompressionCount != 1 {
		t.Errorf("Expected compression count 1, got %d", stats.CompressionCount)
	}

	if stats.DecompressionCount != 1 {
		t.Errorf("Expected decompression count 1, got %d", stats.DecompressionCount)
	}

	if stats.TotalOriginalBytes == 0 {
		t.Error("Expected non-zero original bytes")
	}
}
