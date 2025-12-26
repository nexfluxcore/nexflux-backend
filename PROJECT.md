# NexFlux Project Progress & XP System Documentation

Dokumen ini menjelaskan **parameter keberhasilan project**, cara meningkatkan progress, dan sistem XP gamifikasi yang perlu diimplementasikan di backend.

---

## 📋 Table of Contents

1. [Overview Sistem Progress](#overview-sistem-progress)
2. [Parameter Keberhasilan Project](#parameter-keberhasilan-project)
3. [Cara Meningkatkan Progress](#cara-meningkatkan-progress)
4. [Sistem XP & Rewards](#sistem-xp--rewards)
5. [Backend API Requirements](#backend-api-requirements)
6. [Database Schema Updates](#database-schema-updates)
7. [Flow Diagram](#flow-diagram)

---

## Overview Sistem Progress

NexFlux menggunakan sistem gamifikasi untuk memotivasi pengguna menyelesaikan project IoT mereka. Progress project dihitung berdasarkan **completion rate** dari beberapa komponen utama:

### Komponen Utama Project:
1. **Schema/Circuit Editor** - Desain rangkaian elektronik
2. **Code Editor** - Penulisan firmware/code
3. **Simulation** - Pengujian virtual rangkaian

---

## Parameter Keberhasilan Project

Progress project dihitung berdasarkan weighted completion dari komponen berikut:

| Komponen | Weight | Kondisi Complete |
|----------|--------|------------------|
| **Schema Data** | 30% | Ada data rangkaian (components > 0) |
| **Code Data** | 35% | Ada kode yang ditulis (characters > 100) |
| **Simulation** | 25% | Simulasi berhasil dijalankan (success = true) |
| **Verification** | 10% | Project di-verify/reviewed |

### Formula Perhitungan Progress:
```
progress = (schema_weight * schema_complete) + 
           (code_weight * code_complete) + 
           (simulation_weight * simulation_complete) + 
           (verification_weight * verification_complete)

Dimana:
- schema_complete = 1 jika schema_data.components.length > 0, else 0
- code_complete = 1 jika code_data.content.length > 100, else min(code_length/100, 1)
- simulation_complete = simulation_data.last_run_success ? 1 : 0
- verification_complete = is_verified ? 1 : 0
```

### Progress Breakdown Detail:

| Progress Range | Status | Kondisi |
|---------------|--------|---------|
| 0% | Not Started | Project baru dibuat, belum ada activity |
| 1-29% | In Progress | Ada salah satu: schema/code/simulation dimulai |
| 30-59% | Developing | 2 dari 3 komponen sudah ada data |
| 60-89% | Testing | Schema + Code complete, sedang simulasi |
| 90-99% | Almost Done | Semua complete, menunggu verification |
| 100% | Completed | Semua komponen terisi + verified |

---

## Cara Meningkatkan Progress

### 1. **Circuit Editor (Schema)** - +30%
Untuk mendapatkan progress dari schema:
- Masuk ke `/studio?project={id}`
- Tambahkan minimal **1 komponen** ke canvas
- Simpan rangkaian

**Kondisi lengkap:**
```json
{
  "schema_data": {
    "components": [
      { "id": "comp1", "type": "arduino_uno", "position": {...} }
    ],
    "connections": [...]
  }
}
```

### 2. **Code Editor** - +35%
Untuk mendapatkan progress dari code:
- Masuk ke `/studio?project={id}&mode=code`
- Tulis kode minimal **100 karakter**
- Simpan kode

**Progress partial:**
- 0-49 chars = 0%
- 50-99 chars = 50% dari weight
- 100+ chars = 100% dari weight

**Kondisi lengkap:**
```json
{
  "code_data": {
    "content": "void setup() { ... } void loop() { ... }",
    "language": "arduino",
    "last_saved": "2024-01-15T10:00:00Z"
  }
}
```

### 3. **Simulation** - +25%
Untuk mendapatkan progress dari simulasi:
- Masuk ke `/simulations?project={id}`
- Jalankan simulasi
- Simulasi harus **SUCCESS** (tidak ada error)

**Kondisi lengkap:**
```json
{
  "simulation_data": {
    "last_run_at": "2024-01-15T10:00:00Z",
    "last_run_success": true,
    "results": {...}
  }
}
```

### 4. **Verification (Bonus)** - +10%
- Automatic verification setelah semua komponen complete
- Atau manual review dari admin/mentor

---

## Sistem XP & Rewards

### XP dari Project:

| Action | XP Reward | Kondisi |
|--------|-----------|---------|
| Create Project | +10 XP | Project baru dibuat |
| First Schema Save | +15 XP | Pertama kali simpan schema |
| First Code Save | +20 XP | Pertama kali simpan code |
| First Simulation Run | +10 XP | Pertama kali jalankan simulasi |
| Simulation Success | +25 XP | Simulasi berhasil tanpa error |
| Complete Project (100%) | +50 XP | Project selesai |
| First Complete | +100 XP | Bonus pertama kali complete |

### Total XP Potential Per Project:
- Minimum: **50 XP** (hanya complete)
- Maximum: **230 XP** (semua milestone + first complete)

### XP dari Challenge:

| Challenge Action | XP Reward |
|-----------------|-----------|
| Start Challenge | +5 XP |
| Complete Easy | +50 XP |
| Complete Medium | +100 XP |
| Complete Hard | +200 XP |
| Complete Expert | +500 XP |
| Daily Challenge Bonus | 1.5x - 2x multiplier |

### XP dari Remote Lab:

| Lab Action | XP Reward |
|------------|-----------|
| Complete lab session | +50 XP |
| First successful code upload | +20 XP |
| Read all sensors successfully | +10 XP |
| Control all actuators | +15 XP |
| Complete challenge objective | +100 XP |
| 5-star rating from session | +25 XP |

---

## Backend API Requirements

### Endpoints yang Perlu Diimplementasi:

#### 1. Update Project Progress
```
PUT /api/v1/projects/:id/progress
```

**Request Body:**
```json
{
  "component": "schema|code|simulation|verification",
  "data": {...},
  "action": "save|run|verify"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "progress": 65,
    "progress_change": +30,
    "xp_earned": 15,
    "milestones_unlocked": ["first_schema_save"]
  }
}
```

#### 2. Get Project Progress Detail
```
GET /api/v1/projects/:id/progress
```

**Response:**
```json
{
  "success": true,
  "data": {
    "progress": 65,
    "breakdown": {
      "schema": { "complete": true, "weight": 30, "earned": 30 },
      "code": { "complete": true, "weight": 35, "earned": 35 },
      "simulation": { "complete": false, "weight": 25, "earned": 0 },
      "verification": { "complete": false, "weight": 10, "earned": 0 }
    },
    "milestones": [
      { "id": "first_schema_save", "unlocked_at": "...", "xp_earned": 15 },
      { "id": "first_code_save", "unlocked_at": "...", "xp_earned": 20 }
    ],
    "next_action": "Run simulation to earn +25 XP"
  }
}
```

#### 3. Complete Project
```
POST /api/v1/projects/:id/complete
```

**Response:**
```json
{
  "success": true,
  "data": {
    "completed_at": "2024-01-15T10:00:00Z",
    "xp_earned": 150,
    "total_xp": 230,
    "achievements_unlocked": ["first_complete", "circuit_master"],
    "level_up": {
      "old_level": 5,
      "new_level": 6
    }
  }
}
```

#### 4. Save Schema Data
```
PUT /api/v1/projects/:id/schema
```

**Request Body:**
```json
{
  "components": [...],
  "connections": [...],
  "canvas_settings": {...}
}
```

#### 5. Save Code Data
```
PUT /api/v1/projects/:id/code
```

**Request Body:**
```json
{
  "content": "void setup() {...}",
  "language": "arduino",
  "filename": "main.ino"
}
```

#### 6. Run Simulation
```
POST /api/v1/projects/:id/simulate
```

**Request Body:**
```json
{
  "duration_ms": 5000,
  "speed_multiplier": 1
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "simulation_id": "uuid",
    "status": "success|failed|timeout",
    "results": {
      "output_data": [...],
      "errors": [],
      "warnings": []
    },
    "xp_earned": 25,
    "progress_update": { "old": 65, "new": 90 }
  }
}
```

---

## Database Schema Updates

### Table: `project_progress` (New)
```sql
CREATE TABLE project_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    component VARCHAR(50) NOT NULL, -- 'schema', 'code', 'simulation', 'verification'
    is_complete BOOLEAN DEFAULT false,
    completion_percentage INTEGER DEFAULT 0, -- 0-100
    completed_at TIMESTAMPTZ,
    data JSONB, -- component-specific data
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, component)
);
```

### Table: `project_milestones` (New)
```sql
CREATE TABLE project_milestones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    milestone_type VARCHAR(100) NOT NULL, -- 'first_schema_save', 'first_code_save', etc.
    xp_earned INTEGER DEFAULT 0,
    unlocked_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, milestone_type)
);
```

### Table: `user_xp_transactions` (New)
```sql
CREATE TABLE user_xp_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    xp_amount INTEGER NOT NULL,
    xp_type VARCHAR(50) NOT NULL, -- 'project', 'challenge', 'lab', 'achievement'
    source_id UUID, -- project_id, challenge_id, etc.
    source_type VARCHAR(50), -- 'project', 'challenge', 'lab_session'
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Update Table: `projects`
```sql
ALTER TABLE projects ADD COLUMN IF NOT EXISTS schema_data JSONB DEFAULT '{}';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS code_data JSONB DEFAULT '{}';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS simulation_data JSONB DEFAULT '{}';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS is_verified BOOLEAN DEFAULT false;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS verified_at TIMESTAMPTZ;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS verified_by UUID REFERENCES users(id);
```

### Update Table: `users`
```sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS current_xp INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS total_xp INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS level INTEGER DEFAULT 1;
ALTER TABLE users ADD COLUMN IF NOT EXISTS xp_to_next_level INTEGER DEFAULT 100;
```

---

## Flow Diagram

### Project Progress Flow:
```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           USER CREATES PROJECT                                   │
│                               progress = 0%                                      │
│                              xp_earned = +10                                     │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                    ┌───────────────────┼───────────────────┐
                    ▼                   ▼                   ▼
            ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
            │ CIRCUIT      │    │ CODE EDITOR  │    │ SIMULATION   │
            │ EDITOR       │    │              │    │              │
            │              │    │              │    │              │
            │ +30% weight  │    │ +35% weight  │    │ +25% weight  │
            │ +15 XP first │    │ +20 XP first │    │ +25 XP success│
            └──────────────┘    └──────────────┘    └──────────────┘
                    │                   │                   │
                    └───────────────────┼───────────────────┘
                                        │
                                        ▼
                    ┌────────────────────────────────────────────┐
                    │         CALCULATE TOTAL PROGRESS           │
                    │    progress = sum(component_progress)      │
                    └────────────────────────────────────────────┘
                                        │
                            ┌───────────┴───────────┐
                            │                       │
                      progress < 100%         progress = 100%
                            │                       │
                            ▼                       ▼
                    ┌──────────────┐        ┌──────────────┐
                    │   IN PROGRESS │        │   COMPLETE!  │
                    │   Keep working│        │   +50 XP     │
                    │   on remaining│        │   (bonus if  │
                    │   components  │        │    first)    │
                    └──────────────┘        └──────────────┘
```

### XP Calculation Flow:
```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                               XP CALCULATION                                     │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
                    ┌────────────────────────────────────────────┐
                    │   1. Check if milestone already unlocked    │
                    │   2. Calculate XP based on action           │
                    │   3. Apply multipliers (daily, streak, etc) │
                    └────────────────────────────────────────────┘
                                        │
                                        ▼
                    ┌────────────────────────────────────────────┐
                    │        UPDATE USER XP & LEVEL              │
                    │                                            │
                    │   current_xp += earned_xp                  │
                    │   total_xp += earned_xp                    │
                    │   IF current_xp >= xp_to_next_level:       │
                    │       level++                              │
                    │       current_xp -= xp_to_next_level       │
                    │       xp_to_next_level = next_level_xp()   │
                    └────────────────────────────────────────────┘
                                        │
                                        ▼
                    ┌────────────────────────────────────────────┐
                    │         CHECK ACHIEVEMENTS                  │
                    │   - First Complete                          │
                    │   - 10 Projects                            │
                    │   - Master Coder, etc.                     │
                    └────────────────────────────────────────────┘
```

---

## Testing Checklist

Untuk test progress project dan XP system:

### ✅ Test 1: Create Project
- [ ] Buat project baru
- [ ] Verify: progress = 0%, xp_earned = +10

### ✅ Test 2: Circuit Editor
- [ ] Buka `/studio?project={id}`
- [ ] Tambahkan 1+ komponen
- [ ] Simpan
- [ ] Verify: progress +30%, xp_earned = +15 (first save)

### ✅ Test 3: Code Editor
- [ ] Buka `/studio?project={id}&mode=code`
- [ ] Tulis kode 100+ karakter
- [ ] Simpan
- [ ] Verify: progress +35%, xp_earned = +20 (first save)

### ✅ Test 4: Simulation
- [ ] Buka `/simulations?project={id}`
- [ ] Jalankan simulasi
- [ ] Verify: Jika success, progress +25%, xp_earned = +25

### ✅ Test 5: Complete Project
- [ ] Semua komponen complete
- [ ] Verify: progress = 100%, xp_earned = +50 (completion bonus)

### ✅ Test 6: Level Up
- [ ] Kumpulkan XP sampai melebihi threshold
- [ ] Verify: level naik, current_xp reset

---

## Notes untuk Backend Developer

1. **Milestone hanya berlaku sekali** - Jangan berikan XP duplicate untuk milestone yang sama
2. **Progress harus recalculate** - Setiap update komponen harus recalculate total progress
3. **XP Transaction Log** - Semua XP harus dicatat untuk audit trail
4. **Real-time updates** - Gunakan WebSocket untuk notify frontend saat XP/progress berubah
5. **Untuk Remote Lab** - XP diberikan berdasarkan aksi di lab session (lihat LAB.md)
6. **Daily multiplier** - XP bisa dikalikan untuk daily challenges

---

## Environment Variables

```env
# XP Configuration
XP_PROJECT_CREATE=10
XP_FIRST_SCHEMA_SAVE=15
XP_FIRST_CODE_SAVE=20
XP_FIRST_SIMULATION_RUN=10
XP_SIMULATION_SUCCESS=25
XP_PROJECT_COMPLETE=50
XP_FIRST_PROJECT_COMPLETE_BONUS=100

# Level Configuration
XP_LEVEL_BASE=100
XP_LEVEL_MULTIPLIER=1.5

# Progress Weights
PROGRESS_WEIGHT_SCHEMA=30
PROGRESS_WEIGHT_CODE=35
PROGRESS_WEIGHT_SIMULATION=25
PROGRESS_WEIGHT_VERIFICATION=10
```

---

## Related Documentation

- [LAB.md](./LAB.md) - Remote Virtual Lab documentation
- [Auth.md](./Auth.md) - Authentication & user management
- [ERD.md](./ERD.md) - Complete database schema
