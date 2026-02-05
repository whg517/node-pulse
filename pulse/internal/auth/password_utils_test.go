package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// TestValidatePassword_ValidPasswords tests valid passwords that meet all requirements
func TestValidatePassword_ValidPasswords(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "Minimum valid password",
			password: "Passw0rd",
		},
		{
			name:     "Password with special characters",
			password: "P@ssw0rd!#$",
		},
		{
			name:     "Long password within limit",
			password: "ThisIsAVeryLongPassword123",
		},
		{
			name:     "Password with numbers in middle",
			password: "Pass123word",
		},
		{
			name:     "Exactly 8 characters",
			password: "Test1234",
		},
		{
			name:     "Exactly 32 characters",
			password: "Abcdefgh12345678NineTenChar2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := ValidatePassword(tt.password)

			// Assert
			assert.NoError(t, err, "Password should be valid: %s", tt.password)
		})
	}
}

// TestValidatePassword_TooShort tests passwords shorter than 8 characters
func TestValidatePassword_TooShort(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "7 characters",
			password: "Passw0r",
		},
		{
			name:     "1 character",
			password: "A",
		},
		{
			name:     "Empty password",
			password: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := ValidatePassword(tt.password)

			// Assert
			assert.Error(t, err, "Password should be too short")
			assert.Equal(t, ErrPasswordTooShort, err, "Should return ErrPasswordTooShort")
		})
	}
}

// TestValidatePassword_TooLong tests passwords longer than 32 characters
func TestValidatePassword_TooLong(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "33 characters",
			password: "ThisPasswordIsExactly33CharactersLong1",
		},
		{
			name:     "Very long password",
			password: strings.Repeat("Aa1", 20), // 60 characters
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := ValidatePassword(tt.password)

			// Assert
			assert.Error(t, err, "Password should be too long")
			assert.Equal(t, ErrPasswordTooLong, err, "Should return ErrPasswordTooLong")
		})
	}
}

// TestValidatePassword_MissingUppercase tests passwords without uppercase letters
func TestValidatePassword_MissingUppercase(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "All lowercase",
			password: "password123",
		},
		{
			name:     "Lowercase and numbers",
			password: "test123456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := ValidatePassword(tt.password)

			// Assert
			assert.Error(t, err, "Password should require uppercase")
			assert.Equal(t, ErrPasswordMissingUppercase, err, "Should return ErrPasswordMissingUppercase")
		})
	}
}

// TestValidatePassword_MissingLowercase tests passwords without lowercase letters
func TestValidatePassword_MissingLowercase(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "All uppercase",
			password: "PASSWORD123",
		},
		{
			name:     "Uppercase and numbers",
			password: "TEST123456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := ValidatePassword(tt.password)

			// Assert
			assert.Error(t, err, "Password should require lowercase")
			assert.Equal(t, ErrPasswordMissingLowercase, err, "Should return ErrPasswordMissingLowercase")
		})
	}
}

// TestValidatePassword_MissingDigit tests passwords without digits
func TestValidatePassword_MissingDigit(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "Only letters",
			password: "Password",
		},
		{
			name:     "Mixed case without numbers",
			password: "MyPassword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := ValidatePassword(tt.password)

			// Assert
			assert.Error(t, err, "Password should require digit")
			assert.Equal(t, ErrPasswordMissingDigit, err, "Should return ErrPasswordMissingDigit")
		})
	}
}

// TestValidatePassword_FirstError tests that validation stops at first error
func TestValidatePassword_FirstError(t *testing.T) {
	tests := []struct {
		name           string
		password       string
		expectedError  error
	}{
		{
			name:          "Too short (checked first)",
			password:      "Ab1",
			expectedError: ErrPasswordTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := ValidatePassword(tt.password)

			// Assert
			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

// TestValidatePassword_MultipleRequirementsMissing tests passwords missing multiple requirements
func TestValidatePassword_MultipleRequirementsMissing(t *testing.T) {
	tests := []struct {
		name     string
		password string
		// Returns first error encountered
		expectedError error
	}{
		{
			name:          "Too short and no uppercase, no digit",
			password:      "abc",
			expectedError: ErrPasswordTooShort,
		},
		{
			name:          "No uppercase, no lowercase",
			password:      "12345678",
			expectedError: ErrPasswordMissingUppercase,
		},
		{
			name:          "No lowercase, no digit",
			password:      "ABCDEFGH",
			expectedError: ErrPasswordMissingLowercase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := ValidatePassword(tt.password)

			// Assert
			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

// TestHashPassword tests password hashing
func TestHashPassword(t *testing.T) {
	// Arrange
	password := "testPassword123"

	// Act
	hash, err := HashPassword(password)

	// Assert
	require.NoError(t, err, "HashPassword should not return error")
	assert.NotEmpty(t, hash, "Hash should not be empty")
	assert.NotEqual(t, password, hash, "Hash should not equal plain text password")

	// Assert - Hash should be bcrypt format (starts with $2a$, $2b$, or $2y$)
	assert.True(t, len(hash) >= 60, "Bcrypt hash should be at least 60 characters")
	prefix := hash[:4]
	assert.True(t, prefix == "$2a$" || prefix == "$2b$" || prefix == "$2y$",
		"Hash should start with bcrypt prefix, got: %s", prefix)
}

// TestHashPassword_DifferentPasswords tests different passwords produce different hashes
func TestHashPassword_DifferentPasswords(t *testing.T) {
	// Arrange
	password1 := "password1"
	password2 := "password2"

	// Act
	hash1, err1 := HashPassword(password1)
	hash2, err2 := HashPassword(password2)

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, hash1, hash2, "Different passwords should produce different hashes")
}

// TestHashPassword_SamePasswordDifferentHashes tests same password produces different hashes (salt)
func TestHashPassword_SamePasswordDifferentHashes(t *testing.T) {
	// Arrange
	password := "samePassword"

	// Act
	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, hash1, hash2, "Same password should produce different hashes due to salt")
}

// TestVerifyPassword tests password verification with correct password
func TestVerifyPassword(t *testing.T) {
	// Arrange
	password := "correctPassword"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	// Act
	err = VerifyPassword(password, hash)

	// Assert
	assert.NoError(t, err, "Correct password should verify successfully")
}

// TestVerifyPassword_WrongPassword tests password verification with incorrect password
func TestVerifyPassword_WrongPassword(t *testing.T) {
	// Arrange
	password := "correctPassword"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	// Act
	err = VerifyPassword("wrongPassword", hash)

	// Assert
	assert.Error(t, err, "Wrong password should fail verification")
	assert.Equal(t, bcrypt.ErrMismatchedHashAndPassword, err,
		"Should return bcrypt mismatched error")
}

// TestVerifyPassword_EmptyPassword tests password verification with empty password
func TestVerifyPassword_EmptyPassword(t *testing.T) {
	// Arrange
	password := ""
	hash, err := HashPassword(password)
	require.NoError(t, err)

	// Act
	err = VerifyPassword(password, hash)

	// Assert
	assert.NoError(t, err, "Empty password should verify if hash was created from empty password")
}

// TestHashPassword_MaxBcryptLength tests bcrypt 72-byte limit behavior
func TestHashPassword_MaxBcryptLength(t *testing.T) {
	// Arrange - bcrypt max length is 72 bytes
	password := strings.Repeat("a", 72)

	// Act
	hash, err := HashPassword(password)

	// Assert
	require.NoError(t, err, "72-byte password should hash successfully")
	assert.NotEmpty(t, hash, "Hash should not be empty")

	// Verify it works
	err = VerifyPassword(password, hash)
	assert.NoError(t, err, "72-byte password should verify successfully")
}

// TestHashPassword_ExceedsBcryptLimit tests password exceeding 72 bytes fails
func TestHashPassword_ExceedsBcryptLimit(t *testing.T) {
	// Arrange - bcrypt max length is 72 bytes, 73 should fail
	password := strings.Repeat("a", 73)

	// Act
	hash, err := HashPassword(password)

	// Assert
	assert.Error(t, err, "Password exceeding 72 bytes should fail")
	assert.Equal(t, bcrypt.ErrPasswordTooLong, err, "Should return bcrypt password too long error")
	assert.Empty(t, hash, "Hash should be empty when password is too long")
}

// TestVerifyPassword_SpecialCharacters tests password with special characters
func TestVerifyPassword_SpecialCharacters(t *testing.T) {
	// Arrange
	password := "P@ssw0rd!#$%^&*()"

	// Act
	hash, err := HashPassword(password)
	require.NoError(t, err)

	err = VerifyPassword(password, hash)

	// Assert
	assert.NoError(t, err, "Password with special characters should verify successfully")
}

// TestHashPassword_EmptyPassword tests hashing empty password
func TestHashPassword_EmptyPassword(t *testing.T) {
	// Arrange
	password := ""

	// Act
	hash, err := HashPassword(password)

	// Assert
	require.NoError(t, err, "Empty password should hash successfully")
	assert.NotEmpty(t, hash, "Hash should not be empty even for empty password")
}

// TestVerifyPassword_CaseSensitive tests password verification is case-sensitive
func TestVerifyPassword_CaseSensitive(t *testing.T) {
	// Arrange
	password := "Password123"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	// Act - Try lowercase version
	err = VerifyPassword("password123", hash)

	// Assert
	assert.Error(t, err, "Password verification should be case-sensitive")
}

// TestVerifyPassword_InvalidHash tests verification with invalid hash format
func TestVerifyPassword_InvalidHash(t *testing.T) {
	// Arrange
	password := "testPassword"
	invalidHash := "not_a_valid_hash"

	// Act
	err := VerifyPassword(password, invalidHash)

	// Assert
	assert.Error(t, err, "Invalid hash should cause verification error")
}
