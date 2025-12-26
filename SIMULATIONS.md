# NexFlux Simulations Backend Documentation

## 📋 Overview

Module Simulations menyimpan dan mengelola data simulasi rangkaian yang dibuat oleh pengguna. Setiap simulasi terhubung ke project dan menyimpan skema rangkaian, hasil simulasi, dan metadata.

---

## 🗄️ Database Schema

### Table: `simulations`

```sql
CREATE TABLE simulations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
    last_run_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_simulations_user_id ON simulations(user_id);
CREATE INDEX idx_simulations_project_id ON simulations(project_id);
CREATE INDEX idx_simulations_status ON simulations(status);
CREATE INDEX idx_simulations_type ON simulations(type);
CREATE INDEX idx_simulations_created_at ON simulations(created_at DESC);
```

### Table: `simulation_runs` (History)

```sql
CREATE TABLE simulation_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    simulation_id UUID NOT NULL REFERENCES simulations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Run Data
    status VARCHAR(50) NOT NULL,  -- running, completed, error
    duration_ms INTEGER,
    
    -- Results (JSON)
    result_data JSONB,            -- voltage_nodes, current_branches, component_states
    errors JSONB,
    warnings JSONB,
    
    -- Timestamps
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_simulation_runs_simulation_id ON simulation_runs(simulation_id);
CREATE INDEX idx_simulation_runs_started_at ON simulation_runs(started_at DESC);
```

---

## 📡 API Endpoints

### Base URL
```
/api/v1/simulations
```

---

### 1. List Simulations

**GET** `/api/v1/simulations`

Query Parameters:
| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | int | Page number (default: 1) |
| `limit` | int | Items per page (default: 20, max: 100) |
| `status` | string | Filter by status: `draft`, `running`, `completed`, `paused`, `error` |
| `type` | string | Filter by type: `Basic Electronics`, `IoT`, `Power Electronics`, etc. |
| `project_id` | uuid | Filter by project |
| `search` | string | Search by name |
| `sort` | string | Sort: `created_at`, `updated_at`, `last_run_at`, `name` |
| `order` | string | Order: `asc`, `desc` (default: desc) |

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "LED Blink Circuit",
      "description": "Simple LED blinking simulation",
      "type": "Basic Electronics",
      "status": "completed",
      "thumbnail_url": "https://...",
      "components_count": 5,
      "wires_count": 6,
      "run_count": 12,
      "total_runtime_ms": 45000,
      "last_run_at": "2024-01-15T10:30:00Z",
      "project_id": "uuid",
      "project_name": "My First Project",
      "created_at": "2024-01-10T09:00:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "total_pages": 3
  }
}
```

---

### 2. Get Simulation Stats

**GET** `/api/v1/simulations/stats`

**Response:**
```json
{
  "success": true,
  "data": {
    "total_simulations": 127,
    "running_now": 3,
    "completed": 98,
    "paused": 15,
    "error": 11,
    "total_runtime_hours": 48.5,
    "success_rate": 94.2,
    "simulations_this_week": 12,
    "by_type": {
      "Basic Electronics": 45,
      "IoT": 38,
      "Power Electronics": 22,
      "Wireless": 15,
      "Renewable Energy": 7
    }
  }
}
```

---

### 3. Get Simulation Detail

**GET** `/api/v1/simulations/:id`

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "LED Blink Circuit",
    "description": "Simple LED blinking simulation with Arduino",
    "type": "Basic Electronics",
    "status": "completed",
    "thumbnail_url": "https://...",
    "components_count": 5,
    "wires_count": 6,
    "run_count": 12,
    "total_runtime_ms": 45000,
    "last_run_at": "2024-01-15T10:30:00Z",
    "project_id": "uuid",
    "project_name": "My First Project",
    "schema_data": {
      "components": [...],
      "wires": [...],
      "groundNodeId": "...",
      "powerSourceIds": [...]
    },
    "simulation_settings": {
      "speed": 1,
      "timeStep": 16,
      "maxIterations": 100
    },
    "last_result": {
      "success": true,
      "duration": 1500,
      "voltage_nodes": {...},
      "current_branches": {...},
      "component_states": {...},
      "power_consumption": 0.025
    },
    "created_at": "2024-01-10T09:00:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### 4. Create Simulation

**POST** `/api/v1/simulations`

**Request Body:**
```json
{
  "name": "My New Simulation",
  "description": "Optional description",
  "type": "IoT",
  "project_id": "uuid (optional)",
  "schema_data": {
    "components": [...],
    "wires": [...],
    "groundNodeId": "...",
    "powerSourceIds": [...]
  },
  "simulation_settings": {
    "speed": 1,
    "timeStep": 16
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "new-uuid",
    "name": "My New Simulation",
    ...
  },
  "message": "Simulation created successfully"
}
```

---

### 5. Update Simulation

**PUT** `/api/v1/simulations/:id`

**Request Body:**
```json
{
  "name": "Updated Name",
  "description": "Updated description",
  "type": "Power Electronics",
  "schema_data": {...},
  "simulation_settings": {...}
}
```

**Response:**
```json
{
  "success": true,
  "data": {...},
  "message": "Simulation updated successfully"
}
```

---

### 6. Delete Simulation

**DELETE** `/api/v1/simulations/:id`

**Response:**
```json
{
  "success": true,
  "message": "Simulation deleted successfully"
}
```

---

### 7. Run Simulation

**POST** `/api/v1/simulations/:id/run`

Starts a new simulation run and returns the run ID.

**Request Body (optional):**
```json
{
  "duration_ms": 5000,
  "settings_override": {
    "speed": 2,
    "timeStep": 8
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "run_id": "run-uuid",
    "status": "running",
    "started_at": "2024-01-15T10:30:00Z"
  },
  "message": "Simulation started"
}
```

---

### 8. Stop Simulation

**POST** `/api/v1/simulations/:id/stop`

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "paused",
    "duration_ms": 3500
  },
  "message": "Simulation stopped"
}
```

---

### 9. Get Simulation Run History

**GET** `/api/v1/simulations/:id/runs`

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "run-uuid",
      "status": "completed",
      "duration_ms": 1500,
      "started_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T10:30:01Z",
      "result_data": {...},
      "errors": [],
      "warnings": []
    }
  ]
}
```

---

### 10. Save Simulation Result

**POST** `/api/v1/simulations/:id/result`

Called by frontend when client-side simulation completes.

**Request Body:**
```json
{
  "duration_ms": 1500,
  "result_data": {
    "success": true,
    "voltage_nodes": {...},
    "current_branches": {...},
    "component_states": {...},
    "power_consumption": 0.025
  },
  "errors": [],
  "warnings": []
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "run_id": "run-uuid",
    "xp_earned": 5
  },
  "message": "Simulation result saved"
}
```

---

## 🎮 Gamification

### XP Rewards

| Action | XP |
|--------|-----|
| Create first simulation | 10 |
| Complete simulation | 5 |
| Run 10 simulations | 25 (milestone) |
| Run 50 simulations | 100 (milestone) |
| Share simulation | 15 |
| Fix error simulation | 10 |

---

## 📊 Simulation Types

| Type | Description | Example |
|------|-------------|---------|
| `Basic Electronics` | Simple circuits with LED, resistors | LED blink |
| `IoT` | Arduino/ESP32 with sensors | Temperature monitor |
| `Power Electronics` | Motors, relays, power control | Motor driver |
| `Wireless` | RF, WiFi, Bluetooth modules | Remote control |
| `Renewable Energy` | Solar, wind power circuits | Solar tracker |
| `Audio` | Amplifiers, speakers | Audio amplifier |
| `Digital Logic` | Logic gates, flip-flops | Counter circuit |

---

## 🔌 WebSocket Events (Optional)

For real-time simulation status updates:

```
WS /ws/simulations/:id

Events:
- simulation:started
- simulation:progress { time_ms, component_states }
- simulation:completed { result_data }
- simulation:error { message }
- simulation:stopped
```

---

## 🌐 Environment Variables

```env
# Database
DATABASE_URL=postgresql://...

# Redis (for caching stats)
REDIS_URL=redis://...

# Feature Flags
ENABLE_CLOUD_SIMULATION=false
MAX_SIMULATION_DURATION_MS=300000

# Limits
MAX_SIMULATIONS_FREE=50
MAX_SIMULATIONS_PRO=unlimited
MAX_COMPONENTS_FREE=100
MAX_COMPONENTS_PRO=500
```

---

## 📝 Go Structs

```go
// internal/models/simulation.go

type Simulation struct {
    ID                 uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserID             uuid.UUID       `json:"user_id" gorm:"type:uuid;not null"`
    ProjectID          *uuid.UUID      `json:"project_id" gorm:"type:uuid"`
    Name               string          `json:"name" gorm:"size:255;not null"`
    Description        string          `json:"description"`
    Type               string          `json:"type" gorm:"size:100;default:'Basic Electronics'"`
    ThumbnailURL       string          `json:"thumbnail_url" gorm:"size:500"`
    SchemaData         datatypes.JSON  `json:"schema_data"`
    SimulationSettings datatypes.JSON  `json:"simulation_settings"`
    LastResult         datatypes.JSON  `json:"last_result"`
    ComponentsCount    int             `json:"components_count" gorm:"default:0"`
    WiresCount         int             `json:"wires_count" gorm:"default:0"`
    RunCount           int             `json:"run_count" gorm:"default:0"`
    TotalRuntimeMs     int64           `json:"total_runtime_ms" gorm:"default:0"`
    Status             string          `json:"status" gorm:"size:50;default:'draft'"`
    ErrorMessage       string          `json:"error_message"`
    LastRunAt          *time.Time      `json:"last_run_at"`
    CreatedAt          time.Time       `json:"created_at"`
    UpdatedAt          time.Time       `json:"updated_at"`
    
    // Relations
    User    *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Project *Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

type SimulationRun struct {
    ID           uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    SimulationID uuid.UUID       `json:"simulation_id" gorm:"type:uuid;not null"`
    UserID       uuid.UUID       `json:"user_id" gorm:"type:uuid;not null"`
    Status       string          `json:"status" gorm:"size:50;not null"`
    DurationMs   int             `json:"duration_ms"`
    ResultData   datatypes.JSON  `json:"result_data"`
    Errors       datatypes.JSON  `json:"errors"`
    Warnings     datatypes.JSON  `json:"warnings"`
    StartedAt    time.Time       `json:"started_at"`
    CompletedAt  *time.Time      `json:"completed_at"`
}

type SimulationStats struct {
    TotalSimulations    int            `json:"total_simulations"`
    RunningNow          int            `json:"running_now"`
    Completed           int            `json:"completed"`
    Paused              int            `json:"paused"`
    Error               int            `json:"error"`
    TotalRuntimeHours   float64        `json:"total_runtime_hours"`
    SuccessRate         float64        `json:"success_rate"`
    SimulationsThisWeek int            `json:"simulations_this_week"`
    ByType              map[string]int `json:"by_type"`
}
```

---

## ✅ Implementation Checklist

### Backend
- [ ] Create `simulations` table migration
- [ ] Create `simulation_runs` table migration
- [ ] Implement Simulation model
- [ ] Implement SimulationRun model
- [ ] Create SimulationRepository
- [ ] Create SimulationService
- [ ] Implement CRUD endpoints
- [ ] Implement stats endpoint
- [ ] Implement run/stop endpoints
- [ ] Add XP rewards logic
- [ ] Add WebSocket support (optional)

### Frontend
- [x] Update Simulations page to fetch from API
- [x] Implement filter by status/type
- [x] Implement stats from API
- [x] Connect to Circuit Simulator for create/edit
- [ ] Add search functionality
- [ ] Add pagination
- [ ] Add delete confirmation
