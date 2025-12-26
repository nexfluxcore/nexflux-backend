-- Migration: Simulation Management Tables
-- Description: Creates tables for simulations and simulation runs
-- Date: 2025-12-26

-- Enable UUID extension if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ==================================================
-- Table: simulations
-- User's simulation projects
-- ==================================================
CREATE TABLE IF NOT EXISTS simulations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    
    -- Basic Info
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(100) DEFAULT 'Basic Electronics',  -- IoT, Power Electronics, Wireless, etc.
    thumbnail_url VARCHAR(500),
    
    -- Simulation Data (JSON)
    schema_data JSONB,           -- Circuit schema (components, wires, positions)
    simulation_settings JSONB,   -- Speed, time step, etc.
    last_result JSONB,           -- Last simulation result
    
    -- Stats
    components_count INTEGER DEFAULT 0,
    wires_count INTEGER DEFAULT 0,
    run_count INTEGER DEFAULT 0,
    total_runtime_ms BIGINT DEFAULT 0,
    
    -- Status
    status VARCHAR(50) DEFAULT 'draft',  -- draft, running, completed, paused, error
    error_message TEXT,
    
    -- Timestamps
    last_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_simulations_user_id ON simulations(user_id);
CREATE INDEX IF NOT EXISTS idx_simulations_project_id ON simulations(project_id);
CREATE INDEX IF NOT EXISTS idx_simulations_status ON simulations(status);
CREATE INDEX IF NOT EXISTS idx_simulations_type ON simulations(type);
CREATE INDEX IF NOT EXISTS idx_simulations_created_at ON simulations(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_simulations_last_run_at ON simulations(last_run_at DESC);

-- ==================================================
-- Table: simulation_runs
-- History of simulation runs
-- ==================================================
CREATE TABLE IF NOT EXISTS simulation_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    simulation_id UUID NOT NULL REFERENCES simulations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Run Data
    status VARCHAR(50) NOT NULL,  -- running, completed, error
    duration_ms INTEGER DEFAULT 0,
    
    -- Results (JSON)
    result_data JSONB,            -- voltage_nodes, current_branches, component_states
    errors JSONB DEFAULT '[]',
    warnings JSONB DEFAULT '[]',
    
    -- Timestamps
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_simulation_runs_simulation_id ON simulation_runs(simulation_id);
CREATE INDEX IF NOT EXISTS idx_simulation_runs_user_id ON simulation_runs(user_id);
CREATE INDEX IF NOT EXISTS idx_simulation_runs_status ON simulation_runs(status);
CREATE INDEX IF NOT EXISTS idx_simulation_runs_started_at ON simulation_runs(started_at DESC);

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_simulations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_simulations_updated_at ON simulations;
CREATE TRIGGER update_simulations_updated_at
    BEFORE UPDATE ON simulations
    FOR EACH ROW
    EXECUTE FUNCTION update_simulations_updated_at();
