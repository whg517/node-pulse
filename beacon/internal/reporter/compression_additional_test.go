package reporter

import (
"encoding/binary"
"testing"
)

// TestCompressor_Decompress_NotCompressed tests Decompress with uncompressed data
func TestCompressor_Decompress_NotCompressed(t *testing.T) {
c := NewCompressor(6, 0) // minSize=0 forces compression threshold to default

// Create uncompressed data - should work with correct checksum
data := []byte("hello world test data")
// Use CompressedData with IsCompressed=false and correct checksum
checksum := CalculateChecksum(data)
compressed := &CompressedData{
Data:         data,
IsCompressed: false,
Checksum:     checksum,
OriginalSize: int64(len(data)),
}

// Decompress should succeed
result, err := c.Decompress(compressed)
if err != nil {
t.Fatalf("Decompress failed: %v", err)
}
if string(result) != string(data) {
t.Errorf("Expected %q, got %q", string(data), string(result))
}
}

// TestCompressor_Decompress_ChecksumMismatch_Uncompressed tests checksum mismatch for uncompressed
func TestCompressor_Decompress_ChecksumMismatch_Uncompressed(t *testing.T) {
c := NewCompressor(6, 0)

// Create data with wrong checksum
compressed := &CompressedData{
Data:         []byte("corrupted data"),
IsCompressed: false,
Checksum:     12345, // Wrong checksum
OriginalSize: 14,
}

_, err := c.Decompress(compressed)
if err == nil {
t.Error("Expected checksum mismatch error")
}
}

// TestCompressor_Decompress_TooShort tests decompress with data too short
func TestCompressor_Decompress_TooShort(t *testing.T) {
c := NewCompressor(6, 0)

// Data that is compressed but too short (< 4 bytes)
compressed := &CompressedData{
Data:         []byte{1, 2}, // Only 2 bytes - too short
IsCompressed: true,
Checksum:     0,
OriginalSize: 10,
}

_, err := c.Decompress(compressed)
if err == nil {
t.Error("Expected error for too short compressed data")
}
}

// TestCompressor_Decompress_SizeMismatch tests decompression with size mismatch
func TestCompressor_Decompress_SizeMismatch(t *testing.T) {
c := NewCompressor(6, 0)

data := []byte("test data that will be compressed to trigger size check")
// Make a large enough data block to trigger compression
bigData := make([]byte, 1100)
copy(bigData, data)
compressed, err := c.Compress(bigData)
if err != nil {
t.Fatalf("Compress failed: %v", err)
}

if !compressed.IsCompressed {
t.Skip("Data was not compressed - skipping size mismatch test")
}

// Tamper with the original size to cause size mismatch
compressed.OriginalSize = int64(len(bigData)) + 10

_, err = c.Decompress(compressed)
if err == nil {
t.Error("Expected size mismatch error")
}
}

// TestCompressor_Decompress_ChecksumMismatch_AfterDecompression tests checksum mismatch after decompression
func TestCompressor_Decompress_ChecksumMismatch_AfterDecompression(t *testing.T) {
c := NewCompressor(6, 0)

// Use large enough data to trigger compression (default minSize)
data := make([]byte, 1100)
for i := range data {
data[i] = byte(i % 26)
}

compressed, err := c.Compress(data)
if err != nil {
t.Fatalf("Compress failed: %v", err)
}

if !compressed.IsCompressed {
t.Skip("Data was not compressed - skipping checksum mismatch test")
}

// Tamper with the stored checksum in the first 4 bytes
if len(compressed.Data) >= 4 {
// Set a wrong checksum
binary.LittleEndian.PutUint32(compressed.Data[:4], 0)
}

_, err = c.Decompress(compressed)
if err == nil {
t.Error("Expected checksum mismatch error after decompression")
}
}

// TestCompressor_GetCompressionRatio_WithData tests GetCompressionRatio with data
func TestCompressor_GetCompressionRatio_WithData(t *testing.T) {
c := NewCompressor(6, 0)

// With no data, ratio should be 0
if c.GetCompressionRatio() != 0 {
t.Error("Expected 0 ratio with no data")
}

// Compress some data to get a non-zero ratio
data := make([]byte, 1100)
for i := range data {
data[i] = byte(i % 26)
}
_, err := c.Compress(data)
if err != nil {
t.Fatalf("Compress failed: %v", err)
}

// Now there should be a compression ratio
ratio := c.GetCompressionRatio()
t.Logf("Compression ratio: %f%%", ratio)
}

// TestCompressWithLevel_Error tests CompressWithLevel with invalid level
func TestCompressWithLevel_Error(t *testing.T) {
data := []byte("test data")
_, err := CompressWithLevel(data, 10) // Invalid level
if err == nil {
t.Error("Expected error for invalid compression level")
}
}

// TestDecompress_InvalidData tests Decompress with invalid gzip data
func TestDecompress_InvalidData(t *testing.T) {
invalidData := []byte("this is not valid gzip data")
_, err := Decompress(invalidData)
if err == nil {
t.Error("Expected error for invalid gzip data")
}
}
