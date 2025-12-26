-- Migration: Add Project Progress & XP System Tables
-- Description: Creates tables for project progress tracking, milestones, XP transactions, and simulation results
-- Date: 2025-12-26

-- Enable UUID extension if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ==================================================
-- Table: project_progress
-- Tracks progress for each component of a project
-- ==================================================
CREATE TABLE IF NOT EXISTS project_progress (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    component VARCHAR(50) NOT NULL, -- 'schema', 'code', 'simulation', 'verification'
    is_complete BOOLEAN DEFAULT false,
    completion_percentage INTEGER DEFAULT 0 CHECK (completion_percentage >= 0 AND completion_percentage <= 100),
    completed_at TIMESTAMPTZ,
    data JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, component)
);

CREATE INDEX IF NOT EXISTS idx_project_progress_project_id ON project_progress(project_id);
CREATE INDEX IF NOT EXISTS idx_project_progress_component ON project_progress(component);

-- ==================================================
-- Table: project_milestones
-- Tracks milestones achieved for a project
-- ==================================================
CREATE TABLE IF NOT EXISTS project_milestones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    milestone_type VARCHAR(100) NOT NULL, -- 'first_schema_save', 'first_code_save', etc.
    xp_earned INTEGER DEFAULT 0,
    unlocked_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, milestone_type)
);

CREATE INDEX IF NOT EXISTS idx_project_milestones_project_id ON project_milestones(project_id);
CREATE INDEX IF NOT EXISTS idx_project_milestones_user_id ON project_milestones(user_id);
CREATE INDEX IF NOT EXISTS idx_project_milestones_type ON project_milestones(milestone_type);

-- ==================================================
-- Table: user_xp_transactions
-- Tracks all XP transactions for audit trail
-- ==================================================
CREATE TABLE IF NOT EXISTS user_xp_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    xp_amount INTEGER NOT NULL,
    xp_type VARCHAR(50) NOT NULL, -- 'project', 'challenge', 'lab', 'daily'
    source_id UUID, -- project_id, challenge_id, etc.
    source_type VARCHAR(50), -- 'project', 'challenge', 'lab_session'
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_xp_transactions_user_id ON user_xp_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_xp_transactions_source_id ON user_xp_transactions(source_id);
CREATE INDEX IF NOT EXISTS idx_user_xp_transactions_xp_type ON user_xp_transactions(xp_type);
CREATE INDEX IF NOT EXISTS idx_user_xp_transactions_created_at ON user_xp_transactions(created_at DESC);

-- ==================================================
-- Table: project_simulation_results
-- Stores simulation results
-- ==================================================
CREATE TABLE IF NOT EXISTS project_simulation_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL, -- 'success', 'failed', 'timeout', 'running'
    duration_ms INTEGER DEFAULT 0,
    results JSONB DEFAULT '{}',
    errors JSONB DEFAULT '[]',
    warnings JSONB DEFAULT '[]',
    xp_earned INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_project_simulation_results_project_id ON project_simulation_results(project_id);
CREATE INDEX IF NOT EXISTS idx_project_simulation_results_user_id ON project_simulation_results(user_id);
CREATE INDEX IF NOT EXISTS idx_project_simulation_results_status ON project_simulation_results(status);

-- ==================================================
-- Add columns to users table for XP system
-- ==================================================
DO $$ 
BEGIN
    -- Add current_xp column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='users' AND column_name='current_xp') THEN
        ALTER TABLE users ADD COLUMN current_xp INTEGER DEFAULT 0;
    END IF;

    -- Add total_xp column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='users' AND column_name='total_xp') THEN
        ALTER TABLE users ADD COLUMN total_xp INTEGER DEFAULT 0;
    END IF;

    -- Add level column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='users' AND column_name='level') THEN
        ALTER TABLE users ADD COLUMN level INTEGER DEFAULT 1;
    END IF;

    -- Add target_xp column (XP needed to reach next level)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='users' AND column_name='target_xp') THEN
        ALTER TABLE users ADD COLUMN target_xp INTEGER DEFAULT 1000;
    END IF;
END $$;

-- ==================================================
-- Add columns to projects table for verification
-- ==================================================
DO $$ 
BEGIN
    -- Add is_verified column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='projects' AND column_name='is_verified') THEN
        ALTER TABLE projects ADD COLUMN is_verified BOOLEAN DEFAULT false;
    END IF;

    -- Add verified_at column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='projects' AND column_name='verified_at') THEN
        ALTER TABLE projects ADD COLUMN verified_at TIMESTAMPTZ;
    END IF;

    -- Add verified_by column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='projects' AND column_name='verified_by') THEN
        ALTER TABLE projects ADD COLUMN verified_by UUID REFERENCES users(id);
    END IF;
END $$;

-- ==================================================
-- Add indexes for existing columns
-- ==================================================
CREATE INDEX IF NOT EXISTS idx_users_level ON users(level);
CREATE INDEX IF NOT EXISTS idx_users_total_xp ON users(total_xp DESC);
CREATE INDEX IF NOT EXISTS idx_projects_progress ON projects(progress);
CREATE INDEX IF NOT EXISTS idx_projects_completed_at ON projects(completed_at);

-- ==================================================
-- Update function for updated_at timestamp
-- ==================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to project_progress
DROP TRIGGER IF EXISTS update_project_progress_updated_at ON project_progress;
CREATE TRIGGER update_project_progress_updated_at
    BEFORE UPDATE ON project_progress
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Success message
-- SELECT 'Migration completed: Project Progress & XP System tables created successfully';
