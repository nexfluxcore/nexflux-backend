# Remote Virtual Lab API Documentation

This document describes the Remote Virtual Lab system that enables users to access real hardware remotely through live video streaming and code execution.

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Database Schema](#database-schema)
4. [API Endpoints](#api-endpoints)
5. [LiveKit Integration](#livekit-integration)
6. [Lab Hardware Setup](#lab-hardware-setup)
7. [Code Execution Flow](#code-execution-flow)
8. [Booking & Bidding System](#booking--bidding-system)
9. [XP & Gamification](#xp--gamification)
10. [WebSocket Events](#websocket-events)
11. [Environment Variables](#environment-variables)
12. [Security Considerations](#security-considerations)

---

## Overview

The Remote Virtual Lab allows users to:
- Book and access real IoT hardware remotely
- Write and upload code to microcontrollers
- Watch live video stream of the hardware
- Control sensors and actuators in real-time
- Complete challenges and earn XP

### Key Components:
- **Lab Stations**: Physical mini PCs connected to microcontrollers and sensors
- **LiveKit Server**: Real-time video streaming
- **Code Editor**: Monaco Editor for writing code
- **Execution Engine**: Compiles and uploads code to hardware

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              USER (Browser)                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Video Stream│  │ Code Editor │  │  Controls   │  │   Lab Status/Chat   │ │
│  │  (LiveKit)  │  │  (Monaco)   │  │  (Buttons)  │  │    (WebSocket)      │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘ │
└─────────┼────────────────┼────────────────┼────────────────────┼────────────┘
          │                │                │                    │
          ▼                ▼                ▼                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           BACKEND (Go/Fiber)                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌─────────────────┐  │
│  │ LiveKit API  │  │  Code Exec   │  │  Lab Manager │  │  WebSocket Hub  │  │
│  │   Gateway    │  │    Engine    │  │   Service    │  │    (Events)     │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └────────┬────────┘  │
└─────────┼────────────────┼────────────────┼─────────────────────┼───────────┘
          │                │                │                     │
          ▼                ▼                ▼                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          LAB STATION (Mini PC)                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌─────────────────┐  │
│  │   Camera     │  │   Arduino/   │  │   Sensors    │  │   Agent App     │  │
│  │  (USB/IP)    │  │    ESP32     │  │  & Actuators │  │  (Go/Python)    │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Database Schema

### Table: `labs`
```sql
CREATE TABLE labs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    platform VARCHAR(50) NOT NULL, -- 'arduino_uno', 'esp32', 'raspberry_pi', etc.
    status VARCHAR(20) DEFAULT 'available', -- 'available', 'busy', 'maintenance', 'offline'
    hardware_specs JSONB, -- { "camera": "USB 1080p", "sensors": ["DHT22", "LDR"], "actuators": ["LED", "Servo"] }
    thumbnail_url TEXT,
    livekit_room_name VARCHAR(100),
    current_user_id UUID REFERENCES users(id),
    session_started_at TIMESTAMPTZ,
    max_session_duration INTEGER DEFAULT 30, -- minutes
    queue_count INTEGER DEFAULT 0,
    total_sessions INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Table: `lab_sessions`
```sql
CREATE TABLE lab_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lab_id UUID REFERENCES labs(id) NOT NULL,
    user_id UUID REFERENCES users(id) NOT NULL,
    project_id UUID REFERENCES projects(id),
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'completed', 'expired', 'cancelled'
    started_at TIMESTAMPTZ DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    duration_seconds INTEGER,
    code_submitted TEXT,
    compilation_status VARCHAR(20), -- 'pending', 'compiling', 'success', 'failed'
    compilation_output TEXT,
    xp_earned INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Table: `lab_bookings`
```sql
CREATE TABLE lab_bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lab_id UUID REFERENCES labs(id) NOT NULL,
    user_id UUID REFERENCES users(id) NOT NULL,
    scheduled_at TIMESTAMPTZ NOT NULL,
    duration_minutes INTEGER DEFAULT 30,
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'confirmed', 'active', 'completed', 'cancelled', 'no_show'
    bid_amount INTEGER DEFAULT 0, -- XP bid for priority
    queue_position INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(lab_id, scheduled_at)
);
```

### Table: `lab_queue`
```sql
CREATE TABLE lab_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lab_id UUID REFERENCES labs(id) NOT NULL,
    user_id UUID REFERENCES users(id) NOT NULL,
    position INTEGER NOT NULL,
    bid_amount INTEGER DEFAULT 0,
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    UNIQUE(lab_id, user_id)
);
```

### Table: `lab_hardware_logs`
```sql
CREATE TABLE lab_hardware_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lab_id UUID REFERENCES labs(id) NOT NULL,
    session_id UUID REFERENCES lab_sessions(id),
    event_type VARCHAR(50) NOT NULL, -- 'sensor_read', 'actuator_control', 'code_upload', 'error'
    event_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

## API Endpoints

### Lab Management

#### List Available Labs
```
GET /api/v1/labs
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `platform` | string | Filter by platform (arduino, esp32, etc.) |
| `status` | string | Filter by status (available, busy) |
| `page` | integer | Page number |
| `limit` | integer | Items per page |

**Response:**
```json
{
    "success": true,
    "data": {
        "labs": [
            {
                "id": "uuid",
                "name": "Arduino Lab #1",
                "slug": "arduino-lab-1",
                "platform": "arduino_uno",
                "status": "available",
                "hardware_specs": {
                    "camera": "USB 1080p",
                    "sensors": ["DHT22", "LDR", "Ultrasonic"],
                    "actuators": ["LED RGB", "Servo", "Buzzer"]
                },
                "thumbnail_url": "https://...",
                "queue_count": 0,
                "current_user": null
            },
            {
                "id": "uuid",
                "name": "ESP32 Lab #1",
                "slug": "esp32-lab-1",
                "platform": "esp32",
                "status": "busy",
                "queue_count": 3,
                "current_user": {
                    "id": "uuid",
                    "name": "John Doe",
                    "session_remaining": 1200 // seconds
                }
            }
        ],
        "total": 10,
        "page": 1,
        "limit": 10
    }
}
```

#### Get Lab Details
```
GET /api/v1/labs/:id
```

**Response:**
```json
{
    "success": true,
    "data": {
        "id": "uuid",
        "name": "Arduino Lab #1",
        "slug": "arduino-lab-1",
        "description": "Complete Arduino Uno setup with various sensors",
        "platform": "arduino_uno",
        "status": "available",
        "hardware_specs": {
            "board": "Arduino Uno R3",
            "camera": "Logitech C920 1080p",
            "sensors": [
                { "name": "DHT22", "type": "temperature_humidity", "pin": "D2" },
                { "name": "LDR", "type": "light", "pin": "A0" },
                { "name": "HC-SR04", "type": "ultrasonic", "pins": { "trig": "D3", "echo": "D4" } }
            ],
            "actuators": [
                { "name": "RGB LED", "type": "led", "pins": { "r": "D9", "g": "D10", "b": "D11" } },
                { "name": "Servo Motor", "type": "servo", "pin": "D5" },
                { "name": "Buzzer", "type": "buzzer", "pin": "D6" }
            ]
        },
        "thumbnail_url": "https://...",
        "max_session_duration": 30,
        "total_sessions": 1250,
        "queue": [
            { "position": 1, "user": { "name": "Alice" }, "bid_amount": 50 },
            { "position": 2, "user": { "name": "Bob" }, "bid_amount": 30 }
        ]
    }
}
```

---

### Booking & Queue

#### Join Lab Queue
```
POST /api/v1/labs/:id/queue
```

**Request Body:**
```json
{
    "bid_amount": 50 // XP to bid for priority (optional)
}
```

**Response:**
```json
{
    "success": true,
    "data": {
        "queue_id": "uuid",
        "position": 3,
        "estimated_wait": 1800, // seconds
        "expires_at": "2024-01-15T10:30:00Z"
    }
}
```

#### Leave Queue
```
DELETE /api/v1/labs/:id/queue
```

#### Start Lab Session
```
POST /api/v1/labs/:id/session/start
```

**Response:**
```json
{
    "success": true,
    "data": {
        "session_id": "uuid",
        "livekit_token": "eyJ...", // Token for video stream
        "livekit_url": "wss://livekit.nexflux.io",
        "room_name": "lab-arduino-1-session-xyz",
        "expires_at": "2024-01-15T11:00:00Z",
        "hardware_config": {
            "board": "arduino_uno",
            "serial_port": "/dev/ttyUSB0",
            "baud_rate": 9600
        }
    }
}
```

#### End Lab Session
```
POST /api/v1/labs/:id/session/end
```

**Request Body:**
```json
{
    "feedback": "Great experience!",
    "rating": 5
}
```

---

### Code Execution

#### Submit Code
```
POST /api/v1/labs/:id/code/submit
```

**Request Body:**
```json
{
    "session_id": "uuid",
    "code": "void setup() { pinMode(13, OUTPUT); } void loop() { digitalWrite(13, HIGH); delay(1000); }",
    "language": "arduino", // 'arduino', 'micropython', 'c'
    "filename": "blink.ino"
}
```

**Response:**
```json
{
    "success": true,
    "data": {
        "compilation_id": "uuid",
        "status": "compiling"
    }
}
```

#### Get Compilation Status
```
GET /api/v1/labs/:id/code/status/:compilation_id
```

**Response:**
```json
{
    "success": true,
    "data": {
        "status": "success", // 'pending', 'compiling', 'uploading', 'success', 'failed'
        "output": "Sketch uses 924 bytes (2%) of program storage space...",
        "errors": null,
        "uploaded_at": "2024-01-15T10:35:00Z"
    }
}
```

#### Read Sensor Data
```
GET /api/v1/labs/:id/sensors
```

**Response:**
```json
{
    "success": true,
    "data": {
        "timestamp": "2024-01-15T10:35:00Z",
        "sensors": {
            "DHT22": { "temperature": 25.5, "humidity": 60 },
            "LDR": { "light_level": 512 },
            "HC-SR04": { "distance_cm": 45.2 }
        }
    }
}
```

#### Control Actuator
```
POST /api/v1/labs/:id/actuators/control
```

**Request Body:**
```json
{
    "actuator": "RGB LED",
    "action": "set_color",
    "params": { "r": 255, "g": 0, "b": 128 }
}
```

---

## LiveKit Integration

### Server Configuration

```env
LIVEKIT_URL=wss://livekit.nexflux.io
LIVEKIT_API_KEY=your-api-key
LIVEKIT_API_SECRET=your-api-secret
```

### Token Generation (Backend)

```go
import (
    "github.com/livekit/server-sdk-go"
    "github.com/livekit/protocol/auth"
)

func generateLabToken(userID, roomName string) (string, error) {
    at := auth.NewAccessToken(apiKey, apiSecret)
    
    grant := &auth.VideoGrant{
        Room:     roomName,
        RoomJoin: true,
    }
    
    at.AddGrant(grant).
        SetIdentity(userID).
        SetValidFor(time.Hour)
    
    return at.ToJWT()
}
```

### Room Setup for Lab

Each lab station creates a room with:
- **Camera Track**: Video from USB camera
- **Data Channel**: For sensor data streaming
- **Screen Share**: Optional for circuit diagram overlay

### Lab Agent (Running on Mini PC)

```python
# lab_agent.py
from livekit import api, rtc
import cv2
import serial
import asyncio

class LabAgent:
    def __init__(self, lab_id, livekit_url, api_key, api_secret):
        self.lab_id = lab_id
        self.room = None
        self.camera = cv2.VideoCapture(0)
        self.serial = serial.Serial('/dev/ttyUSB0', 9600)
    
    async def connect(self):
        # Generate token and join room
        token = generate_token(self.lab_id)
        self.room = await rtc.Room.connect(livekit_url, token)
        
        # Publish camera track
        video_track = await self.create_video_track()
        await self.room.local_participant.publish_track(video_track)
        
        # Start sensor data loop
        asyncio.create_task(self.sensor_loop())
    
    async def sensor_loop(self):
        while True:
            data = self.read_sensors()
            # Send via data channel
            await self.room.local_participant.publish_data(
                json.dumps(data).encode(),
                reliable=True
            )
            await asyncio.sleep(0.5)
    
    async def handle_code_upload(self, code: str):
        # Compile and upload to Arduino
        result = await compile_and_upload(code)
        return result
```

---

## Booking & Bidding System

### Queue Priority Algorithm

```
Priority Score = Base Position + (Bid Amount * 0.1) + (User Level * 0.05)
```

- Users with higher bids get priority
- Premium users get slight priority boost
- Queue expires after 30 minutes if not claimed

### XP Bidding Rules:

1. Minimum bid: 0 XP (first-come-first-served)
2. Maximum bid: User's available XP
3. XP is deducted when session starts
4. XP refunded if session cancelled by system

---

## XP & Gamification

### XP Rewards:

| Action | XP Reward |
|--------|-----------|
| Complete lab session | +50 XP |
| First successful code upload | +20 XP |
| Read all sensors successfully | +10 XP |
| Control all actuators | +15 XP |
| Complete challenge objective | +100 XP |
| 5-star rating from session | +25 XP |

### Achievements:

- **First Circuit**: Complete first lab session
- **Code Master**: Upload 50 successful programs
- **Sensor Guru**: Read 100 sensor values
- **Night Owl**: Complete session between 00:00-06:00
- **Speed Demon**: Complete objective in under 5 minutes

---

## WebSocket Events

### Connection
```
ws://api.nexflux.io/ws/labs/:lab_id
```

### Events (Server → Client)

```javascript
// Lab status changed
{ "event": "lab_status", "data": { "status": "busy", "current_user": {...} } }

// Queue updated
{ "event": "queue_update", "data": { "position": 2, "estimated_wait": 900 } }

// Sensor data (real-time)
{ "event": "sensor_data", "data": { "DHT22": { "temp": 25.5 }, ... } }

// Code compilation result
{ "event": "compilation", "data": { "status": "success", "output": "..." } }

// Session time warning
{ "event": "time_warning", "data": { "remaining_seconds": 300 } }

// Session ended
{ "event": "session_ended", "data": { "reason": "timeout", "xp_earned": 75 } }
```

### Events (Client → Server)

```javascript
// Control actuator
{ "event": "actuator_control", "data": { "actuator": "LED", "action": "on" } }

// Send serial command
{ "event": "serial_command", "data": { "command": "GET_TEMP" } }

// Request sensor read
{ "event": "read_sensors", "data": {} }
```

---

## Environment Variables

### Backend
```env
# LiveKit
LIVEKIT_URL=wss://livekit.nexflux.io
LIVEKIT_API_KEY=your-api-key
LIVEKIT_API_SECRET=your-api-secret

# Lab Configuration
LAB_SESSION_DURATION=1800 # 30 minutes in seconds
LAB_QUEUE_EXPIRY=1800 # 30 minutes
LAB_MAX_BID_MULTIPLIER=2 # Max bid = user_level * multiplier

# Code Execution
ARDUINO_CLI_PATH=/usr/local/bin/arduino-cli
PLATFORMIO_PATH=/usr/local/bin/platformio
COMPILE_TIMEOUT=60 # seconds
UPLOAD_TIMEOUT=30 # seconds
```

### Lab Agent (Mini PC)
```env
LAB_ID=arduino-lab-1
BACKEND_URL=https://api.nexflux.io
LIVEKIT_URL=wss://livekit.nexflux.io
SERIAL_PORT=/dev/ttyUSB0
BAUD_RATE=9600
CAMERA_DEVICE=0
```

---

## Security Considerations

1. **Session Isolation**: Each session runs in isolated environment
2. **Code Sandboxing**: Uploaded code is sandboxed and validated
3. **Rate Limiting**: Limit code uploads (5 per minute)
4. **Timeout Protection**: Sessions auto-terminate after max duration
5. **Input Validation**: Sanitize all serial commands
6. **Access Control**: Only session owner can control hardware
7. **Audit Logging**: Log all hardware interactions
8. **Network Isolation**: Lab PCs on separate VLAN

---

## Frontend Integration

### Required Packages
```bash
npm install @livekit/components-react @livekit/components-styles livekit-client
npm install @monaco-editor/react
npm install lucide-react
```

### Component Structure
```
src/pages/
├── RemoteLab/
│   ├── index.tsx          # Main lab listing page
│   ├── LabSession.tsx     # Active lab session page
│   ├── components/
│   │   ├── LabCard.tsx
│   │   ├── VideoStream.tsx
│   │   ├── CodeEditor.tsx
│   │   ├── SensorPanel.tsx
│   │   ├── ActuatorControls.tsx
│   │   ├── QueueStatus.tsx
│   │   └── SessionTimer.tsx
│   └── hooks/
│       ├── useLabSession.ts
│       ├── useSensorData.ts
│       └── useCodeExecution.ts
```

---

## Recommended Code Editor: Monaco Editor

**Why Monaco Editor:**
- Same editor as VS Code (familiar to developers)
- Excellent syntax highlighting for Arduino/C/C++/Python
- IntelliSense support
- Lightweight and fast
- Easy React integration

**Installation:**
```bash
npm install @monaco-editor/react
```

**Basic Usage:**
```tsx
import Editor from '@monaco-editor/react';

<Editor
    height="400px"
    defaultLanguage="cpp"
    theme={isDark ? 'vs-dark' : 'light'}
    value={code}
    onChange={(value) => setCode(value)}
    options={{
        minimap: { enabled: false },
        fontSize: 14,
        automaticLayout: true
    }}
/>
```
