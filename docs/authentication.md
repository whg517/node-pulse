# Node-Pulse Authentication Architecture

This document describes the authentication and authorization system for Node-Pulse, covering user authentication, beacon authentication, session management, and security controls.

## Table of Contents

1. [Overview](#overview)
2. [Transport Security](#transport-security)
3. [Authentication Flow](#authentication-flow)
4. [JWT Token Service](#jwt-token-service)
5. [Session Management](#session-management)
6. [Role-Based Access Control](#role-based-access-control)
7. [Beacon Authentication](#beacon-authentication)
8. [Security Controls](#security-controls)
9. [Password Requirements](#password-requirements)
10. [Database Schema](#database-schema)
11. [Configuration](#configuration)
12. [Security Considerations](#security-considerations)

---

## Overview

Node-Pulse implements a multi-layer authentication system supporting both human users (web UI) and machine agents (beacons). The architecture prioritizes security while maintaining usability.

### Key Components

| Component | Purpose | Technology |
|-----------|---------|------------|
| JWT Service | Access token generation/validation | RS256 signing with RSA-2048 |
| Session Service | User session management | PostgreSQL-backed with refresh tokens |
| RBAC Service | Role-based access control | Permission matrix |
| API Key Service | Beacon authentication | SHA-256 hashed keys |
| mTLS Middleware | Certificate-based auth | X.509 client certificates |

### Supported Authentication Methods

- **User Authentication**: Username/password with JWT access tokens
- **Beacon Authentication**: API key exchange for JWT + optional mTLS
- **Session Management**: Refresh token rotation with grace period

---

## Transport Security

### TLS Requirements

**All API communications MUST use TLS 1.2 or higher.** This is a baseline security requirement for all Node-Pulse deployments.

| Protocol | Status | Notes |
|----------|--------|-------|
| TLS 1.3 | Recommended | Best performance and security |
| TLS 1.2 | Minimum required | Supported for broader compatibility |
| TLS 1.1 and below | Prohibited | Vulnerable to attacks |

### TLS Configuration Best Practices

| Setting | Recommended Value |
|---------|-------------------|
| Cipher Suites | ECDHE + AESGCM (forward secrecy) |
| Certificate | Valid CA-signed, not self-signed for production |
| HSTS | Enabled with `max-age=31536000` |
| Certificate Renewal | Automated (e.g., cert-manager, Let's Encrypt) |

### Endpoint-Specific Security

| Endpoint Type | TLS Requirement | Additional Protection |
|---------------|-----------------|----------------------|
| User Web UI | TLS 1.2+ | httpOnly + Secure cookies |
| Beacon API | TLS 1.2+ | mTLS recommended (see Beacon Authentication) |
| Metrics (`/metrics`) | TLS 1.2+ | Basic Auth or network ACL for standalone mode |

---

## Authentication Flow

### User Login Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│  API Server │────▶│  PostgreSQL │────▶│   Client    │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │                   │
       │  POST /auth/login │                   │                   │
       │  {username, pwd}  │                   │                   │
       │──────────────────▶│                   │                   │
       │                   │  Validate creds   │                   │
       │                   │──────────────────▶│                   │
       │                   │                   │                   │
       │                   │  Create session   │                   │
       │                   │  Generate tokens  │                   │
       │                   │◀──────────────────│                   │
       │                   │                   │                   │
       │  Access Token (cookie)               │                   │
       │  Refresh Token (httpOnly cookie)     │                   │
       │◀──────────────────│                   │                   │
```

### Token Refresh Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│  API Server │────▶│  PostgreSQL │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │
       │  POST /auth/refresh                  │
       │  (refresh_token cookie)              │
       │──────────────────▶│                   │
       │                   │  Validate token   │
       │                   │  Check blacklist  │
       │                   │──────────────────▶│
       │                   │                   │
       │                   │  Rotate token     │
       │                   │  (old revoked,    │
       │                   │   new created)    │
       │                   │◀──────────────────│
       │                   │                   │
       │  New Access Token │                   │
       │  New Refresh Token│                   │
       │◀──────────────────│                   │
```

### Beacon Authentication Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Beacon    │────▶│  API Server │────▶│  PostgreSQL │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │
       │  POST /beacon/token                  │
       │  {api_key}        │                   │
       │──────────────────▶│                   │
       │                   │  Validate API key │
       │                   │  (SHA-256 hash)   │
       │                   │──────────────────▶│
       │                   │                   │
       │                   │  Generate JWT     │
       │                   │  (beacon role)    │
       │                   │◀──────────────────│
       │                   │                   │
       │  JWT Access Token │                   │
       │◀──────────────────│                   │
       │                   │                   │
       │  POST /beacon/heartbeat (with JWT)    │
       │──────────────────▶│                   │
       │                   │  Validate JWT     │
       │                   │  Process heartbeat│
       │                   │──────────────────▶│
```

### Degraded Mode Authentication

When the Pulse Server is unreachable (3 consecutive heartbeat failures), Beacons enter **degraded mode**:

**Authentication Behavior in Degraded Mode:**
1. Beacon continues using locally cached configuration
2. Beacon attempts JWT renewal with exponential backoff (1s, 2s, 4s intervals)
3. If JWT expires during degraded mode, Beacon re-authenticates using API key when server becomes available
4. No heartbeat data is lost - data is cached locally (max 10MB) and transmitted upon reconnection

**JWT Expiration Handling:**
- If JWT expires while server is unreachable, Beacon queues data locally
- Upon server recovery, Beacon immediately re-authenticates with API key
- Cached data is transmitted in order after successful re-authentication

---

## JWT Token Service

### Overview

The JWT service handles access token generation and validation using RS256 (RSA Signature with SHA-256) asymmetric signing.

### Key Features

| Feature | Description |
|---------|-------------|
| Signing Algorithm | RS256 (RSA-2048) |
| Key Rotation | Supported via `kid` header |
| Token Expiration | 15 minutes (users), 1 hour (beacons) |
| Clock Skew Tolerance | 60 seconds |
| Blacklist Support | Database-backed revocation |

### Token Structure

**Header:**
```json
{
  "alg": "RS256",
  "typ": "JWT",
  "kid": "key-id-here"
}
```

**Payload (Claims):**
| Claim | Type | Description |
|-------|------|-------------|
| `user_id` | string | Unique user identifier (UUID) |
| `role` | string | User role (admin/operator/viewer/beacon) |
| `session_id` | string | Session identifier (optional) |
| `scope` | array | Permission scopes (optional) |
| `jti` | string | JWT ID for blacklist tracking |
| `iss` | string | Issuer: "node-pulse" |
| `sub` | string | Subject: user_id |
| `aud` | string | Audience: "node-pulse-api" |
| `exp` | number | Expiration timestamp |
| `iat` | number | Issued at timestamp |
| `nbf` | number | Not valid before timestamp |

### Key Management

- **Auto-generation**: RSA-2048 keys are auto-generated if not provided
- **Environment configuration**: Keys can be provided via environment variables
- **Key rotation**: Supported through `kid` header; tokens with mismatched `kid` are rejected

### Validation Process

1. Parse token and verify RS256 signature using public key
2. Validate `kid` header matches current key ID (if configured)
3. Verify issuer claim matches "node-pulse"
4. Verify audience claim matches "node-pulse-api"
5. Check expiration with 60-second clock skew tolerance
6. Query token blacklist database for `jti`
7. Extract and return claims

---

## Session Management

### Overview

Sessions track user authentication state across multiple devices and enable refresh token rotation for security.

### Session Limits

| Parameter | Value | Description |
|-----------|-------|-------------|
| Max Sessions per User | 10 | Oldest session evicted when exceeded |
| Session Expiry | 7 days | Default session validity |
| Max Validity | 30 days | Absolute maximum (90 days with "remember me") |
| Grace Period | 5 minutes | Tolerance for concurrent refresh requests |

### Refresh Token Rotation

Refresh tokens are **single-use** - each refresh operation:
1. Validates the current refresh token
2. Revokes the current token
3. Issues a new refresh token
4. Links the new token to the previous one (`replaced_by` field)

### Grace Period Handling

To handle race conditions (e.g., multiple browser tabs refreshing simultaneously):

1. When a revoked token is detected, check revocation timestamp
2. If within 5-minute grace period AND same IP address → Allow refresh
3. If outside grace period OR different IP → Reject and potentially revoke token family

### Token Family Tracking

Token families enable detection of token reuse attacks:
- Each new token stores reference to the token it replaced (`replaced_by`)
- If a reused token is detected outside grace period, all tokens in the family are revoked
- This prevents attackers from using stolen refresh tokens

### Session Mutex Management

To prevent race conditions during session operations:
- Per-user mutex locks ensure atomic session creation/refresh
- Background cleanup removes unused mutexes every 5 minutes
- Prevents memory leaks from accumulated mutexes

---

## Role-Based Access Control

### Role Hierarchy

```
admin (level 3)
  └── operator (level 2)
        └── beacon (level 1)
        └── viewer (level 0)
```

### Permission Matrix

| Resource | Admin | Operator | Viewer | Beacon |
|----------|-------|----------|--------|--------|
| **Users** |
| View | ✅ | ❌ | ❌ | ❌ |
| Create | ✅ | ❌ | ❌ | ❌ |
| Update | ✅ | ❌ | ❌ | ❌ |
| Delete | ✅ | ❌ | ❌ | ❌ |
| **Nodes** |
| View | ✅ | ✅ | ✅ | ❌ |
| Create | ✅ | ✅ | ❌ | ❌ |
| Update | ✅ | ✅ | ❌ | ❌ |
| Delete | ✅ | ✅ | ❌ | ❌ |
| **Probes** |
| View | ✅ | ✅ | ✅ | ❌ |
| Create | ✅ | ✅ | ❌ | ❌ |
| Update | ✅ | ✅ | ❌ | ❌ |
| Delete | ✅ | ✅ | ❌ | ❌ |
| **Alerts** |
| View | ✅ | ✅ | ✅ | ❌ |
| Create | ✅ | ✅ | ❌ | ❌ |
| Update | ✅ | ✅ | ❌ | ❌ |
| Delete | ✅ | ✅ | ❌ | ❌ |
| **Webhooks** |
| View | ✅ | ❌ | ❌ | ❌ |
| Create | ✅ | ❌ | ❌ | ❌ |
| Update | ✅ | ❌ | ❌ | ❌ |
| Delete | ✅ | ❌ | ❌ | ❌ |
| **Export** |
| View | ✅ | ❌ | ❌ | ❌ |
| Create | ✅ | ❌ | ❌ | ❌ |
| **System** |
| View | ✅ | ✅ | ✅ | ❌ |
| Admin | ✅ | ❌ | ❌ | ❌ |
| **Beacon** |
| Read | ✅ | ✅ | ❌ | ❌ |
| Write | ✅ | ✅ | ❌ | ✅ |
| **Config** |
| View | ✅ | ❌ | ❌ | ✅ |
| Update | ✅ | ❌ | ❌ | ❌ |

### Resource-Level Access Control

In addition to role-based permissions, operators can only modify resources they created:

- **Nodes**: Operator can only update/delete nodes they created
- **Probes**: Operator can only update/delete probes they created
- **Alerts**: Operator can only update/delete alert rules they created

Admins have unrestricted access to all resources.

---

## Beacon Authentication

### API Key Format

| Attribute | Value |
|-----------|-------|
| Type | 256-bit random token |
| Encoding | Base64 URL-safe |
| Storage | SHA-256 hash in database |
| Prefix | First 8 characters stored for identification |
| Example | `np_live_abc123...` |

### API Key Lifecycle

1. **Generation**: Admin creates key via API or Dashboard
2. **Distribution**: Key securely delivered to beacon (out of band)
3. **Exchange**: Beacon exchanges key for JWT at `/beacon/token` (via Authorization header)
4. **Usage**: Beacon includes JWT in Authorization header for subsequent requests
5. **Rotation**: Recommended every 90 days; warning at 7 days before expiry
6. **Revocation**: Admin can revoke key at any time

### Beacon Token Exchange

**Request Format:**
```http
POST /api/v1/beacon/token HTTP/1.1
Host: pulse.example.com
Authorization: Bearer np_live_abc123def456ghi789...
Content-Type: application/json
Content-Length: 2

{}
```

**Response Format:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

**Security Notes:**
- API key MUST be sent in `Authorization` header, NOT in request body
- Response contains only JWT access token (no refresh token for beacons)
- Beacon re-authenticates with API key when JWT expires (1 hour default)

### API Key Expiration and Rotation

| Parameter | Value | Description |
|-----------|-------|-------------|
| Default Expiry | 1 year | Configurable per key |
| Expiry Warning | 7 days before | Dashboard alert + API response header |
| Rotation Window | 24 hours | Old key remains valid during rotation |
| Max Keys per Beacon | 2 | Allows rotation without downtime |

**Rotation Process:**
1. Admin generates new API key for beacon
2. Both old and new keys are valid for 24 hours
3. Beacon configuration updated with new key
4. Old key automatically revoked after 24 hours or on first successful use of new key

### mTLS (Mutual TLS)

For production environments, mTLS provides additional security:

**Configuration:**
| Setting | Description |
|---------|-------------|
| `PULSE_MTLS_ENABLED` | Enable mTLS enforcement |
| `PULSE_MTLS_CA_CERTS` | Trusted CA certificates (PEM) |
| `PULSE_MTLS_CN_PREFIXES` | Allowed Common Name prefixes |
| `PULSE_MTLS_ALLOWED_OUS` | Allowed Organizational Units |
| `PULSE_MTLS_MIN_CERT_EXPIRY_DAYS` | Minimum certificate validity |

**Certificate Validation:**
1. Verify TLS connection exists
2. Extract peer certificates
3. Validate certificate not expired
4. Check minimum expiry days (default: 7)
5. Verify CN against allowed prefixes (if configured)
6. Verify OU against allowed list (if configured)
7. Verify CA chain (if configured)
8. Check for `clientAuth` extended key usage

**Production Recommendation:** mTLS is **strongly recommended** for production deployments. The system provides flexible enforcement modes:

| Mode | Setting | Behavior |
|------|---------|----------|
| Disabled | `PULSE_MTLS_ENABLED=false` | mTLS not required |
| Warning | `PULSE_MTLS_ENABLED=warn` | Log warnings for missing mTLS, allow requests |
| Strict | `PULSE_MTLS_ENABLED=strict` | Reject requests without valid mTLS |

**Migration Path:**
1. Start with `disabled` during initial deployment
2. Switch to `warn` to identify beacons needing certificate updates
3. After all beacons configured, enable `strict` mode

**Note:** For high-security environments, enable `strict` mode from the start and provision certificates before beacon deployment.

---

## Security Controls

### Rate Limiting

| Endpoint Type | Per-Minute Limit |
|---------------|------------------|
| Login | 5 attempts |
| Refresh | 10 attempts |
| Logout | 10 attempts |
| API Key Exchange | 11 attempts |

### Account Lockout

| Failed Attempts | Lockout Duration |
|-----------------|------------------|
| 5 | 10 minutes |

After 5 consecutive failed login attempts, the account is locked for 10 minutes.

### Progressive Rate Limiting

Penalty escalation for repeated violations:

| Violation Level | Lockout Duration | Condition |
|-----------------|------------------|-----------|
| Level 1 | 60 seconds | First violation |
| Level 2 | 5 minutes | Second violation within 24h |
| Level 3 | 1 hour | Third violation within 24h |
| Level 4 | 24 hours | Fourth+ violation within 24h |

### SSRF Protection

Webhook URLs are validated to prevent Server-Side Request Forgery:

**Blocked IP Ranges:**
| CIDR | Description |
|------|-------------|
| 10.0.0.0/8 | RFC 1918 Private |
| 172.16.0.0/12 | RFC 1918 Private |
| 192.168.0.0/16 | RFC 1918 Private |
| 127.0.0.0/8 | Loopback |
| 169.254.169.254/32 | Cloud metadata (AWS/GCP/Azure) |
| ::1/128 | IPv6 loopback |
| fc00::/7 | IPv6 ULA |
| fe80::/10 | IPv6 link-local |
| 0.0.0.0/8 | Invalid addresses |
| 224.0.0.0/4 | IP multicast |
| 255.255.255.255/32 | Broadcast |

**Validation Process:**
1. Parse URL and verify HTTPS scheme (only HTTPS allowed)
2. Check domain allowlist (if configured)
3. Resolve DNS for hostname
4. Verify ALL resolved IPs are not in blocked ranges

### Timing Attack Prevention

| Protection | Implementation |
|------------|----------------|
| Constant Auth Delay | 150ms fixed delay on authentication failure |
| Password Comparison | bcrypt constant-time comparison |
| Content-Type Validation | Auth endpoints require `application/json` |

### Audit Logging

All authentication events are logged to `auth_audit_logs`:
- Login attempts (success/failure)
- Token refresh operations
- Session management (create/revoke)
- API key operations
- RBAC permission checks

### Audit Log Retention

| Log Type | Retention Period | Cleanup Method |
|----------|------------------|----------------|
| Authentication Events | 90 days | Automated daily cleanup |
| Session Events | 90 days | Automated daily cleanup |
| API Key Operations | 90 days | Automated daily cleanup |
| RBAC Permission Checks | 30 days | Automated daily cleanup |

**Retention Policy Implementation:**
- Background job runs daily at 02:00 UTC
- Deletes records older than retention period
- Audit logs are immutable (no updates after creation)

---

## Password Requirements

### Validation Rules

| Rule | Requirement |
|------|-------------|
| Minimum Length | 8 characters |
| Maximum Length | 32 characters |
| Uppercase | At least one (A-Z) |
| Lowercase | At least one (a-z) |
| Digit | At least one (0-9) |

### Hashing

| Parameter | Value |
|-----------|-------|
| Algorithm | bcrypt |
| Cost Factor | 12 |
| Output Length | 60 characters |

---

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    user_id UUID PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    email VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'viewer',
    is_active BOOLEAN DEFAULT true,
    locked_until TIMESTAMPTZ,
    failed_login_attempts INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Sessions Table

```sql
CREATE TABLE sessions (
    session_id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    device_id VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    remember_me BOOLEAN DEFAULT false,
    expires_at TIMESTAMPTZ NOT NULL,
    max_valid_until TIMESTAMPTZ NOT NULL,
    last_activity_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Refresh Tokens Table

```sql
CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    token_id UUID UNIQUE NOT NULL,
    token_hash TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    session_id UUID REFERENCES sessions(session_id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    max_valid_until TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    replaced_by UUID REFERENCES refresh_tokens(token_id),
    user_agent TEXT,
    ip_address INET,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### API Keys Table

```sql
CREATE TABLE api_keys (
    id SERIAL PRIMARY KEY,
    key_hash TEXT UNIQUE NOT NULL,
    key_prefix TEXT NOT NULL,
    user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_used_at TIMESTAMPTZ
);
```

### Token Blacklist Table

```sql
CREATE TABLE token_blacklist (
    jti TEXT PRIMARY KEY,
    revoked_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);
```

### Rate Limits Table

```sql
CREATE TABLE rate_limits (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL,
    window_type VARCHAR(10) NOT NULL,
    window_start TIMESTAMPTZ NOT NULL,
    request_count INTEGER DEFAULT 1,
    UNIQUE(key, window_type, window_start)
);
```

### Password Reset Tokens Table

```sql
CREATE TABLE password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,  -- SHA-256 hash of reset token
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,  -- NULL until used
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
```

**Token Properties:**
| Property | Value |
|-----------|-------|
| Token Format | 256-bit random, URL-safe base64 |
| Hash Algorithm | SHA-256 |
| Expiry | 1 hour |
| Single Use | Yes (marked used after successful reset) |
| Cleanup | Automated deletion of expired tokens daily |

### Audit Log Table

```sql
CREATE TABLE auth_audit_logs (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    user_id UUID,
    ip_address INET,
    details JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for retention cleanup queries
CREATE INDEX idx_auth_audit_logs_created_at ON auth_audit_logs(created_at);

-- Index for user-specific queries
CREATE INDEX idx_auth_audit_logs_user_id ON auth_audit_logs(user_id);
```

**Retention:** 90 days for authentication events, 30 days for RBAC checks. Automated cleanup runs daily.

### Password Reset Tokens Table

```sql
CREATE TABLE password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    ip_address INET
);

-- Index for token lookup
CREATE INDEX idx_password_reset_tokens_hash ON password_reset_tokens(token_hash);

-- Index for cleanup
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
```

**Token Properties:**
| Property | Value |
|----------|-------|
| Token Length | 256-bit random |
| Storage | SHA-256 hash only |
| Expiry | 1 hour |
| Single Use | Yes (marked used after use) |
| Cleanup | Expired tokens deleted daily |

---

## Configuration

### Environment Variables

#### JWT Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PULSE_JWT_PRIVATE_KEY` | RSA private key (PEM format) | Auto-generated |
| `PULSE_JWT_PUBLIC_KEY` | RSA public key (PEM format) | Auto-generated |
| `PULSE_JWT_KEY_ID` | Key identifier for rotation | "default-key" |

#### Session Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PULSE_SESSION_COOKIE_SECURE` | Secure cookie flag | false |

#### mTLS Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PULSE_MTLS_ENABLED` | Enable mTLS | false |
| `PULSE_MTLS_CA_CERTS` | CA certificates (inline PEM) | - |
| `PULSE_MTLS_CA_CERT_FILE` | Path to CA cert file | - |
| `PULSE_MTLS_CN_PREFIXES` | Allowed CN prefixes (comma-separated) | - |
| `PULSE_MTLS_ALLOWED_OUS` | Allowed OUs (comma-separated) | - |
| `PULSE_MTLS_MIN_CERT_EXPIRY_DAYS` | Minimum cert validity | 7 |

#### Server Mode

| Variable | Description | Default |
|----------|-------------|---------|
| `PULSE_SERVER_MODE` | Server mode (debug/production) | debug |

**Note:** When `PULSE_SERVER_MODE=production`, mTLS warning mode is recommended. Enable strict mode after certificate provisioning is complete.

### Token Expiration Settings

| Setting | User (Web UI) | Beacon (M2M) |
|---------|---------------|--------------|
| Access Token | 15 minutes | 1 hour |
| Refresh Token | 7 days | N/A (API key exchange) |
| Maximum Validity | 30 days (90 with "remember me") | N/A |
| Grace Period | 5 minutes | N/A |

**Rationale for Beacon Token Expiration:**
- Beacons use API keys as primary authentication mechanism
- Longer JWT expiry (1 hour) reduces re-authentication frequency
- API key provides security boundary; JWT is short-lived session token

---

## Security Considerations

### Threat Model

| Threat | Mitigation |
|--------|------------|
| Credential theft | bcrypt hashing, rate limiting, account lockout |
| Token theft | Short-lived access tokens, httpOnly cookies |
| Refresh token replay | Single-use tokens, family tracking, IP validation |
| Session hijacking | IP tracking, device fingerprinting (optional) |
| SSRF attacks | IP range blocking, DNS resolution validation |
| Brute force | Rate limiting, progressive lockout |
| Privilege escalation | RBAC, resource-level access control |
| Beacon impersonation | mTLS (recommended), API key validation, Authorization header |
| Timing attacks | 150ms constant delay on auth failure |
| CSRF attacks | SameSite cookie attribute, CSRF token for sensitive operations |
| Email enumeration | Generic password reset responses |
| API key exposure | Authorization header (not body), HTTPS only |

### CSRF Protection

The system implements multiple layers of CSRF protection:

| Protection | Implementation |
|------------|----------------|
| SameSite Cookies | `SameSite=Strict` for session cookies |
| CSRF Token | Required for state-changing operations (POST/PUT/DELETE) |
| Origin Validation | Server validates `Origin` header against allowed origins |
| Referer Check | Fallback validation when Origin header is missing |

**CSRF Token Lifecycle:**
1. Token generated on login, stored in httpOnly cookie
2. Token included in response body for frontend storage
3. Frontend includes token in `X-CSRF-Token` header for mutations
4. Server validates token before processing state-changing requests

### Best Practices

1. **Production Deployment**
   - Enable mTLS for beacon authentication (use `warn` mode first)
   - Set `PULSE_SESSION_COOKIE_SECURE=true`
   - Configure proper CA certificates
   - Use strong, unique JWT keys from secrets manager
   - Ensure TLS 1.2+ for all communications

2. **Key Management**
   - Store JWT private keys in HSM or secrets manager
   - Rotate keys periodically using `kid` header
   - Never commit keys to version control
   - Rotate API keys every 90 days

3. **Monitoring**
   - Monitor authentication audit logs
   - Alert on repeated failed login attempts
   - Track unusual refresh patterns
   - Monitor for token reuse attempts
   - Alert when API keys approach expiration (7 days)

4. **Incident Response**
   - Revoke compromised API keys immediately
   - Use session revocation for suspected token theft
   - Review audit logs for unauthorized access
   - Consider revoking all user sessions if breach suspected

### Device Fingerprinting (Optional Enhancement)

For enhanced session hijacking protection, device fingerprinting can be implemented:

**Fingerprint Components:**
| Component | Source | Stability |
|-----------|--------|-----------|
| User Agent | `User-Agent` header | Medium |
| Screen Resolution | Client-side JS | Medium |
| Timezone | Client-side JS | High |
| Language | `Accept-Language` header | High |

**Implementation Notes:**
- Fingerprint stored in session record
- Significant fingerprint changes trigger re-authentication prompt
- Can be enabled via `PULSE_DEVICE_FINGERPRINT_ENABLED=true`

### Known Limitations

1. **Database Dependency**: Token blacklist requires database availability. If database is unavailable, token revocation cannot be verified.

2. **Single-Region Sessions**: Current implementation uses single database. For multi-region deployments, consider distributed cache for blacklist.

3. **Token Refresh Race**: While grace period handles most cases, extreme concurrent refresh attempts (>5 within grace period) may cause issues.

4. **API Key Rate Limit**: The per-minute limit of 11 for API key exchange is intentionally set to an unusual value to prevent common attack patterns.

5. **Password Reset Token Security**: Reset tokens are single-use, expire after 1 hour, and stored as SHA-256 hashes. Email responses are generic to prevent user enumeration.

6. **Session Ownership**: Users can only revoke their own sessions. The DELETE /auth/sessions/:id endpoint enforces ownership by comparing session.user_id with JWT user_id.

---

## API Endpoints

### Authentication Endpoints

| Endpoint | Method | Auth Required | Description |
|----------|--------|---------------|-------------|
| `/api/v1/auth/login` | POST | No | User login |
| `/api/v1/auth/refresh` | POST | No | Refresh tokens (refresh token in cookie) |
| `/api/v1/auth/logout` | POST | Yes | Logout user |
| `/api/v1/auth/verify` | GET | Yes | Validate current token, return claims |
| `/api/v1/auth/me` | GET | Yes | Get current user |
| `/api/v1/auth/sessions` | GET | Yes | List user sessions |
| `/api/v1/auth/sessions/:id` | DELETE | Yes | Revoke specific session |
| `/api/v1/auth/sessions/revoke-all` | POST | Yes | Revoke all user's own sessions |
| `/api/v1/auth/session-info` | GET | Yes | Session expiration info |

**Session Ownership Rules:**
- Users can only delete their own sessions (enforced by user_id from JWT context)
- Admins can delete any user's sessions via admin endpoints

### Password Management Endpoints

| Endpoint | Method | Auth Required | Description |
|----------|--------|---------------|-------------|
| `/api/v1/auth/password/change` | POST | Yes | Change current user's password |
| `/api/v1/auth/password/reset/request` | POST | No | Request password reset email |
| `/api/v1/auth/password/reset/confirm` | POST | No | Confirm password reset with token |

**Password Change Flow:**
```
POST /api/v1/auth/password/change
Authorization: Bearer <jwt>
Content-Type: application/json

{
  "current_password": "OldPass123",
  "new_password": "NewPass456"
}

Response:
{
  "message": "Password changed successfully",
  "sessions_revoked": true  // All other sessions revoked for security
}
```

**Password Reset Flow:**
```
Step 1: Request reset
POST /api/v1/auth/password/reset/request
{
  "email": "user@example.com"
}
Response: { "message": "If email exists, reset link sent" }  // Generic response prevents enumeration

Step 2: Confirm reset (via email link)
POST /api/v1/auth/password/reset/confirm
{
  "token": "reset-token-from-email",
  "new_password": "NewPass456"
}
Response: { "message": "Password reset successfully" }
```

**Reset Token Properties:**
| Property | Value |
|----------|-------|
| Token Length | 256-bit random |
| Expiry | 1 hour |
| Single Use | Yes (invalidated after use) |
| Storage | SHA-256 hash in database |

### Beacon Endpoints

| Endpoint | Method | Auth Required | Description |
|----------|--------|---------------|-------------|
| `/api/v1/beacon/token` | POST | API Key (header) | Exchange API key for JWT |
| `/api/v1/beacon/heartbeat` | POST | JWT (mTLS recommended) | Submit heartbeat data |
| `/api/v1/beacon/heartbeat/compressed` | POST | JWT (mTLS recommended) | Submit compressed heartbeat |

**API Key Authentication (Beacon Token Exchange):**

API keys MUST be sent via the `Authorization` header, not in the request body:

```http
POST /api/v1/beacon/token
Authorization: Bearer np_live_abc123def456...
Content-Type: application/json

{}

Response:
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

**Why Authorization Header:**
- Prevents API key logging in request bodies
- Follows RFC 6750 Bearer Token specification
- Consistent with standard HTTP authentication patterns
- Enables better security monitoring and filtering

**Note:** Response contains only JWT access token. Beacons do NOT receive refresh tokens - they re-authenticate with API key when JWT expires.

### API Key Management Endpoints (Admin)

| Endpoint | Method | Role Required | Description |
|----------|--------|---------------|-------------|
| `/api/v1/admin/api-keys` | GET | admin | List all API keys |
| `/api/v1/admin/api-keys/:id` | GET | admin | Get API key details |
| `/api/v1/admin/api-keys` | POST | admin | Create new API key |
| `/api/v1/admin/api-keys/:id` | DELETE | admin | Revoke API key |
| `/api/v1/admin/api-keys/:id/rotate` | POST | admin | Rotate API key |

**Create API Key:**
```
POST /api/v1/admin/api-keys
Authorization: Bearer <admin-jwt>
{
  "name": "beacon-singapore-01",
  "beacon_id": "uuid-of-beacon",  // Optional: link to beacon
  "expires_in_days": 365  // Optional: default 365
}

Response:
{
  "key_id": "uuid-key-id",
  "key": "np_live_abc123...",  // ONLY returned once - store securely!
  "name": "beacon-singapore-01",
  "expires_at": "2027-03-09T00:00:00Z",
  "created_at": "2026-03-09T00:00:00Z"
}
```

**Rotate API Key (24-hour overlap):**
```
POST /api/v1/admin/api-keys/:id/rotate
Authorization: Bearer <admin-jwt>

Response:
{
  "old_key_id": "uuid-old",
  "new_key_id": "uuid-new",
  "new_key": "np_live_xyz789...",  // New key - store securely!
  "overlap_expires_at": "2026-03-10T00:00:00Z",  // Old key valid until this time
  "message": "Both keys valid for 24 hours. Update beacon configuration."
}
```

### Admin Endpoints

| Endpoint | Method | Role Required | Description |
|----------|--------|---------------|-------------|
| `/api/v1/admin/auth/revoke-all/:userId` | POST | admin | Revoke all user sessions |
| `/api/v1/admin/users` | GET | admin | List all users |
| `/api/v1/admin/users/:id` | GET | admin | Get user details |
| `/api/v1/admin/users` | POST | admin | Create user |
| `/api/v1/admin/users/:id` | PUT | admin | Update user |
| `/api/v1/admin/users/:id` | DELETE | admin | Delete user |

### Audit Log Endpoints (Admin)

| Endpoint | Method | Role Required | Description |
|----------|--------|---------------|-------------|
| `/api/v1/admin/audit/logs` | GET | admin | Query audit logs |
| `/api/v1/admin/audit/logs/:id` | GET | admin | Get specific audit log |

**Audit Log Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `user_id` | UUID | Filter by user |
| `event_type` | string | Filter by event type |
| `from` | timestamp | Start time |
| `to` | timestamp | End time |
| `page` | int | Pagination |
| `limit` | int | Results per page (max 100) |

---

## Endpoint Security Summary

### Public Endpoints (No Auth)

| Endpoint | Rate Limit | Protections |
|----------|------------|-------------|
| `POST /auth/login` | 5/min | Account lockout, constant-time delay, bcrypt |
| `POST /auth/refresh` | 10/min | Token hash rate limiting, rotation, grace period |
| `POST /auth/password/reset/request` | 3/min | Email enumeration prevention |
| `POST /auth/password/reset/confirm` | 5/min | Single-use token, 1-hour expiry |
| `POST /beacon/token` | 11/min | SHA-256 key validation, IP logging |

### Protected Endpoints (JWT Required)

| Endpoint | Additional Protections |
|----------|----------------------|
| `GET /auth/verify` | Returns fresh token status |
| `POST /auth/logout` | Blacklists refresh token |
| `POST /auth/password/change` | Requires current password, revokes other sessions |
| `DELETE /auth/sessions/:id` | Ownership enforced (user_id match) |

### Admin Endpoints (RBAC)

| Endpoint | Audit Logged |
|----------|-------------|
| All `/admin/*` endpoints | Yes - all admin actions logged |

---

## References

- [JWT Best Practices (RFC 8725)](https://tools.ietf.org/html/rfc8725)
- [OAuth 2.0 Security Best Current Practice](https://tools.ietf.org/html/draft-ietf-oauth-security-topics)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [bcrypt Cost Factor](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#bcrypt)
