-- Migration: Security Features
-- Description: Creates tables for security settings, sessions, login history, 2FA, and security logs
-- Date: 2025-12-27

-- Enable UUID extension if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ==================================================
-- Add security columns to users table
-- ==================================================
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_login_attempts INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS two_factor_enabled BOOLEAN DEFAULT FALSE;

-- ==================================================
-- Update user_sessions table structure
-- ==================================================
-- Drop old columns if exist and add new ones
ALTER TABLE user_sessions ADD COLUMN IF NOT EXISTS token_hash VARCHAR(255);
ALTER TABLE user_sessions ADD COLUMN IF NOT EXISTS device VARCHAR(100);
ALTER TABLE user_sessions ADD COLUMN IF NOT EXISTS browser VARCHAR(100);
ALTER TABLE user_sessions ADD COLUMN IF NOT EXISTS location VARCHAR(200);
ALTER TABLE user_sessions ADD COLUMN IF NOT EXISTS user_agent TEXT;
ALTER TABLE user_sessions ADD COLUMN IF NOT EXISTS last_active_at TIMESTAMPTZ DEFAULT NOW();

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_user_sessions_token_hash ON user_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires ON user_sessions(expires_at);

-- ==================================================
-- Table: login_history
-- Records all login attempts (success and fail)
-- ==================================================
CREATE TABLE IF NOT EXISTS login_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    device VARCHAR(100),
    browser VARCHAR(100),
    ip_address VARCHAR(45),
    location VARCHAR(200),
    user_agent TEXT,
    status VARCHAR(20) NOT NULL CHECK (status IN ('success', 'failed')),
    failure_reason VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_login_history_user_id ON login_history(user_id);
CREATE INDEX IF NOT EXISTS idx_login_history_email ON login_history(email);
CREATE INDEX IF NOT EXISTS idx_login_history_created ON login_history(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_login_history_ip ON login_history(ip_address);

-- ==================================================
-- Table: user_2fa
-- Two-Factor Authentication settings
-- ==================================================
CREATE TABLE IF NOT EXISTS user_2fa (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    secret_key VARCHAR(100) NOT NULL,
    is_enabled BOOLEAN DEFAULT FALSE,
    backup_codes JSONB DEFAULT '[]',
    enabled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_2fa_user_id ON user_2fa(user_id);

-- ==================================================
-- Table: password_history
-- Tracks password changes to prevent reuse
-- ==================================================
CREATE TABLE IF NOT EXISTS password_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_history_user_id ON password_history(user_id);
CREATE INDEX IF NOT EXISTS idx_password_history_created ON password_history(created_at DESC);

-- ==================================================
-- Table: security_logs
-- Security event audit trail
-- ==================================================
CREATE TABLE IF NOT EXISTS security_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    details JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_security_logs_user_id ON security_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_security_logs_event_type ON security_logs(event_type);
CREATE INDEX IF NOT EXISTS idx_security_logs_created ON security_logs(created_at DESC);

-- ==================================================
-- Trigger for updated_at on user_2fa
-- ==================================================
CREATE OR REPLACE FUNCTION update_user_2fa_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_user_2fa_updated_at ON user_2fa;
CREATE TRIGGER update_user_2fa_updated_at
    BEFORE UPDATE ON user_2fa
    FOR EACH ROW
    EXECUTE FUNCTION update_user_2fa_updated_at();
