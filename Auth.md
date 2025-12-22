# NexFlux Authentication API Specification

## Overview

This document describes the Backend API endpoints required for the NexFlux authentication system, including registration, login, password reset, and token management.

## Base URL

```
/api/auth
```

---

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    username VARCHAR(100) UNIQUE,
    avatar_url TEXT,
    role ENUM('student', 'instructor', 'admin') DEFAULT 'student',
    
    -- Profile fields
    bio TEXT,
    location VARCHAR(255),
    phone VARCHAR(50),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    
    -- Gamification
    level INTEGER DEFAULT 1,
    total_xp INTEGER DEFAULT 0,
    streak_days INTEGER DEFAULT 0,
    last_activity_date DATE,
    
    -- Account status
    is_verified BOOLEAN DEFAULT FALSE,
    is_pro BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Social links (JSON)
    social_links JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
```

### Password Reset Tokens Table

```sql
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Index for token lookup
CREATE INDEX idx_reset_tokens_token ON password_reset_tokens(token);
CREATE INDEX idx_reset_tokens_user ON password_reset_tokens(user_id);
```

### Email Verification Tokens Table

```sql
CREATE TABLE email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_verification_tokens_token ON email_verification_tokens(token);
```

---

## API Endpoints

### 1. Register User

**POST** `/auth/register`

Creates a new user account.

#### Request Body

```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "name": "John Doe"
}
```

#### Response

**201 Created**
```json
{
  "success": true,
  "message": "Registration successful. Please check your email to verify your account.",
  "data": {
    "user": {
      "id": "uuid-here",
      "email": "user@example.com",
      "name": "John Doe",
      "role": "student"
    }
  }
}
```

**400 Bad Request**
```json
{
  "success": false,
  "message": "Email already registered"
}
```

#### Validation Rules
- `email`: Required, valid email format, unique
- `password`: Required, min 8 characters, must contain uppercase, lowercase, number
- `name`: Required, min 2 characters

---

### 2. Login

**POST** `/auth/login`

Authenticates user and returns JWT token.

#### Request Body

```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

#### Response

**200 OK**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "uuid-here",
      "email": "user@example.com",
      "name": "John Doe",
      "username": "johndoe",
      "avatar_url": "https://...",
      "role": "student",
      "level": 5,
      "total_xp": 1250,
      "streak_days": 7,
      "is_pro": false
    }
  }
}
```

**401 Unauthorized**
```json
{
  "success": false,
  "message": "Invalid email or password"
}
```

---

### 3. Forgot Password

**POST** `/auth/forgot-password`

Initiates password reset flow by sending reset email.

#### Request Body

```json
{
  "email": "user@example.com"
}
```

#### Response

**200 OK**
```json
{
  "success": true,
  "message": "If an account with that email exists, a password reset link has been sent."
}
```

> **Note**: Always return success message to prevent email enumeration attacks.

#### Backend Logic

1. Check if email exists in database
2. If exists:
   - Generate secure random token (32 bytes, hex encoded)
   - Create password_reset_tokens record with 1-hour expiry
   - Send email with reset link: `{FRONTEND_URL}/reset-password?token={token}`
3. If not exists:
   - Still return success message (security best practice)
   - Optionally log the attempt

#### Email Template

```html
Subject: Reset Your NexFlux Password

Hi {name},

You recently requested to reset your password for your NexFlux account.

Click the button below to reset it:

[Reset Password Button -> {FRONTEND_URL}/reset-password?token={token}]

This link will expire in 1 hour.

If you didn't request a password reset, please ignore this email.

Thanks,
The NexFlux Team
```

---

### 4. Verify Reset Token

**GET** `/auth/verify-reset-token`

Validates if a password reset token is valid and not expired.

#### Query Parameters

| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| token     | string | Yes      | Reset token from URL |

#### Response

**200 OK**
```json
{
  "success": true,
  "message": "Token is valid"
}
```

**400 Bad Request**
```json
{
  "success": false,
  "message": "Invalid or expired token"
}
```

#### Validation Checks
- Token exists in database
- Token has not been used (`used_at` is null)
- Token has not expired (`expires_at > NOW()`)

---

### 5. Reset Password

**POST** `/auth/reset-password`

Resets user password using valid token.

#### Request Body

```json
{
  "token": "reset-token-here",
  "password": "NewSecurePass123!"
}
```

#### Response

**200 OK**
```json
{
  "success": true,
  "message": "Password has been reset successfully"
}
```

**400 Bad Request**
```json
{
  "success": false,
  "message": "Invalid or expired token"
}
```

#### Backend Logic

1. Validate token (exists, not used, not expired)
2. Hash new password with bcrypt (cost factor 12)
3. Update user's password_hash
4. Mark token as used (`used_at = NOW()`)
5. Optionally invalidate all existing sessions
6. Send confirmation email

---

### 6. Verify Email

**GET** `/auth/verify-email`

Verifies user's email address.

#### Query Parameters

| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| token     | string | Yes      | Verification token |

#### Response

**200 OK**
```json
{
  "success": true,
  "message": "Email verified successfully"
}
```

**400 Bad Request**
```json
{
  "success": false,
  "message": "Invalid or expired verification link"
}
```

---

### 7. Resend Verification Email

**POST** `/auth/resend-verification`

Resends email verification link.

#### Request Body

```json
{
  "email": "user@example.com"
}
```

#### Response

**200 OK**
```json
{
  "success": true,
  "message": "Verification email sent"
}
```

---

### 8. Change Password (Authenticated)

**PUT** `/auth/change-password`

Changes password for authenticated user.

#### Headers

```
Authorization: Bearer <token>
```

#### Request Body

```json
{
  "current_password": "OldPassword123!",
  "new_password": "NewPassword456!"
}
```

#### Response

**200 OK**
```json
{
  "success": true,
  "message": "Password changed successfully"
}
```

**400 Bad Request**
```json
{
  "success": false,
  "message": "Current password is incorrect"
}
```

---

### 9. Logout

**POST** `/auth/logout`

Logs out user (invalidates token if using token blacklist).

#### Headers

```
Authorization: Bearer <token>
```

#### Response

**200 OK**
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

---

### 10. Refresh Token

**POST** `/auth/refresh-token`

Refreshes JWT token.

#### Request Body

```json
{
  "refresh_token": "refresh-token-here"
}
```

#### Response

**200 OK**
```json
{
  "success": true,
  "data": {
    "token": "new-jwt-token",
    "refresh_token": "new-refresh-token",
    "expires_in": 3600
  }
}
```

---

### 11. Get Current User

**GET** `/auth/me`

Returns currently authenticated user's profile.

#### Headers

```
Authorization: Bearer <token>
```

#### Response

**200 OK**
```json
{
  "success": true,
  "data": {
    "id": "uuid-here",
    "email": "user@example.com",
    "name": "John Doe",
    "username": "johndoe",
    "avatar_url": "https://...",
    "role": "student",
    "bio": "IoT enthusiast",
    "location": "Jakarta, Indonesia",
    "level": 5,
    "total_xp": 1250,
    "streak_days": 7,
    "is_pro": false,
    "is_verified": true,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

---

## JWT Token Structure

### Payload

```json
{
  "sub": "user-uuid",
  "email": "user@example.com",
  "role": "student",
  "iat": 1703239200,
  "exp": 1703242800
}
```

### Configuration

- **Algorithm**: HS256 or RS256 (recommended for production)
- **Access Token Expiry**: 1 hour
- **Refresh Token Expiry**: 7 days

---

## Security Recommendations

### Password Hashing
- Use bcrypt with cost factor 12+
- Never store plain text passwords

### Token Generation
- Use cryptographically secure random generator
- Minimum 32 bytes (64 hex characters)

### Rate Limiting
| Endpoint | Limit |
|----------|-------|
| `/auth/login` | 5 attempts per 15 minutes per IP |
| `/auth/register` | 3 attempts per hour per IP |
| `/auth/forgot-password` | 3 attempts per hour per email |

### Headers
```
X-RateLimit-Limit: 5
X-RateLimit-Remaining: 3
X-RateLimit-Reset: 1703239200
```

### CORS Configuration
```javascript
{
  origin: ['https://app.nexflux.io', 'http://localhost:5173'],
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'],
  allowedHeaders: ['Content-Type', 'Authorization']
}
```

---

## Error Response Format

All error responses follow this format:

```json
{
  "success": false,
  "message": "Human-readable error message",
  "error": {
    "code": "ERROR_CODE",
    "details": {}
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_CREDENTIALS` | 401 | Wrong email/password |
| `EMAIL_EXISTS` | 400 | Email already registered |
| `INVALID_TOKEN` | 400 | Token invalid or expired |
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `RATE_LIMITED` | 429 | Too many requests |
| `UNAUTHORIZED` | 401 | Missing or invalid auth token |
| `FORBIDDEN` | 403 | Insufficient permissions |

---

## Environment Variables

```bash
# JWT
JWT_SECRET=your-super-secret-key-min-32-chars
JWT_EXPIRES_IN=1h
JWT_REFRESH_EXPIRES_IN=7d

# Email (for password reset)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=your-email@example.com
SMTP_PASS=your-email-password
FROM_EMAIL=noreply@nexflux.io
FROM_NAME=NexFlux

# Frontend URL (for email links)
FRONTEND_URL=https://app.nexflux.io

# Password Reset
RESET_TOKEN_EXPIRES_IN=3600000 # 1 hour in milliseconds
```

---

## Implementation Checklist

- [ ] POST `/auth/register` - User registration
- [ ] POST `/auth/login` - User login
- [ ] POST `/auth/forgot-password` - Request password reset
- [ ] GET `/auth/verify-reset-token` - Validate reset token
- [ ] POST `/auth/reset-password` - Reset password with token
- [ ] GET `/auth/verify-email` - Verify email address
- [ ] POST `/auth/resend-verification` - Resend verification email
- [ ] PUT `/auth/change-password` - Change password (authenticated)
- [ ] POST `/auth/logout` - Logout user
- [ ] POST `/auth/refresh-token` - Refresh access token
- [ ] GET `/auth/me` - Get current user profile
