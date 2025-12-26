# NexFlux Circuit Simulator Module

## 📋 Overview

Module Circuit Simulator adalah implementasi custom untuk mendesain dan mensimulasikan rangkaian elektronik langsung di browser. Simulasi berjalan di **client-side** menggunakan custom engine, dengan opsi untuk menyimpan skema dan hasil ke server.

---

## 🏗️ Frontend Architecture

```
src/pages/CircuitSimulator/
├── engine/                    # Custom Simulation Engine
│   ├── types.ts              # Type definitions
│   └── CircuitEngine.ts      # Modified Nodal Analysis solver
├── store/
│   └── circuitStore.ts       # Zustand state management
├── nodes/
│   └── CircuitNodes.tsx      # Custom React Flow nodes
├── components/
│   ├── ComponentLibrary.tsx  # All available components
│   ├── ComponentPanel.tsx    # Left sidebar - component picker
│   ├── PropertiesPanel.tsx   # Right sidebar - property editor
│   └── SimulationControls.tsx # Play/pause/stop controls
└── index.tsx                 # Main page component
```

---

## 🔌 Komponen yang Tersedia

### Power
- ⚡ **Power Source** - DC voltage source (1-24V adjustable)
- 🔌 **Ground** - Circuit ground reference (0V)

### Passive Components
- **Resistor** - 1Ω to 10MΩ
- **Capacitor** - Stores electrical charge
- **Inductor** - Magnetic field energy storage
- **Potentiometer** - Variable resistor

### Active Components
- 💡 **LED** - Light Emitting Diode (animated brightness)
- **Diode** - One-way current flow
- **NPN/PNP Transistor** - BJT transistors
- **Switch** - Toggle ON/OFF
- **Push Button** - Momentary contact

### Sensors
- 🌡️ **DHT22** - Temperature & Humidity
- ☀️ **LDR** - Light Dependent Resistor
- 📏 **Ultrasonic** - Distance sensor

### Actuators
- ⚙️ **DC Motor** - Animated RPM display
- 🔊 **Buzzer** - Sound output
- **Servo Motor** - Positional control
- **Relay** - Electromagnetic switch

### Microcontrollers
- 🎛️ **Arduino Uno** - Full pinout (D0-13, A0-5, 5V, 3.3V, GND)
- 🎛️ **ESP32** - WiFi & BLE enabled

### Display
- 📺 **LCD 16x2** - Character display with I2C

---

## 🔧 Custom Simulation Engine

Engine menggunakan **Modified Nodal Analysis (MNA)** untuk menyelesaikan persamaan rangkaian:

### Key Features:
- **Gauss-Seidel iterative solver** untuk menghitung voltase di setiap node
- **Kirchhoff's Current Law** compliance
- **Real-time updates** pada 60fps
- **Error detection** (short circuit, overcurrent, overvoltage)
- **Component state tracking** (LED brightness, motor RPM, etc.)

### Simulation Flow:
```
1. Load Circuit → Initialize Node Voltages
2. Build Node Connections Map
3. Iterative Solver Loop:
   - Calculate voltage at each node
   - Apply Kirchhoff's Current Law
   - Check convergence
4. Calculate Branch Currents (Ohm's Law)
5. Update Component States (LED, Motor, etc.)
6. Check Circuit Health
7. Notify UI with results
```

---

## 📦 Dependencies

```json
{
  "@xyflow/react": "^12.x",
  "zustand": "^4.x"
}
```

---

# 🗄️ Backend Documentation

## Database Schema

### Table: `circuits`

Menyimpan skema rangkaian yang dibuat user (terpisah dari simulations untuk reusability).

```sql
CREATE TABLE circuits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    
    -- Basic Info
    name VARCHAR(255) NOT NULL,
    description TEXT,
    thumbnail_url VARCHAR(500),
    
    -- Circuit Schema (JSON)
    schema_data JSONB NOT NULL DEFAULT '{}',
    /*
    schema_data structure:
    {
        "components": [
            {
                "id": "comp_123",
                "type": "resistor",
                "name": "R1",
                "position": {"x": 100, "y": 200},
                "rotation": 0,
                "properties": {"resistance": 1000},
                "pins": [...]
            }
        ],
        "wires": [
            {
                "id": "wire_456",
                "startComponentId": "comp_123",
                "startPinId": "p1",
                "endComponentId": "comp_789",
                "endPinId": "p2",
                "resistance": 0.01
            }
        ],
        "groundNodeId": "comp_gnd_gnd",
        "powerSourceIds": ["comp_pwr"]
    }
    */
    
    -- Stats
    components_count INTEGER DEFAULT 0,
    wires_count INTEGER DEFAULT 0,
    
    -- Flags
    is_template BOOLEAN DEFAULT FALSE,
    is_public BOOLEAN DEFAULT FALSE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_circuits_user_id ON circuits(user_id);
CREATE INDEX idx_circuits_project_id ON circuits(project_id);
CREATE INDEX idx_circuits_is_template ON circuits(is_template) WHERE is_template = TRUE;
CREATE INDEX idx_circuits_is_public ON circuits(is_public) WHERE is_public = TRUE;
```

### Table: `circuit_templates`

Pre-built circuit templates untuk learning.

```sql
CREATE TABLE circuit_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
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
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_circuit_templates_category ON circuit_templates(category);
CREATE INDEX idx_circuit_templates_difficulty ON circuit_templates(difficulty);
```

---

## 📡 API Endpoints

### Base URL
```
/api/v1/circuits
```

---

### 1. List User's Circuits

**GET** `/api/v1/circuits`

Query Parameters:
| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | int | Page number (default: 1) |
| `limit` | int | Items per page (default: 20) |
| `project_id` | uuid | Filter by project |
| `search` | string | Search by name |

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "LED Blink Circuit",
      "description": "Simple LED blinking with resistor",
      "thumbnail_url": "https://...",
      "components_count": 5,
      "wires_count": 6,
      "project_id": "uuid",
      "project_name": "My Project",
      "is_template": false,
      "is_public": false,
      "created_at": "2024-01-10T09:00:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 15
  }
}
```

---

### 2. Get Circuit Detail

**GET** `/api/v1/circuits/:id`

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "LED Blink Circuit",
    "description": "Simple LED blinking with resistor",
    "thumbnail_url": "https://...",
    "schema_data": {
      "components": [
        {
          "id": "comp_1",
          "type": "power_source",
          "name": "Power Source",
          "position": {"x": 100, "y": 100},
          "rotation": 0,
          "properties": {"voltage": 5},
          "pins": [
            {"id": "vcc", "name": "VCC", "type": "output"},
            {"id": "gnd", "name": "GND", "type": "ground"}
          ],
          "isActive": true,
          "isBurned": false
        },
        {
          "id": "comp_2",
          "type": "resistor",
          "name": "R1",
          "position": {"x": 250, "y": 100},
          "rotation": 0,
          "properties": {"resistance": 220},
          "pins": [...]
        },
        {
          "id": "comp_3",
          "type": "led",
          "name": "LED1",
          "position": {"x": 400, "y": 100},
          "rotation": 0,
          "properties": {"color": "#ff0000", "forwardVoltage": 2},
          "pins": [...]
        },
        {
          "id": "comp_4",
          "type": "ground",
          "name": "GND",
          "position": {"x": 550, "y": 100},
          "pins": [...]
        }
      ],
      "wires": [
        {
          "id": "wire_1",
          "startComponentId": "comp_1",
          "startPinId": "vcc",
          "endComponentId": "comp_2",
          "endPinId": "p1",
          "resistance": 0.01
        }
      ],
      "groundNodeId": "comp_4_gnd",
      "powerSourceIds": ["comp_1"]
    },
    "components_count": 4,
    "wires_count": 3,
    "project_id": "uuid",
    "is_template": false,
    "is_public": false,
    "created_at": "2024-01-10T09:00:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### 3. Create Circuit

**POST** `/api/v1/circuits`

**Request Body:**
```json
{
  "name": "My New Circuit",
  "description": "Optional description",
  "project_id": "uuid (optional)",
  "schema_data": {
    "components": [...],
    "wires": [...],
    "groundNodeId": "...",
    "powerSourceIds": [...]
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "new-uuid",
    "name": "My New Circuit",
    ...
  },
  "message": "Circuit created successfully",
  "xp_earned": 10
}
```

---

### 4. Update Circuit (Auto-save)

**PUT** `/api/v1/circuits/:id`

**Request Body:**
```json
{
  "name": "Updated Name (optional)",
  "description": "Updated description (optional)",
  "schema_data": {
    "components": [...],
    "wires": [...],
    "groundNodeId": "...",
    "powerSourceIds": [...]
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {...},
  "message": "Circuit saved"
}
```

---

### 5. Delete Circuit

**DELETE** `/api/v1/circuits/:id`

**Response:**
```json
{
  "success": true,
  "message": "Circuit deleted successfully"
}
```

---

### 6. Duplicate Circuit

**POST** `/api/v1/circuits/:id/duplicate`

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "new-uuid",
    "name": "LED Blink Circuit (Copy)",
    ...
  },
  "message": "Circuit duplicated"
}
```

---

### 7. Generate Thumbnail

**POST** `/api/v1/circuits/:id/thumbnail`

Called after circuit is saved to generate preview image.

**Request Body:**
```json
{
  "image_data": "base64_encoded_png"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "thumbnail_url": "https://storage.../circuit_123_thumb.png"
  }
}
```

---

### 8. List Circuit Templates

**GET** `/api/v1/circuits/templates`

Query Parameters:
| Parameter | Type | Description |
|-----------|------|-------------|
| `category` | string | Filter: `beginner`, `intermediate`, `advanced` |
| `difficulty` | string | Filter by difficulty |
| `search` | string | Search by name |

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "template-uuid",
      "name": "LED Traffic Light",
      "description": "Create a traffic light with 3 LEDs",
      "category": "beginner",
      "difficulty": "Beginner",
      "thumbnail_url": "https://...",
      "estimated_time_minutes": 15,
      "xp_reward": 25,
      "tags": ["led", "basic", "arduino"],
      "use_count": 1250
    }
  ]
}
```

---

### 9. Use Template (Clone to User)

**POST** `/api/v1/circuits/templates/:id/use`

**Request Body:**
```json
{
  "name": "My Traffic Light (optional)",
  "project_id": "uuid (optional)"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "new-circuit-uuid",
    "name": "LED Traffic Light",
    ...
  },
  "message": "Template applied",
  "xp_earned": 5
}
```

---

### 10. Export Circuit (SPICE/JSON)

**GET** `/api/v1/circuits/:id/export`

Query Parameters:
| Parameter | Type | Description |
|-----------|------|-------------|
| `format` | string | `json`, `spice`, `kicad` |

**Response (JSON):**
```json
{
  "success": true,
  "data": {
    "format": "json",
    "filename": "led_blink_circuit.json",
    "content": "{...}"
  }
}
```

**Response (SPICE):**
```json
{
  "success": true,
  "data": {
    "format": "spice",
    "filename": "led_blink_circuit.cir",
    "content": "* LED Blink Circuit\nV1 1 0 DC 5\nR1 1 2 220\nD1 2 0 LED\n.model LED D(Is=1e-20)...\n.end"
  }
}
```

---

## 🎮 Gamification

### XP Rewards

| Action | XP |
|--------|-----|
| Create first circuit | 15 |
| Save circuit (first time) | 10 |
| Add 5+ components | 5 |
| Add 10+ components | 10 |
| Use a template | 5 |
| Complete simulation | 5 |
| Share circuit publicly | 20 |

---

## 📝 Go Structs

```go
// internal/models/circuit.go

type Circuit struct {
    ID              uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserID          uuid.UUID       `json:"user_id" gorm:"type:uuid;not null"`
    ProjectID       *uuid.UUID      `json:"project_id" gorm:"type:uuid"`
    Name            string          `json:"name" gorm:"size:255;not null"`
    Description     string          `json:"description"`
    ThumbnailURL    string          `json:"thumbnail_url" gorm:"size:500"`
    SchemaData      datatypes.JSON  `json:"schema_data" gorm:"not null;default:'{}'"`
    ComponentsCount int             `json:"components_count" gorm:"default:0"`
    WiresCount      int             `json:"wires_count" gorm:"default:0"`
    IsTemplate      bool            `json:"is_template" gorm:"default:false"`
    IsPublic        bool            `json:"is_public" gorm:"default:false"`
    CreatedAt       time.Time       `json:"created_at"`
    UpdatedAt       time.Time       `json:"updated_at"`
    
    // Relations
    User    *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Project *Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

type CircuitTemplate struct {
    ID                   uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Name                 string          `json:"name" gorm:"size:255;not null"`
    Description          string          `json:"description"`
    Category             string          `json:"category" gorm:"size:100;not null"`
    ThumbnailURL         string          `json:"thumbnail_url" gorm:"size:500"`
    SchemaData           datatypes.JSON  `json:"schema_data" gorm:"not null"`
    Difficulty           string          `json:"difficulty" gorm:"size:50;default:'Beginner'"`
    EstimatedTimeMinutes int             `json:"estimated_time_minutes" gorm:"default:15"`
    XPReward             int             `json:"xp_reward" gorm:"default:10"`
    Tags                 pq.StringArray  `json:"tags" gorm:"type:text[]"`
    UseCount             int             `json:"use_count" gorm:"default:0"`
    CreatedAt            time.Time       `json:"created_at"`
}

// Schema Data Structures (for JSON parsing)
type CircuitSchema struct {
    Components     []CircuitComponent `json:"components"`
    Wires          []Wire             `json:"wires"`
    GroundNodeID   string             `json:"groundNodeId"`
    PowerSourceIDs []string           `json:"powerSourceIds"`
}

type CircuitComponent struct {
    ID         string                 `json:"id"`
    Type       string                 `json:"type"`
    Name       string                 `json:"name"`
    Position   Point                  `json:"position"`
    Rotation   int                    `json:"rotation"`
    Properties map[string]interface{} `json:"properties"`
    Pins       []Pin                  `json:"pins"`
    IsActive   bool                   `json:"isActive"`
    IsBurned   bool                   `json:"isBurned"`
}

type Wire struct {
    ID               string  `json:"id"`
    StartComponentID string  `json:"startComponentId"`
    StartPinID       string  `json:"startPinId"`
    EndComponentID   string  `json:"endComponentId"`
    EndPinID         string  `json:"endPinId"`
    Resistance       float64 `json:"resistance"`
    Current          float64 `json:"current"`
}

type Point struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}

type Pin struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Type        string `json:"type"` // input, output, bidirectional, power, ground
    Position    Point  `json:"position"`
    Voltage     float64 `json:"voltage"`
    Current     float64 `json:"current"`
    IsConnected bool   `json:"isConnected"`
}
```

---

## 🔧 Request/Response DTOs

```go
// internal/dto/circuit_dto.go

type CreateCircuitRequest struct {
    Name        string          `json:"name" validate:"required,max=255"`
    Description string          `json:"description"`
    ProjectID   *uuid.UUID      `json:"project_id"`
    SchemaData  json.RawMessage `json:"schema_data" validate:"required"`
}

type UpdateCircuitRequest struct {
    Name        string          `json:"name" validate:"omitempty,max=255"`
    Description string          `json:"description"`
    SchemaData  json.RawMessage `json:"schema_data"`
}

type CircuitResponse struct {
    ID              uuid.UUID       `json:"id"`
    Name            string          `json:"name"`
    Description     string          `json:"description"`
    ThumbnailURL    string          `json:"thumbnail_url"`
    SchemaData      json.RawMessage `json:"schema_data,omitempty"`
    ComponentsCount int             `json:"components_count"`
    WiresCount      int             `json:"wires_count"`
    ProjectID       *uuid.UUID      `json:"project_id"`
    ProjectName     string          `json:"project_name,omitempty"`
    IsTemplate      bool            `json:"is_template"`
    IsPublic        bool            `json:"is_public"`
    CreatedAt       time.Time       `json:"created_at"`
    UpdatedAt       time.Time       `json:"updated_at"`
}

type CircuitListResponse struct {
    Success bool              `json:"success"`
    Data    []CircuitResponse `json:"data"`
    Meta    PaginationMeta    `json:"meta"`
}
```

---

## 🌐 Environment Variables

```env
# Circuit Settings
MAX_COMPONENTS_FREE=50
MAX_COMPONENTS_PRO=500
MAX_CIRCUITS_FREE=20
MAX_CIRCUITS_PRO=unlimited

# Storage (for thumbnails)
CIRCUIT_THUMBNAILS_BUCKET=circuit-thumbnails
```

---

## ✅ Implementation Checklist

### Backend
- [ ] Create `circuits` table migration
- [ ] Create `circuit_templates` table migration
- [ ] Implement Circuit model
- [ ] Implement CircuitTemplate model
- [ ] Create CircuitRepository
- [ ] Create CircuitService
- [ ] Implement CRUD endpoints
- [ ] Implement template endpoints
- [ ] Implement duplicate endpoint
- [ ] Implement export endpoint (JSON/SPICE)
- [ ] Implement thumbnail upload
- [ ] Add XP rewards logic
- [ ] Seed initial templates

### Frontend (Done ✅)
- [x] Circuit Simulator page
- [x] Component Library
- [x] Custom nodes (React Flow)
- [x] Properties Panel
- [x] Simulation Controls
- [x] Circuit Engine (MNA solver)
- [x] Zustand store
- [ ] Auto-save to backend
- [ ] Load circuit from backend
- [ ] Template browser
- [ ] Export functionality
