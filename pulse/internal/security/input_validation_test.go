package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func intPtr(i int) *int {
	return &i
}

func TestValidateInput_RequiredField(t *testing.T) {
	data := map[string]interface{}{}
	rules := []ValidationRule{
		{Field: "name", Type: "string", Required: true},
	}

	result := ValidateInput(data, rules)
	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "name", result.Errors[0].Field)
}

func TestValidateInput_MissingOptionalField(t *testing.T) {
	data := map[string]interface{}{}
	rules := []ValidationRule{
		{Field: "description", Type: "string", Required: false},
	}

	result := ValidateInput(data, rules)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestValidateInput_NilValue(t *testing.T) {
	data := map[string]interface{}{"name": nil}
	rules := []ValidationRule{
		{Field: "name", Type: "string", Required: false},
	}

	result := ValidateInput(data, rules)
	assert.True(t, result.Valid)
}

func TestValidateInput_StringType(t *testing.T) {
	minLen := 3
	maxLen := 10

	data := map[string]interface{}{"name": "ab"}
	rules := []ValidationRule{
		{Field: "name", Type: "string", Required: true, MinLength: &minLen, MaxLength: &maxLen},
	}

	result := ValidateInput(data, rules)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0].Message, "at least 3 characters")
}

func TestValidateInput_StringTooLong(t *testing.T) {
	maxLen := 5
	data := map[string]interface{}{"name": "toolongname"}
	rules := []ValidationRule{
		{Field: "name", Type: "string", MaxLength: &maxLen},
	}

	result := ValidateInput(data, rules)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0].Message, "most 5 characters")
}

func TestValidateInput_StringSanitize(t *testing.T) {
	minLen := 1
	data := map[string]interface{}{"name": "  hello  "}
	rules := []ValidationRule{
		{Field: "name", Type: "string", Sanitize: true, MinLength: &minLen},
	}

	result := ValidateInput(data, rules)
	assert.True(t, result.Valid)
}

func TestValidateInput_IntType(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		valid bool
	}{
		{"valid int", 50, true},
		{"int64", int64(50), true},
		{"float64", float64(50), true},
		{"not int", "string", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]interface{}{"count": tt.value}
			rules := []ValidationRule{
				{Field: "count", Type: "int"},
			}
			result := ValidateInput(data, rules)
			assert.Equal(t, tt.valid, result.Valid)
		})
	}
}

func TestValidateInput_UUIDType(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		valid bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", true},
		{"invalid uuid", "not-a-uuid", false},
		{"not string", 123, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]interface{}{"id": tt.value}
			rules := []ValidationRule{
				{Field: "id", Type: "uuid"},
			}
			result := ValidateInput(data, rules)
			assert.Equal(t, tt.valid, result.Valid, tt.name)
		})
	}
}

func TestValidateInput_EmailType(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		valid bool
	}{
		{"valid email", "user@example.com", true},
		{"invalid email", "not-an-email", false},
		{"not string", 123, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]interface{}{"email": tt.value}
			rules := []ValidationRule{
				{Field: "email", Type: "email"},
			}
			result := ValidateInput(data, rules)
			assert.Equal(t, tt.valid, result.Valid, tt.name)
		})
	}
}

func TestValidateInput_URLType(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		valid bool
	}{
		{"valid https url", "https://example.com", true},
		{"valid http url", "http://example.com", true},
		{"invalid url", "://invalid", false},
		{"not string", 123, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]interface{}{"url": tt.value}
			rules := []ValidationRule{
				{Field: "url", Type: "url"},
			}
			result := ValidateInput(data, rules)
			assert.Equal(t, tt.valid, result.Valid, tt.name)
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		// SanitizeString only removes control chars and trims whitespace - no HTML escaping
		{"<script>", "<script>"},
		{"normal text", "normal text"},
		{"line1\nline2", "line1\nline2"}, // newlines preserved
	}

	for _, tt := range tests {
		result := SanitizeString(tt.input)
		assert.Equal(t, tt.expected, result, "Input: %q", tt.input)
	}
}

func TestValidateIPAddress(t *testing.T) {
	tests := []struct {
		ip    string
		valid bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"::1", true},
		{"2001:db8::1", true},
		{"not-an-ip", false},
		{"", false},
		{"999.999.999.999", false},
	}

	for _, tt := range tests {
		err := ValidateIPAddress(tt.ip)
		if tt.valid {
			assert.NoError(t, err, "IP %q should be valid", tt.ip)
		} else {
			assert.Error(t, err, "IP %q should be invalid", tt.ip)
		}
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password - 12 chars with all requirements", "Secure@123456", false},
		{"too short - less than 12 chars", "Abc1@xyz", true},
		{"no uppercase", "secure@123456", true},
		{"no lowercase", "SECURE@123456", true},
		{"no digit", "Secure@abcdef", true},
		{"no special char", "SecureAbc1234", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if tt.wantErr {
				assert.Error(t, err, tt.name)
			} else {
				assert.NoError(t, err, tt.name)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"valid username", "john_doe", false},
		{"valid alphanumeric", "user123", false},
		{"valid with hyphen", "user-name", false},
		{"too short", "ab", true},
		{"too long", "a_very_long_username_that_exceeds_the_limit_of_fifty_characters", true},
		{"with spaces", "user name", true},
		{"with special chars", "user@name", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if tt.wantErr {
				assert.Error(t, err, tt.name)
			} else {
				assert.NoError(t, err, tt.name)
			}
		})
	}
}
