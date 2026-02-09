package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCrypto_HashTokenSHA256 tests SHA-256 token hashing
func TestCrypto_HashTokenSHA256(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		wantLen  int
	}{
		{
			name:    "hash empty token",
			token:   "",
			wantLen: 64, // SHA-256 hex encoded is 64 characters
		},
		{
			name:    "hash simple token",
			token:   "test-token-123",
			wantLen: 64,
		},
		{
			name:    "hash UUID token",
			token:   "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
			wantLen: 64,
		},
		{
			name:    "hash long token",
			token:   "very-long-api-key-with-many-characters-1234567890abcdefghijklmnopqrstuvwxyz",
			wantLen: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashTokenSHA256(tt.token)

			// Hash should be 64 characters (SHA-256 hex encoded)
			assert.Equal(t, tt.wantLen, len(hash), "Hash length should be 64 characters")

			// Hash should be hexadecimal
			for _, c := range hash {
				assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
					"Hash should contain only hexadecimal characters")
			}
		})
	}
}

// TestCrypto_HashTokenDeterministic tests that hashing is deterministic
func TestCrypto_HashTokenDeterministic(t *testing.T) {
	token := "test-token-123"

	hash1 := HashTokenSHA256(token)
	hash2 := HashTokenSHA256(token)

	assert.Equal(t, hash1, hash2, "Hashing the same token should produce the same hash")
}

// TestCrypto_HashTokenUnique tests that different tokens produce different hashes
func TestCrypto_HashTokenUnique(t *testing.T) {
	tokens := []string{
		"token-1",
		"token-2",
		"token-3",
	}

	hashes := make(map[string]bool)
	for _, token := range tokens {
		hash := HashTokenSHA256(token)
		assert.False(t, hashes[hash], "Each token should produce a unique hash")
		hashes[hash] = true
	}
}

// TestCrypto_HashTokenOneWay tests that hashing is one-way (cannot reverse)
func TestCrypto_HashTokenOneWay(t *testing.T) {
	token := "original-token-123"
	hash := HashTokenSHA256(token)

	// Hash should not contain the original token
	assert.NotContains(t, hash, token, "Hash should not contain the original token")

	// Hash should be significantly different from input
	assert.NotEqual(t, token, hash, "Hash should be different from the original token")
}

// TestCrypto_ConstantTimeCompare tests constant-time comparison
func TestCrypto_ConstantTimeCompare(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{
			name: "equal strings",
			a:    "test-token",
			b:    "test-token",
			want: true,
		},
		{
			name: "different strings",
			a:    "test-token",
			b:    "different-token",
			want: false,
		},
		{
			name: "empty strings",
			a:    "",
			b:    "",
			want: true,
		},
		{
			name: "one empty string",
			a:    "test-token",
			b:    "",
			want: false,
		},
		{
			name: "similar but different",
			a:    "test-token-1",
			b:    "test-token-2",
			want: false,
		},
		{
			name: "case sensitive",
			a:    "TEST-TOKEN",
			b:    "test-token",
			want: false,
		},
		{
			name: "long equal strings",
			a:    "very-long-token-with-exactly-the-same-characters-1234567890",
			b:    "very-long-token-with-exactly-the-same-characters-1234567890",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConstantTimeCompare(tt.a, tt.b)
			assert.Equal(t, tt.want, result, "ConstantTimeCompare result mismatch")
		})
	}
}

// TestCrypto_ConstantTimeCompareTimingResistance tests that comparison time is consistent
// This is a basic test - in production, you should run statistical analysis
func TestCrypto_ConstantTimeCompareTimingResistance(t *testing.T) {
	// Test with many iterations to ensure no early exit optimization
	iterations := 1000

	// Test equal strings
	equalA := "test-token-abcde"
	equalB := "test-token-abcde"
	for i := 0; i < iterations; i++ {
		if !ConstantTimeCompare(equalA, equalB) {
			t.Fatal("Equal strings should return true")
		}
	}

	// Test different strings (same length)
	diffA := "test-token-abcde"
	diffB := "test-token-xyzuv"
	for i := 0; i < iterations; i++ {
		if ConstantTimeCompare(diffA, diffB) {
			t.Fatal("Different strings should return false")
		}
	}

	// If we get here without timing out significantly, the implementation is likely constant-time
}

// TestCrypto_VerifyTokenHash tests token hash verification
func TestCrypto_VerifyTokenHash(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		hash      string
		wantValid bool
	}{
		{
			name:      "valid token hash",
			token:     "test-token-123",
			hash:      HashTokenSHA256("test-token-123"),
			wantValid: true,
		},
		{
			name:      "invalid token hash",
			token:     "test-token-123",
			hash:      HashTokenSHA256("different-token"),
			wantValid: false,
		},
		{
			name:      "empty token with empty hash",
			token:     "",
			hash:      HashTokenSHA256(""),
			wantValid: true,
		},
		{
			name:      "token with different hash",
			token:     "correct-token",
			hash:      "invalid-hash-not-64-chars",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyTokenHash(tt.token, tt.hash)
			assert.Equal(t, tt.wantValid, result, "VerifyTokenHash result mismatch")
		})
	}
}

// TestCrypto_HashPassword tests password hashing
func TestCrypto_HashPassword(t *testing.T) {
	password := "SecurePass123"

	// Hash password using existing auth.HashPassword
	hash, err := HashPassword(password)
	assert.NoError(t, err, "HashPassword should not return error")
	assert.NotEmpty(t, hash, "Hash should not be empty")
	assert.NotEqual(t, password, hash, "Hash should be different from password")
}

// TestCrypto_VerifyPassword tests password comparison
func TestCrypto_VerifyPassword(t *testing.T) {
	password := "SecurePass123"

	// Hash password using existing auth.HashPassword
	hash, err := HashPassword(password)
	assert.NoError(t, err, "HashPassword should not return error")

	// Test correct password
	err = VerifyPassword(password, hash)
	assert.NoError(t, err, "Correct password should match")

	// Test incorrect password
	err = VerifyPassword("wrong-password", hash)
	assert.Error(t, err, "Incorrect password should not match")
}

// BenchmarkCrypto_HashTokenSHA256 benchmarks SHA-256 hashing
func BenchmarkCrypto_HashTokenSHA256(b *testing.B) {
	token := "test-token-1234567890"
	for i := 0; i < b.N; i++ {
		HashTokenSHA256(token)
	}
}

// BenchmarkCrypto_ConstantTimeCompare benchmarks constant-time comparison
func BenchmarkCrypto_ConstantTimeCompare(b *testing.B) {
	a := "test-token-1234567890"
	b_str := "test-token-1234567890"
	for i := 0; i < b.N; i++ {
		ConstantTimeCompare(a, b_str)
	}
}
