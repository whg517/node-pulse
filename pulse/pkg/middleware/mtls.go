package middleware

import (
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	// ErrTLSCertificateRequired is returned when mTLS is required but not provided
	ErrTLSCertificateRequired = errors.New("TLS certificate is required")
	// ErrCertificateInvalid is returned when certificate validation fails
	ErrCertificateInvalid = errors.New("certificate validation failed")
)

// TLSConfig holds mTLS configuration
type TLSConfig struct {
	// Mode controls mTLS enforcement: "disabled", "warn", or "strict"
	Mode string
	// CASpecifiedCertificates when true, requires certificate to be signed by internal CA
	CASpecifiedCertificates bool
	// AllowedCNPrefixes is list of allowed Common Name prefixes (empty = any CN allowed)
	AllowedCNPrefixes []string
	// AllowedOUs is list of allowed Organization Units (empty = any OU allowed)
	AllowedOUs []string
	// MinCertExpiryDays is minimum certificate validity in days
	MinCertExpiryDays int
	// TrustedCAPEMs contains PEM-encoded CA certificates (one or more)
	TrustedCAPEMs []string
}

// mTLSConfig holds the global mTLS configuration
var (
	mTLSConfig     *TLSConfig
	mTLSConfigOnce sync.Once
)

// InitMTLSConfig initializes mTLS configuration from environment
// Environment variables:
//
//	PULSE_MTLS_ENABLED=disabled|warn|strict - Enable mTLS enforcement (default: disabled for dev, strict for production)
//	  - "disabled": mTLS is not enforced
//	  - "warn": mTLS violations are logged but requests are allowed (for gradual migration)
//	  - "strict": mTLS is fully enforced (rejects invalid certificates)
//	PULSE_MTLS_CA_CERTS - Path to CA certificates file (PEM format, multiple allowed)
//	PULSE_MTLS_CN_PREFIXES - Comma-separated list of allowed CN prefixes
//	PULSE_MTLS_ALLOWED_OUS - Comma-separated list of allowed OUs
//	PULSE_MTLS_MIN_CERT_EXPIRY_DAYS=7 - Minimum certificate validity in days
//
// Note: For backward compatibility, boolean values true/false/1/0 are also accepted:
//   - true maps to "strict"
//   - false maps to "disabled"
func InitMTLSConfig() {
	mTLSConfigOnce.Do(func() {
		// Get mTLS mode from environment
		mode := getEnvOrDefault("PULSE_MTLS_ENABLED", "")

		// Handle backward compatibility with boolean values
		if mode == "" {
			// Check server mode - mTLS is MANDATORY for production
			serverMode := getEnvOrDefault("PULSE_SERVER_MODE", "debug")
			if serverMode == "production" {
				mode = "strict"
			} else {
				mode = "disabled"
			}
		} else if mode == "true" || mode == "1" || mode == "enabled" {
			// Backward compatibility: boolean true maps to "strict"
			mode = "strict"
		} else if mode == "false" || mode == "0" || mode == "disabled" {
			mode = "disabled"
		}

		// Validate mode value
		if mode != "disabled" && mode != "warn" && mode != "strict" {
			// Invalid mode, default to disabled with a warning
			mode = "disabled"
		}

		// If disabled, no need to load other config
		if mode == "disabled" {
			mTLSConfig = &TLSConfig{Mode: mode}
			return
		}

		// Load trusted CA certificates
		caCerts := loadTrustedCACerts()

		// Get allowed CN prefixes
		cnPrefixes := []string{}
		if cnStr := getEnvOrDefault("PULSE_MTLS_CN_PREFIXES", ""); cnStr != "" {
			cnPrefixes = strings.Split(cnStr, ",")
			for i := range cnPrefixes {
				cnPrefixes[i] = strings.TrimSpace(cnPrefixes[i])
			}
		}

		// Get allowed OUs
		allowedOUs := []string{}
		if ouStr := getEnvOrDefault("PULSE_MTLS_ALLOWED_OUS", ""); ouStr != "" {
			allowedOUs = strings.Split(ouStr, ",")
			for i := range allowedOUs {
				allowedOUs[i] = strings.TrimSpace(allowedOUs[i])
			}
		}

		// Minimum certificate expiry days
		minExpiryDays := getEnvInt("PULSE_MTLS_MIN_CERT_EXPIRY_DAYS", 7)

		mTLSConfig = &TLSConfig{
			Mode:                    mode,
			CASpecifiedCertificates: len(caCerts) > 0,
			AllowedCNPrefixes:       cnPrefixes,
			AllowedOUs:              allowedOUs,
			MinCertExpiryDays:       minExpiryDays,
			TrustedCAPEMs:           caCerts,
		}
	})
}

// loadTrustedCACerts loads trusted CA certificates from environment or file
func loadTrustedCACerts() []string {
	var caCerts []string

	// Check for inline CA certs (newline separated)
	if inlineCerts := getEnvOrDefault("PULSE_MTLS_CA_CERTS", ""); inlineCerts != "" {
		certs := strings.Split(inlineCerts, "\n\n")
		for _, cert := range certs {
			cert = strings.TrimSpace(cert)
			if cert != "" {
				caCerts = append(caCerts, cert)
			}
		}
	}

	// Check for CA cert file path
	if certFile := getEnvOrDefault("PULSE_MTLS_CA_CERT_FILE", ""); certFile != "" {
		// TODO: Load from file in production
		// For now, we'll rely on inline certs or default development setup
	}

	// If no CA certs specified in production, use a default
	if len(caCerts) == 0 {
		// In development, we may allow self-signed certs
		// In production, this should be configured properly
		serverMode := getEnvOrDefault("PULSE_SERVER_MODE", "debug")
		if serverMode != "production" {
			// Development mode - allow self-signed (no CA verification)
			caCerts = []string{} // Empty = no CA verification
		}
	}

	return caCerts
}

// MTLSAuthMiddleware creates a middleware that enforces mTLS for beacons
// This middleware should be applied to beacon-specific routes
//
// Usage:
//
//	beacon.Use(middleware.MTLSAuthMiddleware())
func MTLSAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if mTLS is configured
		if mTLSConfig == nil || mTLSConfig.Mode == "disabled" {
			// mTLS not configured or disabled, skip validation
			c.Next()
			return
		}

		// Get TLS connection state from request
		if c.Request.TLS == nil {
			// No TLS connection
			if mTLSConfig.Mode == "warn" {
				// Warn mode: log warning but allow request
				fmt.Printf("[MTLS WARN] Request from %s without TLS connection (mTLS warn mode)\n", c.ClientIP())
				c.Next()
				return
			}
			// Strict mode: reject request
			c.JSON(http.StatusForbidden, gin.H{
				"code":    "MTLS_REQUIRED",
				"message": "TLS connection is required for beacon authentication",
			})
			c.Abort()
			return
		}

		// Extract peer certificates
		state := c.Request.TLS
		if len(state.PeerCertificates) == 0 {
			if mTLSConfig.Mode == "warn" {
				// Warn mode: log warning but allow request
				fmt.Printf("[MTLS WARN] Request from %s without client certificate (mTLS warn mode)\n", c.ClientIP())
				c.Next()
				return
			}
			// Strict mode: reject request
			c.JSON(http.StatusForbidden, gin.H{
				"code":    "CERTIFICATE_REQUIRED",
				"message": "Client certificate is required",
			})
			c.Abort()
			return
		}

		// Validate certificates
		if err := validatePeerCertificates(state.PeerCertificates); err != nil {
			if mTLSConfig.Mode == "warn" {
				// Warn mode: log warning but allow request
				fmt.Printf("[MTLS WARN] Request from %s with invalid certificate: %s (mTLS warn mode)\n", c.ClientIP(), err.Error())
				c.Next()
				return
			}
			// Strict mode: reject request
			c.JSON(http.StatusForbidden, gin.H{
				"code":    "CERTIFICATE_INVALID",
				"message": formatValidationError(err),
			})
			c.Abort()
			return
		}

		// Extract CN from certificate and store in context
		cert := state.PeerCertificates[0]
		if cert.Subject.CommonName != "" {
			c.Set("beacon_cn", cert.Subject.CommonName)
		}

		c.Next()
	}
}

// validatePeerCertificates validates peer certificates against mTLS configuration
func validatePeerCertificates(certificates []*x509.Certificate) error {
	if len(certificates) == 0 {
		return ErrTLSCertificateRequired
	}

	cert := certificates[0]

	// Check certificate expiration
	if !cert.NotBefore.IsZero() && cert.NotAfter.Before(getCurrentTime()) {
		return errors.New("certificate has expired")
	}

	// Check minimum expiry days
	minExpiry := time.Duration(mTLSConfig.MinCertExpiryDays) * 24 * time.Hour
	certExpiry := cert.NotAfter.Sub(getCurrentTime())
	if certExpiry < minExpiry {
		return errors.New("certificate expires too soon")
	}

	// Check CN prefixes if configured
	if len(mTLSConfig.AllowedCNPrefixes) > 0 {
		cnMatched := false
		cn := cert.Subject.CommonName
		for _, prefix := range mTLSConfig.AllowedCNPrefixes {
			if strings.HasPrefix(cn, prefix) {
				cnMatched = true
				break
			}
		}
		if !cnMatched {
			return errors.New("certificate Common Name not in allowed list")
		}
	}

	// Check OUs if configured
	if len(mTLSConfig.AllowedOUs) > 0 {
		ouMatched := false
		for _, ou := range cert.Subject.OrganizationalUnit {
			for _, allowedOU := range mTLSConfig.AllowedOUs {
				if ou == allowedOU {
					ouMatched = true
					break
				}
			}
		}
		if len(cert.Subject.OrganizationalUnit) > 0 && !ouMatched {
			return errors.New("certificate Organizational Unit not in allowed list")
		}
	}

	// Verify CA if specified
	if mTLSConfig.CASpecifiedCertificates && len(mTLSConfig.TrustedCAPEMs) > 0 {
		// Create cert pool with trusted CAs
		caPool := x509.NewCertPool()
		for _, caCertPEM := range mTLSConfig.TrustedCAPEMs {
			if !caPool.AppendCertsFromPEM([]byte(caCertPEM)) {
				return errors.New("failed to parse CA certificate")
			}
		}

		// Create intermediate options
		opts := x509.VerifyOptions{
			DNSName:       cert.Subject.CommonName,
			Intermediates: caPool,
		}

		// Verify certificate
		if _, err := cert.Verify(opts); err != nil {
			return errors.New("certificate verification failed: " + err.Error())
		}
	}

	// Check key usage and extended key usage
	if len(cert.ExtKeyUsage) == 0 {
		return errors.New("certificate has no extended key usage")
	}

	// Check for client auth extended key usage (required for beacons)
	hasClientAuth := false
	for _, usage := range cert.ExtKeyUsage {
		if usage == x509.ExtKeyUsageClientAuth {
			hasClientAuth = true
			break
		}
	}

	if !hasClientAuth {
		return errors.New("certificate must have clientAuth extended key usage")
	}

	return nil
}

// formatValidationError creates a user-friendly error message from validation error
func formatValidationError(err error) string {
	if errors.Is(err, ErrTLSCertificateRequired) {
		return "Client certificate is required"
	}
	if errors.Is(err, ErrCertificateInvalid) {
		return "Invalid certificate: " + err.Error()
	}
	return "Certificate validation failed: " + err.Error()
}

// getCurrentTime returns the current time (for mocking in tests)
var getCurrentTime = func() time.Time {
	return time.Now()
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvBool gets boolean environment variable with default
func getEnvBool(key string, defaultValue bool) bool {
	val := getEnvOrDefault(key, "")
	if val == "" {
		return defaultValue
	}
	return val == "true" || val == "1" || val == "enabled"
}

// getEnvInt gets integer environment variable with default
func getEnvInt(key string, defaultValue int) int {
	val := getEnvOrDefault(key, "")
	if val == "" {
		return defaultValue
	}
	intVal, err := parseIntSafe(val)
	if err != nil {
		return defaultValue
	}
	return intVal
}

// parseIntSafe parses string to int safely
func parseIntSafe(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// GetMTLSConfig returns the current mTLS configuration (for testing)
func GetMTLSConfig() *TLSConfig {
	if mTLSConfig == nil {
		return &TLSConfig{Mode: "disabled"}
	}
	return mTLSConfig
}

// SetMTLSConfig sets mTLS configuration (for testing)
func SetMTLSConfig(config *TLSConfig) {
	mTLSConfig = config
}
