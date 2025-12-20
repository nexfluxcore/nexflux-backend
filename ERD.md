# NexFlux Backend - Database ERD & API Documentation

> Dokumentasi struktur database dan API untuk backend NexFlux (Golang)  
> Generated: 2024-12-20  
> Version: 1.0.0

---

## Table of Contents

1. [Entity Relationship Diagram](#entity-relationship-diagram)
2. [Database Tables](#database-tables)
   - [Users Module](#users-module)
   - [Projects Module](#projects-module)
   - [Components Module](#components-module)
   - [Challenges Module](#challenges-module)
   - [Notifications Module](#notifications-module)
   - [Gamification Module](#gamification-module)
3. [API Endpoints](#api-endpoints)
4. [Request/Response Payloads](#requestresponse-payloads)

---

## Entity Relationship Diagram

```
┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│      users       │       │     projects     │       │    components    │
├──────────────────┤       ├──────────────────┤       ├──────────────────┤
│ id (PK)          │◄──┐   │ id (PK)          │       │ id (PK)          │
│ email            │   │   │ user_id (FK)     │───────│ category_id (FK) │
│ password_hash    │   │   │ name             │       │ name             │
│ name             │   │   │ description      │       │ description      │
│ role             │   │   │ thumbnail_url    │       │ specs (JSON)     │
│ avatar_url       │   │   │ difficulty       │       │ price            │
│ bio              │   │   │ progress         │       │ stock            │
│ level            │   │   │ xp_reward        │       │ rating           │
│ total_xp         │   │   │ is_public        │       │ image_url        │
│ streak_days      │   │   │ is_favorite      │       │ created_at       │
│ language         │   │   │ created_at       │       │ updated_at       │
│ theme            │   │   │ updated_at       │       └────────┬─────────┘
│ created_at       │   │   └────────┬─────────┘                │
│ updated_at       │   │            │                          │
└────────┬─────────┘   │            │         ┌────────────────┘
         │             │            │         │
         │             │            ▼         ▼
         │             │   ┌──────────────────────┐
         │             │   │  project_components  │
         │             │   ├──────────────────────┤
         │             │   │ id (PK)              │
         │             │   │ project_id (FK)      │
         │             │   │ component_id (FK)    │
         │             │   │ quantity             │
         │             │   └──────────────────────┘
         │             │
         │             │   ┌──────────────────┐       ┌──────────────────┐
         │             │   │    challenges    │       │challenge_progress│
         │             │   ├──────────────────┤       ├──────────────────┤
         │             └──►│ id (PK)          │◄──────│ id (PK)          │
         │                 │ title            │       │ user_id (FK)     │
         │                 │ description      │       │ challenge_id (FK)│
         │                 │ difficulty       │       │ progress         │
         │                 │ xp_reward        │       │ status           │
         │                 │ category         │       │ started_at       │
         │                 │ type             │       │ completed_at     │
         │                 │ time_limit       │       └──────────────────┘
         │                 │ participants     │
         │                 │ is_active        │
         │                 │ created_at       │
         │                 └──────────────────┘
         │
         │   ┌──────────────────┐       ┌──────────────────┐
         │   │   notifications  │       │   achievements   │
         │   ├──────────────────┤       ├──────────────────┤
         └──►│ id (PK)          │       │ id (PK)          │
             │ user_id (FK)     │       │ name             │
             │ type             │       │ description      │
             │ title            │       │ icon             │
             │ message          │       │ rarity           │
             │ data (JSON)      │       │ requirement_type │
             │ is_read          │       │ requirement_value│
             │ created_at       │       │ xp_reward        │
             └──────────────────┘       └────────┬─────────┘
                                                 │
                                                 ▼
                                        ┌──────────────────┐
                                        │user_achievements │
                                        ├──────────────────┤
                                        │ id (PK)          │
                                        │ user_id (FK)     │
                                        │ achievement_id(FK│
                                        │ unlocked_at      │
                                        └──────────────────┘
```

---

## Database Tables

### Users Module

#### Table: `users`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Unique identifier |
| email | VARCHAR(255) | UNIQUE, NOT NULL | User email |
| password_hash | VARCHAR(255) | NOT NULL | Bcrypt hashed password |
| name | VARCHAR(100) | NOT NULL | Display name |
| username | VARCHAR(50) | UNIQUE | Username (@handle) |
| role | ENUM | DEFAULT 'student' | 'student', 'educator', 'maker', 'institution', 'admin' |
| avatar_url | VARCHAR(500) | | Profile picture URL |
| bio | TEXT | | User biography |
| level | INT | DEFAULT 1 | Current level |
| total_xp | INT | DEFAULT 0 | Total XP earned |
| current_xp | INT | DEFAULT 0 | XP progress to next level |
| target_xp | INT | DEFAULT 1000 | XP needed for next level |
| streak_days | INT | DEFAULT 0 | Current streak count |
| last_active_at | TIMESTAMP | | Last activity timestamp |
| language | VARCHAR(5) | DEFAULT 'en' | Preferred language: 'en', 'id', 'jp' |
| theme | VARCHAR(10) | DEFAULT 'system' | Theme preference: 'light', 'dark', 'system' |
| is_pro | BOOLEAN | DEFAULT false | Pro subscription status |
| email_verified | BOOLEAN | DEFAULT false | Email verification status |
| created_at | TIMESTAMP | DEFAULT NOW() | Account creation date |
| updated_at | TIMESTAMP | | Last update timestamp |

#### Table: `user_settings`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Unique identifier |
| user_id | UUID | FOREIGN KEY | Reference to users |
| notification_email | BOOLEAN | DEFAULT true | Email notifications |
| notification_push | BOOLEAN | DEFAULT true | Push notifications |
| notification_marketing | BOOLEAN | DEFAULT false | Marketing emails |
| notification_updates | BOOLEAN | DEFAULT true | Product updates |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | | |

#### Table: `user_sessions`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Session ID |
| user_id | UUID | FOREIGN KEY | Reference to users |
| token | VARCHAR(500) | NOT NULL | JWT token |
| device_info | VARCHAR(255) | | Device/browser info |
| ip_address | VARCHAR(45) | | IP address |
| expires_at | TIMESTAMP | NOT NULL | Token expiration |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### Projects Module

#### Table: `projects`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Unique identifier |
| user_id | UUID | FOREIGN KEY, NOT NULL | Project owner |
| name | VARCHAR(200) | NOT NULL | Project name |
| description | TEXT | | Project description |
| thumbnail_url | VARCHAR(500) | | Thumbnail image URL |
| difficulty | ENUM | DEFAULT 'Beginner' | 'Beginner', 'Intermediate', 'Advanced' |
| progress | INT | DEFAULT 0 | Progress percentage (0-100) |
| xp_reward | INT | DEFAULT 100 | XP awarded on completion |
| is_public | BOOLEAN | DEFAULT false | Public visibility |
| is_favorite | BOOLEAN | DEFAULT false | Favorited by owner |
| is_template | BOOLEAN | DEFAULT false | Is a template project |
| hardware_platform | VARCHAR(50) | | 'arduino_uno', 'esp32', 'raspberry_pi', etc. |
| tags | VARCHAR(255)[] | | Array of tags |
| schema_data | JSONB | | Circuit schema JSON data |
| code_data | JSONB | | Code/firmware data |
| simulation_data | JSONB | | Simulation configuration |
| completed_at | TIMESTAMP | | Completion timestamp |
| created_at | TIMESTAMP | DEFAULT NOW() | Creation date |
| updated_at | TIMESTAMP | | Last update |

#### Table: `project_collaborators`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| project_id | UUID | FOREIGN KEY | Reference to projects |
| user_id | UUID | FOREIGN KEY | Collaborator user |
| role | ENUM | DEFAULT 'viewer' | 'owner', 'editor', 'viewer' |
| invited_at | TIMESTAMP | DEFAULT NOW() | |
| accepted_at | TIMESTAMP | | |

#### Table: `project_components`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| project_id | UUID | FOREIGN KEY | Reference to projects |
| component_id | UUID | FOREIGN KEY | Reference to components |
| quantity | INT | DEFAULT 1 | Number of components used |
| position_x | FLOAT | | X position on canvas |
| position_y | FLOAT | | Y position on canvas |
| rotation | FLOAT | DEFAULT 0 | Rotation angle |
| config_data | JSONB | | Component configuration |

---

### Components Module

#### Table: `component_categories`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| name | VARCHAR(100) | NOT NULL | Category name |
| slug | VARCHAR(100) | UNIQUE | URL-friendly slug |
| icon | VARCHAR(50) | | Icon identifier |
| color | VARCHAR(50) | | Color class (gradient) |
| description | TEXT | | Category description |
| order | INT | DEFAULT 0 | Display order |

#### Table: `components`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| category_id | UUID | FOREIGN KEY | Reference to categories |
| name | VARCHAR(200) | NOT NULL | Component name |
| description | TEXT | | Component description |
| manufacturer | VARCHAR(100) | | Manufacturer name |
| part_number | VARCHAR(100) | | Part/model number |
| specs | JSONB | | Specifications as JSON |
| price | DECIMAL(12,2) | | Price in IDR |
| stock | INT | DEFAULT 0 | Available stock |
| rating | DECIMAL(2,1) | DEFAULT 0 | Average rating (0-5) |
| rating_count | INT | DEFAULT 0 | Number of ratings |
| image_url | VARCHAR(500) | | Component image |
| datasheet_url | VARCHAR(500) | | Link to datasheet PDF |
| simulation_model | VARCHAR(100) | | Simulation model identifier |
| is_active | BOOLEAN | DEFAULT true | Active/available |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | | |

#### Table: `component_requests`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| user_id | UUID | FOREIGN KEY | Requester |
| component_name | VARCHAR(200) | NOT NULL | Requested component name |
| manufacturer | VARCHAR(100) | | Manufacturer |
| part_number | VARCHAR(100) | | Part number |
| category | VARCHAR(50) | NOT NULL | Category |
| description | TEXT | NOT NULL | Description |
| use_case | TEXT | NOT NULL | Use case explanation |
| features | VARCHAR(100)[] | | Feature tags |
| datasheet_url | VARCHAR(500) | | Datasheet link |
| product_url | VARCHAR(500) | | Product page link |
| priority | ENUM | DEFAULT 'medium' | 'low', 'medium', 'high', 'urgent' |
| status | ENUM | DEFAULT 'pending' | 'pending', 'reviewing', 'approved', 'rejected' |
| admin_notes | TEXT | | Admin response notes |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | | |

#### Table: `user_favorite_components`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| user_id | UUID | FOREIGN KEY | |
| component_id | UUID | FOREIGN KEY | |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### Challenges Module

#### Table: `challenges`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| title | VARCHAR(200) | NOT NULL | Challenge title |
| description | TEXT | | Challenge description |
| difficulty | ENUM | NOT NULL | 'Easy', 'Medium', 'Hard', 'Expert' |
| xp_reward | INT | NOT NULL | XP awarded |
| category | VARCHAR(50) | | 'Basics', 'IoT', 'Projects', 'Advanced', 'Industrial' |
| type | ENUM | DEFAULT 'regular' | 'daily', 'weekly', 'special', 'regular' |
| time_limit_hours | INT | | Time limit in hours |
| max_participants | INT | | Max participants (null = unlimited) |
| prerequisites | UUID[] | | Array of prerequisite challenge IDs |
| instructions | JSONB | | Step-by-step instructions |
| validation_criteria | JSONB | | Criteria for completion |
| is_active | BOOLEAN | DEFAULT true | Currently active |
| starts_at | TIMESTAMP | | Start time (for timed challenges) |
| ends_at | TIMESTAMP | | End time |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | | |

#### Table: `challenge_progress`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| user_id | UUID | FOREIGN KEY | |
| challenge_id | UUID | FOREIGN KEY | |
| progress | INT | DEFAULT 0 | Progress percentage (0-100) |
| status | ENUM | DEFAULT 'not_started' | 'not_started', 'in_progress', 'completed', 'failed' |
| current_step | INT | DEFAULT 0 | Current step index |
| submission_data | JSONB | | User's submission data |
| feedback | TEXT | | System/admin feedback |
| xp_earned | INT | | Actual XP earned |
| time_spent_minutes | INT | DEFAULT 0 | Time spent |
| started_at | TIMESTAMP | | When user started |
| completed_at | TIMESTAMP | | When completed |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | | |

#### Table: `daily_challenges`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| challenge_id | UUID | FOREIGN KEY | Base challenge |
| date | DATE | UNIQUE, NOT NULL | Date for this daily |
| xp_multiplier | DECIMAL(3,2) | DEFAULT 1.0 | XP multiplier (e.g., 2.0 for 2X) |
| participant_count | INT | DEFAULT 0 | Current participants |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### Notifications Module

#### Table: `notifications`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| user_id | UUID | FOREIGN KEY | Recipient |
| type | ENUM | NOT NULL | 'achievement', 'xp', 'project', 'challenge', 'social', 'system' |
| title | VARCHAR(200) | NOT NULL | Notification title |
| message | TEXT | NOT NULL | Notification body |
| data | JSONB | | Additional data (links, IDs, etc.) |
| is_read | BOOLEAN | DEFAULT false | Read status |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

### Gamification Module

#### Table: `achievements`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| name | VARCHAR(100) | NOT NULL | Achievement name |
| description | TEXT | | Achievement description |
| icon | VARCHAR(50) | | Icon identifier |
| rarity | ENUM | NOT NULL | 'common', 'rare', 'epic', 'legendary' |
| requirement_type | VARCHAR(50) | | 'challenges_completed', 'projects_created', 'streak_days', etc. |
| requirement_value | INT | | Required value to unlock |
| xp_reward | INT | DEFAULT 0 | XP awarded |
| is_active | BOOLEAN | DEFAULT true | |
| created_at | TIMESTAMP | DEFAULT NOW() | |

#### Table: `user_achievements`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| user_id | UUID | FOREIGN KEY | |
| achievement_id | UUID | FOREIGN KEY | |
| unlocked_at | TIMESTAMP | DEFAULT NOW() | When unlocked |

#### Table: `leaderboards`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| type | ENUM | NOT NULL | 'weekly', 'monthly', 'all_time' |
| period_start | DATE | | Period start date |
| period_end | DATE | | Period end date |
| created_at | TIMESTAMP | DEFAULT NOW() | |

#### Table: `leaderboard_entries`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| leaderboard_id | UUID | FOREIGN KEY | |
| user_id | UUID | FOREIGN KEY | |
| rank | INT | NOT NULL | Rank position |
| xp | INT | NOT NULL | XP earned in period |
| created_at | TIMESTAMP | DEFAULT NOW() | |

#### Table: `user_streaks`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| user_id | UUID | FOREIGN KEY, UNIQUE | |
| current_streak | INT | DEFAULT 0 | Current streak days |
| longest_streak | INT | DEFAULT 0 | Longest streak ever |
| last_activity_date | DATE | | Last active date |
| streak_protects | INT | DEFAULT 0 | Streak freeze count |
| updated_at | TIMESTAMP | | |

---

## API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | Register new user |
| POST | `/auth/login` | User login |
| POST | `/auth/logout` | User logout |
| POST | `/auth/refresh` | Refresh access token |
| POST | `/auth/forgot-password` | Request password reset |
| POST | `/auth/reset-password` | Reset password |
| GET | `/auth/verify-email/:token` | Verify email |

### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/users/me` | Get current user profile |
| PUT | `/users/me` | Update current user profile |
| PUT | `/users/me/avatar` | Upload profile picture |
| PUT | `/users/me/password` | Change password |
| GET | `/users/me/settings` | Get user settings |
| PUT | `/users/me/settings` | Update user settings |
| GET | `/users/me/stats` | Get user gamification stats |
| GET | `/users/:id` | Get user public profile |

### Projects

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects` | List user's projects |
| POST | `/projects` | Create new project |
| GET | `/projects/:id` | Get project details |
| PUT | `/projects/:id` | Update project |
| DELETE | `/projects/:id` | Delete project |
| POST | `/projects/:id/duplicate` | Duplicate project |
| PUT | `/projects/:id/favorite` | Toggle favorite |
| GET | `/projects/:id/collaborators` | List collaborators |
| POST | `/projects/:id/collaborators` | Add collaborator |
| DELETE | `/projects/:id/collaborators/:userId` | Remove collaborator |
| GET | `/projects/templates` | List project templates |

### Components

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/components` | List components |
| GET | `/components/:id` | Get component details |
| GET | `/components/categories` | List categories |
| GET | `/components/search` | Search components |
| POST | `/components/request` | Request new component |
| GET | `/components/favorites` | List user's favorites |
| POST | `/components/:id/favorite` | Toggle favorite |

### Challenges

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/challenges` | List challenges |
| GET | `/challenges/daily` | Get today's daily challenge |
| GET | `/challenges/:id` | Get challenge details |
| POST | `/challenges/:id/start` | Start a challenge |
| PUT | `/challenges/:id/progress` | Update progress |
| POST | `/challenges/:id/submit` | Submit challenge completion |
| GET | `/challenges/progress` | Get user's challenge progress |

### Notifications

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/notifications` | List notifications |
| PUT | `/notifications/:id/read` | Mark as read |
| PUT | `/notifications/read-all` | Mark all as read |
| DELETE | `/notifications/:id` | Delete notification |

### Gamification

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/achievements` | List all achievements |
| GET | `/achievements/user` | List user's unlocked achievements |
| GET | `/leaderboard` | Get leaderboard |
| GET | `/streak` | Get user's streak info |

---

## Request/Response Payloads

### Authentication

#### POST /auth/register

**Request:**
```json
{
  "email": "user@example.com",
  "password_hash": "securepassword123",
  "name": "John Doe",
  "role": "student"
}
```

**Response (201):**
```json
{
  "success": true,
  "message": "Registration successful. Please verify your email.",
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "role": "student"
    }
  }
}
```

#### POST /auth/login

**Request:**
```json
{
  "email": "user@example.com",
  "password_hash": "securepassword123"
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2024-12-21T00:00:00Z",
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "role": "student",
      "avatar_url": null,
      "level": 12,
      "total_xp": 2350,
      "current_xp": 2350,
      "target_xp": 3000,
      "streak_days": 7,
      "is_pro": false,
      "language": "id",
      "theme": "dark"
    }
  }
}
```

---

### Users

#### GET /users/me

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "username": "@johndoe",
    "role": "student",
    "avatar_url": "https://...",
    "bio": "IoT enthusiast...",
    "level": 12,
    "total_xp": 2350,
    "current_xp": 2350,
    "target_xp": 3000,
    "streak_days": 7,
    "language": "id",
    "theme": "dark",
    "is_pro": false,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### PUT /users/me

**Request:**
```json
{
  "name": "John Doe Updated",
  "username": "@johnd",
  "bio": "Updated bio...",
  "language": "en",
  "theme": "light"
}
```

---

### Projects

#### GET /projects

**Query Parameters:**
- `page` (int): Page number
- `limit` (int): Items per page
- `search` (string): Search query
- `difficulty` (string): Filter by difficulty
- `is_favorite` (bool): Filter favorites only
- `sort` (string): 'created_at', 'updated_at', 'name'

**Response (200):**
```json
{
  "success": true,
  "data": {
    "projects": [
      {
        "id": "uuid",
        "name": "Arduino Smart Home",
        "description": "Sistem kontrol rumah pintar...",
        "thumbnail_url": "https://...",
        "difficulty": "Intermediate",
        "progress": 75,
        "xp_reward": 500,
        "is_public": false,
        "is_favorite": true,
        "hardware_platform": "arduino_uno",
        "tags": ["iot", "arduino", "smart-home"],
        "components_count": 12,
        "collaborators_count": 2,
        "created_at": "2024-12-01T00:00:00Z",
        "updated_at": "2024-12-20T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 25,
      "total_pages": 3
    }
  }
}
```

#### POST /projects

**Request:**
```json
{
  "name": "My New Project",
  "description": "Project description...",
  "difficulty": "Beginner",
  "hardware_platform": "esp32",
  "tags": ["iot", "esp32"],
  "is_public": false
}
```

**Response (201):**
```json
{
  "success": true,
  "message": "Project created successfully",
  "data": {
    "id": "uuid",
    "name": "My New Project",
    "created_at": "2024-12-20T00:00:00Z"
  }
}
```

---

### Components

#### GET /components

**Query Parameters:**
- `page` (int): Page number
- `limit` (int): Items per page
- `category` (string): Category slug
- `search` (string): Search query
- `in_stock` (bool): Filter in-stock only
- `sort` (string): 'name', 'price', 'rating'

**Response (200):**
```json
{
  "success": true,
  "data": {
    "components": [
      {
        "id": "uuid",
        "name": "Arduino Uno R3",
        "category": {
          "id": "uuid",
          "name": "Microcontrollers",
          "slug": "microcontrollers"
        },
        "description": "Classic microcontroller board...",
        "manufacturer": "Arduino",
        "part_number": "A000066",
        "specs": {
          "processor": "ATmega328P",
          "clock_speed": "16MHz",
          "digital_io": 14,
          "analog_inputs": 6,
          "voltage": "5V"
        },
        "price": 85000,
        "stock": 25,
        "rating": 4.8,
        "rating_count": 156,
        "image_url": "https://...",
        "is_favorite": true
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 256,
      "total_pages": 13
    }
  }
}
```

#### POST /components/request

**Request:**
```json
{
  "component_name": "BME280",
  "manufacturer": "Bosch",
  "part_number": "BME280-001",
  "category": "sensor",
  "description": "Temperature, humidity, pressure sensor...",
  "use_case": "Weather monitoring project...",
  "features": ["I2C", "SPI", "3.3V", "Low Power"],
  "datasheet_url": "https://example.com/datasheet.pdf",
  "product_url": "https://example.com/product",
  "priority": "medium"
}
```

---

### Challenges

#### GET /challenges

**Query Parameters:**
- `type` (string): 'all', 'daily', 'weekly', 'special'
- `difficulty` (string): 'Easy', 'Medium', 'Hard', 'Expert'
- `status` (string): 'available', 'in_progress', 'completed', 'locked'
- `category` (string): Category filter

**Response (200):**
```json
{
  "success": true,
  "data": {
    "challenges": [
      {
        "id": "uuid",
        "title": "LED Blink Basics",
        "description": "Pelajari dasar-dasar kontrol LED...",
        "difficulty": "Easy",
        "xp_reward": 100,
        "category": "Basics",
        "type": "regular",
        "time_left": null,
        "participants": 5420,
        "user_status": "completed",
        "user_progress": 100
      }
    ],
    "stats": {
      "total_completed": 12,
      "total_xp_earned": 1500,
      "current_streak": 7
    }
  }
}
```

#### GET /challenges/daily

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "challenge": {
      "id": "uuid",
      "title": "Buat Smart Light Controller",
      "description": "Rancang sistem kontrol lampu pintar...",
      "difficulty": "Medium",
      "xp_reward": 200
    },
    "date": "2024-12-20",
    "xp_multiplier": 2.0,
    "time_left_hours": 8,
    "participant_count": 1200,
    "user_status": "not_started"
  }
}
```

#### POST /challenges/:id/submit

**Request:**
```json
{
  "submission_data": {
    "project_id": "uuid",
    "notes": "Completed all requirements...",
    "attachments": ["https://..."]
  }
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Challenge completed!",
  "data": {
    "xp_earned": 200,
    "bonus_xp": 50,
    "total_xp": 250,
    "new_level": 13,
    "level_up": true,
    "achievements_unlocked": [
      {
        "id": "uuid",
        "name": "Challenge Master",
        "rarity": "rare"
      }
    ]
  }
}
```

---

### Notifications

#### GET /notifications

**Query Parameters:**
- `page` (int): Page number
- `limit` (int): Items per page
- `type` (string): Filter by type
- `unread_only` (bool): Only unread

**Response (200):**
```json
{
  "success": true,
  "data": {
    "notifications": [
      {
        "id": "uuid",
        "type": "achievement",
        "title": "Achievement Unlocked!",
        "message": "You've earned the 'First Challenge' badge!",
        "data": {
          "achievement_id": "uuid",
          "link": "/achievements"
        },
        "is_read": false,
        "created_at": "2024-12-20T10:00:00Z"
      }
    ],
    "unread_count": 5,
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45
    }
  }
}
```

---

### Gamification

#### GET /users/me/stats

**Response (200):**
```json
{
  "success": true,
  "data": {
    "level": 12,
    "total_xp": 2350,
    "current_xp": 2350,
    "target_xp": 3000,
    "xp_to_next_level": 650,
    "streak": {
      "current": 7,
      "longest": 14,
      "last_active": "2024-12-20"
    },
    "challenges": {
      "completed": 12,
      "in_progress": 2,
      "total_available": 50
    },
    "projects": {
      "total": 8,
      "completed": 5,
      "in_progress": 3
    },
    "achievements": {
      "unlocked": 8,
      "total": 25
    },
    "daily_goal": {
      "completed": 3,
      "target": 5
    },
    "xp_this_week": 2450
  }
}
```

#### GET /leaderboard

**Query Parameters:**
- `type` (string): 'weekly', 'monthly', 'all_time'
- `limit` (int): Number of entries

**Response (200):**
```json
{
  "success": true,
  "data": {
    "type": "weekly",
    "period": {
      "start": "2024-12-16",
      "end": "2024-12-22"
    },
    "entries": [
      {
        "rank": 1,
        "user": {
          "id": "uuid",
          "name": "ElectroPro",
          "avatar_url": "https://..."
        },
        "xp": 45000,
        "is_current_user": false
      },
      {
        "rank": 4,
        "user": {
          "id": "uuid",
          "name": "John Doe",
          "avatar_url": "https://..."
        },
        "xp": 2350,
        "is_current_user": true
      }
    ],
    "current_user_rank": 4
  }
}
```

---

## Error Response Format

All API errors follow this format:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Email is required",
    "details": {
      "field": "email",
      "value": null
    }
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `FORBIDDEN` | 403 | No permission for this action |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Invalid request data |
| `DUPLICATE_ENTRY` | 409 | Resource already exists |
| `RATE_LIMITED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Server error |

---

## Notes for Backend Development

### Technologies
- **Language**: Go (Golang) 1.21+
- **Framework**: Fiber or Gin
- **ORM**: GORM
- **Database**: PostgreSQL 15+
- **Cache**: Redis
- **Auth**: JWT with refresh tokens

### Important Considerations

1. **Password Hashing**: Use bcrypt with cost factor 12
2. **UUID Generation**: Use UUID v4 for all IDs
3. **Timestamps**: Use UTC for all timestamps
4. **Pagination**: Default limit 20, max 100
5. **Rate Limiting**: 100 requests/minute per user
6. **File Uploads**: Use MinIO/S3 for file storage
7. **WebSockets**: For real-time notifications and collaboration

---

*Generated for NexFlux Backend Development*
