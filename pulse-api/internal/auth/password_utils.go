package auth

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrPasswordTooShort is returned when password is less than 8 characters
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	// ErrPasswordTooLong is returned when password is more than 32 characters
	ErrPasswordTooLong = errors.New("password must be less than 32 characters")
	// ErrPasswordMissingUppercase is returned when password has no uppercase letter
	ErrPasswordMissingUppercase = errors.New("password must contain at least one uppercase letter")
	// ErrPasswordMissingLowercase is returned when password has no lowercase letter
	ErrPasswordMissingLowercase = errors.New("password must contain at least one lowercase letter")
	// ErrPasswordMissingDigit is returned when password has no digit
	ErrPasswordMissingDigit = errors.New("password must contain at least one digit")
)

// ValidatePassword validates password strength according to security requirements:
// - Minimum 8 characters
// - Maximum 32 characters
// - At least one uppercase letter (A-Z)
// - At least one lowercase letter (a-z)
// - At least one digit (0-9)
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	if len(password) > 32 {
		return ErrPasswordTooLong
	}

	var hasUpper, hasLower, hasDigit bool

	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsLower(char) {
			hasLower = true
		} else if unicode.IsDigit(char) {
			hasDigit = true
		}

		if hasUpper && hasLower && hasDigit {
			break
		}
	}

	if !hasUpper {
		return ErrPasswordMissingUppercase
	}

	if !hasLower {
		return ErrPasswordMissingLowercase
	}

	if !hasDigit {
		return ErrPasswordMissingDigit
	}

	return nil
}

// HashPassword hashes a plain text password using bcrypt with cost factor 12
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword compares a plain text password with a bcrypt hash
// Returns nil if password matches, error if it doesn't
func VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
