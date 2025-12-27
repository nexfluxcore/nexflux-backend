# Security Settings API Requirements

Dokumentasi lengkap API endpoints yang dibutuhkan untuk fitur Security Settings di NexFlux frontend.

---

## 📋 Table of Contents

1. [Password Management](#1-password-management)
2. [Two-Factor Authentication](#2-two-factor-authentication)
3. [Session Management](#3-session-management)
4. [Login History](#4-login-history)
5. [Database Schema](#5-database-schema)
6. [Security Best Practices](#6-security-best-practices)

---

## 1. Password Management

### PUT `/auth/password`
Change user's password.

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "current_password": "oldPassword123",
  "new_password": "newSecurePassword456"
}
```

**Validation Requirements:**
- `current_password`: Required, must match current password
- `new_password`: Required, minimum 8 characters

**Success Response (200):**
```json
{
  "success": true,
  "message": "Password changed successfully"
}
```

**Error Responses:**

| Status | Code | Message |
|--------|------|---------|
| 400 | INVALID_REQUEST | "New password must be at least 8 characters" |
| 401 | INVALID_PASSWORD | "Current password is incorrect" |
| 401 | UNAUTHORIZED | "Invalid or expired token" |
| 422 | SAME_PASSWORD | "New password cannot be the same as current password" |

**Backend Implementation Notes:**
```go
// Example Go/Fiber handler
func ChangePassword(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string)
    
    var req struct {
        CurrentPassword string `json:"current_password" validate:"required"`
        NewPassword     string `json:"new_password" validate:"required,min=8"`
    }
    
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    // Verify current password
    user, err := userRepo.GetByID(userID)
    if err != nil || !bcrypt.ComparePassword(user.PasswordHash, req.CurrentPassword) {
        return c.Status(401).JSON(fiber.Map{"error": "Current password is incorrect"})
    }
    
    // Hash and update new password
    newHash, _ := bcrypt.HashPassword(req.NewPassword)
    userRepo.UpdatePassword(userID, newHash)
    
    // Optionally: Invalidate all other sessions
    sessionRepo.RevokeAllExceptCurrent(userID, currentSessionID)
    
    return c.JSON(fiber.Map{"success": true, "message": "Password changed successfully"})
}
```

---

## 2. Two-Factor Authentication

### POST `/auth/2fa/enable`
Enable two-factor authentication.

**Request:** No body required

**Response (200):**
```json
{
  "success": true,
  "data": {
    "secret": "JBSWY3DPEHPK3PXP",
    "qr_code_url": "otpauth://totp/NexFlux:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=NexFlux",
    "backup_codes": [
      "ABC123DEF",
      "GHI456JKL",
      "MNO789PQR",
      "STU012VWX",
      "YZA345BCD"
    ]
  },
  "message": "Two-factor authentication enabled. Please save your backup codes."
}
```

### POST `/auth/2fa/verify`
Verify 2FA code during setup.

**Request Body:**
```json
{
  "code": "123456"
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Two-factor authentication verified and activated"
}
```

### POST `/auth/2fa/disable`
Disable two-factor authentication.

**Request Body:**
```json
{
  "code": "123456",
  "password": "userPassword123"
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Two-factor authentication disabled"
}
```

### GET `/auth/2fa/status`
Check 2FA status.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "enabled": true,
    "enabled_at": "2024-12-26T10:00:00Z",
    "backup_codes_remaining": 3
  }
}
```

---

## 3. Session Management

### GET `/auth/sessions`
Get all active sessions for the user.

**Response (200):**
```json
{
  "success": true,
  "data": [
    {
      "id": "session-uuid-1",
      "device": "MacBook Pro",
      "browser": "Chrome 120.0.6099",
      "ip_address": "192.168.1.100",
      "location": "Jakarta, Indonesia",
      "last_active": "2024-12-27T08:00:00Z",
      "is_current": true,
      "created_at": "2024-12-26T10:00:00Z"
    },
    {
      "id": "session-uuid-2",
      "device": "iPhone 15 Pro",
      "browser": "Safari Mobile 17.2",
      "ip_address": "103.253.145.22",
      "location": "Surabaya, Indonesia",
      "last_active": "2024-12-27T06:30:00Z",
      "is_current": false,
      "created_at": "2024-12-25T14:00:00Z"
    }
  ]
}
```

### DELETE `/auth/sessions/:sessionId`
Revoke a specific session.

**Response (200):**
```json
{
  "success": true,
  "message": "Session revoked successfully"
}
```

### DELETE `/auth/sessions`
Revoke all sessions except current.

**Response (200):**
```json
{
  "success": true,
  "message": "All other sessions have been revoked",
  "data": {
    "revoked_count": 3
  }
}
```

**Backend Implementation Notes:**
```go
// Parsing User-Agent for device info
func parseUserAgent(ua string) (device, browser string) {
    // Use ua-parser library
    parser := uaparser.New()
    client := parser.Parse(ua)
    
    device = client.Device.Family
    if client.Device.Model != "" {
        device = client.Device.Model
    }
    
    browser = fmt.Sprintf("%s %s", client.UserAgent.Family, client.UserAgent.Major)
    return
}

// Get location from IP
func getLocationFromIP(ip string) string {
    // Use MaxMind GeoIP2 or similar
    record, err := geoIP.City(net.ParseIP(ip))
    if err != nil {
        return "Unknown"
    }
    return fmt.Sprintf("%s, %s", record.City.Names["en"], record.Country.Names["en"])
}
```

---

## 4. Login History

### GET `/auth/login-history`
Get recent login attempts.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | number | 20 | Max results |
| `page` | number | 1 | Page number |

**Response (200):**
```json
{
  "success": true,
  "data": [
    {
      "id": "login-uuid-1",
      "device": "MacBook Pro",
      "browser": "Chrome 120.0.6099",
      "ip_address": "192.168.1.100",
      "location": "Jakarta, Indonesia",
      "status": "success",
      "created_at": "2024-12-27T08:00:00Z",
      "failure_reason": null
    },
    {
      "id": "login-uuid-2",
      "device": "Unknown Device",
      "browser": "Firefox 121.0",
      "ip_address": "203.0.113.45",
      "location": "Singapore",
      "status": "failed",
      "created_at": "2024-12-26T22:15:00Z",
      "failure_reason": "invalid_password"
    },
    {
      "id": "login-uuid-3",
      "device": "Windows PC",
      "browser": "Edge 120.0.2210",
      "ip_address": "203.0.113.100",
      "location": "Malaysia",
      "status": "failed",
      "created_at": "2024-12-26T18:30:00Z",
      "failure_reason": "2fa_failed"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "total_pages": 3
  }
}
```

**Failure Reasons:**
| Code | Description |
|------|-------------|
| `invalid_password` | Wrong password entered |
| `invalid_email` | Email not found |
| `2fa_failed` | 2FA code incorrect |
| `account_locked` | Account temporarily locked |
| `ip_blocked` | IP address blocked |

---

## 5. Database Schema

### Table: `user_sessions`
```sql
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    device VARCHAR(100),
    browser VARCHAR(100),
    ip_address INET,
    location VARCHAR(200),
    user_agent TEXT,
    last_active_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    INDEX idx_sessions_user_id (user_id),
    INDEX idx_sessions_token (token_hash),
    INDEX idx_sessions_expires (expires_at)
);
```

### Table: `login_history`
```sql
CREATE TABLE login_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255), -- Store email even for failed attempts
    device VARCHAR(100),
    browser VARCHAR(100),
    ip_address INET,
    location VARCHAR(200),
    user_agent TEXT,
    status VARCHAR(20) NOT NULL, -- 'success', 'failed'
    failure_reason VARCHAR(50), -- 'invalid_password', 'invalid_email', '2fa_failed', etc.
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    INDEX idx_login_history_user_id (user_id),
    INDEX idx_login_history_created (created_at),
    INDEX idx_login_history_ip (ip_address)
);
```

### Table: `user_2fa`
```sql
CREATE TABLE user_2fa (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    secret_key VARCHAR(100) NOT NULL, -- Encrypted TOTP secret
    is_enabled BOOLEAN DEFAULT false,
    backup_codes JSONB, -- Encrypted array of backup codes
    enabled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### User table additions:
```sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_login_attempts INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMPTZ;
```

---

## 6. Security Best Practices

### Password Security
1. **Hashing**: Use bcrypt with cost factor 12+
2. **History**: Prevent reuse of last 5 passwords
3. **Complexity**: Minimum 8 chars, recommend mix of upper/lower/numbers/symbols
4. **Rate Limiting**: Max 5 failed attempts per hour

### Session Security
1. **Token Rotation**: Rotate session tokens regularly
2. **Secure Cookies**: Use HttpOnly, Secure, SameSite=Strict
3. **IP Binding**: Optional - bind session to IP range
4. **Maximum Sessions**: Limit to 5 concurrent devices

### 2FA Security
1. **TOTP**: Use RFC 6238 compliant TOTP
2. **Backup Codes**: Generate 5-10 one-time backup codes
3. **Recovery**: Require identity verification for recovery
4. **Rate Limiting**: Max 3 2FA attempts per minute

### Login Security
1. **Rate Limiting**: Max 5 login attempts per 15 minutes
2. **Account Lockout**: Lock for 30 minutes after 10 failures
3. **Logging**: Log all attempts with IP and device
4. **Alerts**: Email user on suspicious activity

### API Security Headers
```go
// Security middleware
func SecurityHeaders(c *fiber.Ctx) error {
    c.Set("X-Content-Type-Options", "nosniff")
    c.Set("X-Frame-Options", "DENY")
    c.Set("X-XSS-Protection", "1; mode=block")
    c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
    c.Set("Content-Security-Policy", "default-src 'self'")
    return c.Next()
}
```

---

## 📊 Summary of Required Endpoints

| Method | Endpoint | Description | Priority |
|--------|----------|-------------|----------|
| PUT | `/auth/password` | Change password | High |
| POST | `/auth/2fa/enable` | Enable 2FA | Medium |
| POST | `/auth/2fa/verify` | Verify 2FA setup | Medium |
| POST | `/auth/2fa/disable` | Disable 2FA | Medium |
| GET | `/auth/2fa/status` | Check 2FA status | Medium |
| GET | `/auth/sessions` | List active sessions | High |
| DELETE | `/auth/sessions/:id` | Revoke single session | High |
| DELETE | `/auth/sessions` | Revoke all sessions | High |
| GET | `/auth/login-history` | Get login history | Medium |

---

## 🔧 Environment Variables

```env
# Password Security
PASSWORD_MIN_LENGTH=8
PASSWORD_BCRYPT_COST=12
PASSWORD_HISTORY_COUNT=5

# Session Security
SESSION_MAX_AGE=7d
SESSION_MAX_DEVICES=5
SESSION_ROTATE_INTERVAL=1h

# 2FA Security
TOTP_ISSUER=NexFlux
TOTP_PERIOD=30
TOTP_DIGITS=6
BACKUP_CODES_COUNT=5

# Rate Limiting
LOGIN_RATE_LIMIT=5/15m
PASSWORD_RATE_LIMIT=3/1h
TWO_FA_RATE_LIMIT=3/1m

# GeoIP (Optional)
GEOIP_DATABASE_PATH=/path/to/GeoLite2-City.mmdb
```

---

*Last Updated: 2024-12-27*
*Frontend Version: NexFlux v1.0*
