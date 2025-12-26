-- Migration: Circuit Simulator Tables
-- Description: Creates tables for circuits and circuit templates
-- Date: 2025-12-26

-- Enable UUID extension if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ==================================================
-- Table: circuits
-- User's circuit schematics
-- ==================================================
CREATE TABLE IF NOT EXISTS circuits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    
    -- Basic Info
    name VARCHAR(255) NOT NULL,
    description TEXT,
    thumbnail_url VARCHAR(500),
    
    -- Circuit Schema (JSON)
    schema_data JSONB NOT NULL DEFAULT '{}',
    
    -- Stats
    components_count INTEGER DEFAULT 0,
    wires_count INTEGER DEFAULT 0,
    
    -- Flags
    is_template BOOLEAN DEFAULT FALSE,
    is_public BOOLEAN DEFAULT FALSE,
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_circuits_user_id ON circuits(user_id);
CREATE INDEX IF NOT EXISTS idx_circuits_project_id ON circuits(project_id);
CREATE INDEX IF NOT EXISTS idx_circuits_is_template ON circuits(is_template) WHERE is_template = TRUE;
CREATE INDEX IF NOT EXISTS idx_circuits_is_public ON circuits(is_public) WHERE is_public = TRUE;
CREATE INDEX IF NOT EXISTS idx_circuits_created_at ON circuits(created_at DESC);

-- ==================================================
-- Table: circuit_templates
-- Pre-built circuit templates for learning
-- ==================================================
CREATE TABLE IF NOT EXISTS circuit_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Info
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,  -- 'beginner', 'intermediate', 'advanced'
    thumbnail_url VARCHAR(500),
    
    -- Schema
    schema_data JSONB NOT NULL,
    
    -- Learning
    difficulty VARCHAR(50) DEFAULT 'Beginner',
    estimated_time_minutes INTEGER DEFAULT 15,
    xp_reward INTEGER DEFAULT 10,
    
    -- Tags
    tags TEXT[] DEFAULT '{}',
    
    -- Stats
    use_count INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_circuit_templates_category ON circuit_templates(category);
CREATE INDEX IF NOT EXISTS idx_circuit_templates_difficulty ON circuit_templates(difficulty);
CREATE INDEX IF NOT EXISTS idx_circuit_templates_use_count ON circuit_templates(use_count DESC);

-- ==================================================
-- Seed Initial Templates
-- ==================================================
INSERT INTO circuit_templates (name, description, category, difficulty, schema_data, tags, xp_reward, estimated_time_minutes)
VALUES 
(
    'LED Blink',
    'Simple LED circuit with resistor. Perfect for beginners!',
    'beginner',
    'Beginner',
    '{"components":[{"id":"pwr","type":"power_source","name":"5V Power","position":{"x":100,"y":100},"properties":{"voltage":5}},{"id":"r1","type":"resistor","name":"R1","position":{"x":250,"y":100},"properties":{"resistance":220}},{"id":"led1","type":"led","name":"LED1","position":{"x":400,"y":100},"properties":{"color":"#ff0000"}},{"id":"gnd","type":"ground","name":"GND","position":{"x":550,"y":100}}],"wires":[{"id":"w1","startComponentId":"pwr","startPinId":"vcc","endComponentId":"r1","endPinId":"p1"},{"id":"w2","startComponentId":"r1","startPinId":"p2","endComponentId":"led1","endPinId":"anode"},{"id":"w3","startComponentId":"led1","startPinId":"cathode","endComponentId":"gnd","endPinId":"gnd"}]}'::jsonb,
    ARRAY['led', 'basic', 'beginner'],
    10,
    10
),
(
    'Traffic Light',
    'Create a traffic light simulation with 3 LEDs (Red, Yellow, Green)',
    'beginner',
    'Beginner',
    '{"components":[{"id":"pwr","type":"power_source","name":"5V Power","position":{"x":100,"y":200},"properties":{"voltage":5}},{"id":"r1","type":"resistor","name":"R1","position":{"x":200,"y":100},"properties":{"resistance":220}},{"id":"r2","type":"resistor","name":"R2","position":{"x":200,"y":200},"properties":{"resistance":220}},{"id":"r3","type":"resistor","name":"R3","position":{"x":200,"y":300},"properties":{"resistance":220}},{"id":"led1","type":"led","name":"RED","position":{"x":350,"y":100},"properties":{"color":"#ff0000"}},{"id":"led2","type":"led","name":"YELLOW","position":{"x":350,"y":200},"properties":{"color":"#ffff00"}},{"id":"led3","type":"led","name":"GREEN","position":{"x":350,"y":300},"properties":{"color":"#00ff00"}},{"id":"gnd","type":"ground","name":"GND","position":{"x":500,"y":200}}],"wires":[]}'::jsonb,
    ARRAY['led', 'traffic', 'beginner'],
    15,
    15
),
(
    'Temperature Sensor with DHT22',
    'Read temperature and humidity using DHT22 sensor with Arduino',
    'intermediate',
    'Intermediate',
    '{"components":[{"id":"arduino","type":"arduino_uno","name":"Arduino Uno","position":{"x":300,"y":200}},{"id":"dht22","type":"dht22","name":"DHT22","position":{"x":500,"y":200},"properties":{"temperature":25,"humidity":60}}],"wires":[]}'::jsonb,
    ARRAY['sensor', 'arduino', 'dht22', 'iot'],
    25,
    20
),
(
    'DC Motor Control',
    'Control a DC motor speed with PWM using Arduino',
    'intermediate',
    'Intermediate',
    '{"components":[{"id":"arduino","type":"arduino_uno","name":"Arduino Uno","position":{"x":200,"y":200}},{"id":"motor","type":"dc_motor","name":"DC Motor","position":{"x":500,"y":200},"properties":{"rpm":0,"maxRpm":3000}}],"wires":[]}'::jsonb,
    ARRAY['motor', 'pwm', 'arduino', 'actuator'],
    30,
    25
)
ON CONFLICT DO NOTHING;

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_circuits_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_circuits_updated_at ON circuits;
CREATE TRIGGER update_circuits_updated_at
    BEFORE UPDATE ON circuits
    FOR EACH ROW
    EXECUTE FUNCTION update_circuits_updated_at();
