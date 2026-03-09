package security

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"unicode"
)

// ValidationRule defines a rule for input validation
type ValidationRule struct {
	Field     string
	Type      string // "string", "int", "uuid", "email", "url"
	Required  bool
	MinLength *int
	MaxLength *int
	Pattern   *string // Regex pattern
	Enum      []string
	Sanitize  bool // Trim whitespace, etc.
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// ValidationResult contains the results of validation
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// ValidateInput validates input data against a set of rules
func ValidateInput(data map[string]interface{}, rules []ValidationRule) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []ValidationError{}}

	for _, rule := range rules {
		value, exists := data[rule.Field]

		// Check required fields
		if rule.Required && !exists {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   rule.Field,
				Message: fmt.Sprintf("Field '%s' is required", rule.Field),
			})
			continue
		}

		// Skip non-required fields that don't exist
		if !exists || value == nil {
			continue
		}

		// Type-specific validation
		var err error
		switch rule.Type {
		case "string":
			err = validateString(value, rule)
		case "int":
			err = validateInt(value, rule)
		case "uuid":
			err = validateUUID(value, rule)
		case "email":
			err = validateEmail(value, rule)
		case "url":
			err = validateURL(value, rule)
		}

		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   rule.Field,
				Message: err.Error(),
			})
		}
	}

	return result
}

func validateString(value interface{}, rule ValidationRule) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("Field '%s' must be a string", rule.Field)
	}

	// Sanitize if requested
	if rule.Sanitize {
		str = strings.TrimSpace(str)
	}

	// Check min length
	if rule.MinLength != nil && len(str) < *rule.MinLength {
		return fmt.Errorf("Field '%s' must be at least %d characters", rule.Field, *rule.MinLength)
	}

	// Check max length
	if rule.MaxLength != nil && len(str) > *rule.MaxLength {
		return fmt.Errorf("Field '%s' must be at most %d characters", rule.Field, *rule.MaxLength)
	}

	// Check pattern
	if rule.Pattern != nil {
		matched, err := regexp.MatchString(*rule.Pattern, str)
		if err != nil {
			return fmt.Errorf("Invalid pattern for field '%s'", rule.Field)
		}
		if !matched {
			return fmt.Errorf("Field '%s' does not match required pattern", rule.Field)
		}
	}

	// Check enum
	if len(rule.Enum) > 0 {
		found := false
		for _, allowed := range rule.Enum {
			if str == allowed {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Field '%s' must be one of: %v", rule.Field, rule.Enum)
		}
	}

	return nil
}

func validateInt(value interface{}, rule ValidationRule) error {
	var num int
	switch v := value.(type) {
	case int:
		num = v
	case int64:
		num = int(v)
	case float64:
		num = int(v)
	default:
		return fmt.Errorf("Field '%s' must be an integer", rule.Field)
	}
	_ = num // Use num
	return nil
}

func validateUUID(value interface{}, rule ValidationRule) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("Field '%s' must be a string", rule.Field)
	}

	// Basic UUID format validation
	uuidPattern := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !uuidPattern.MatchString(str) {
		return fmt.Errorf("Field '%s' must be a valid UUID", rule.Field)
	}

	return nil
}

func validateEmail(value interface{}, rule ValidationRule) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("Field '%s' must be a string", rule.Field)
	}

	// Basic email validation
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailPattern.MatchString(str) {
		return fmt.Errorf("Field '%s' must be a valid email address", rule.Field)
	}

	return nil
}

func validateURL(value interface{}, rule ValidationRule) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("Field '%s' must be a string", rule.Field)
	}

	// Basic URL validation
	urlPattern := regexp.MustCompile(`^https?://[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;=]+$`)
	if !urlPattern.MatchString(str) {
		return fmt.Errorf("Field '%s' must be a valid URL", rule.Field)
	}

	return nil
}

// SanitizeString removes potentially dangerous characters from strings
func SanitizeString(input string) string {
	// Remove null bytes and other control characters except newline and tab
	var sb strings.Builder
	for _, r := range input {
		if r == '\n' || r == '\t' || r == '\r' {
			sb.WriteRune(r)
		} else if !unicode.IsControl(r) {
			sb.WriteRune(r)
		}
	}
	return strings.TrimSpace(sb.String())
}

// ValidateIPAddress checks if a string is a valid IP address
func ValidateIPAddress(ipStr string) error {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", ipStr)
	}
	return nil
}

// ValidatePasswordStrength checks password strength
func ValidatePasswordStrength(password string) error {
	if len(password) < 12 {
		return fmt.Errorf("password must be at least 12 characters")
	}
	if len(password) > 128 {
		return fmt.Errorf("password must be at most 128 characters")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// ValidateUsername checks if a username is valid
func ValidateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters")
	}
	if len(username) > 50 {
		return fmt.Errorf("username must be at most 50 characters")
	}

	// Username pattern: alphanumeric, underscore, hyphen only
	usernamePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernamePattern.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, underscores, and hyphens")
	}

	return nil
}
