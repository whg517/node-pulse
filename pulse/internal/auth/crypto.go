package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// HashTokenSHA256 generates a SHA-256 hash of a token (for refresh tokens, API keys)
// This is a one-way hash - the original token cannot be recovered
func HashTokenSHA256(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// ConstantTimeCompare compares two strings in constant time
// This prevents timing attacks when comparing tokens, API keys, or hashes
// Returns true if the strings are equal, false otherwise
func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// VerifyTokenHash verifies a token against its SHA-256 hash
// Returns true if the token matches the hash
func VerifyTokenHash(token, tokenHash string) bool {
	return ConstantTimeCompare(HashTokenSHA256(token), tokenHash)
}
