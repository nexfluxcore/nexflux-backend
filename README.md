# NexFlux Virtual Lab - Backend API

> Backend API untuk NexFlux Virtual Lab - Platform pembelajaran elektronik dan IoT dengan gamifikasi.

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![Gin Framework](https://img.shields.io/badge/Gin-Web_Framework-00ADD8)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-Cache-DC382D?logo=redis)

---

## 📋 Table of Contents

1. [Features](#features)
2. [Tech Stack](#tech-stack)
3. [Project Structure](#project-structure)
4. [Getting Started](#getting-started)
5. [Database Migrations](#database-migrations)
6. [API Endpoints](#api-endpoints)
7. [Authentication](#authentication)

---

## ✨ Features

### Core Modules
- **👤 User Management** - Profile, settings, preferences
- **🔐 Authentication** - Email/password + OAuth (Google, GitHub, Apple)
- **📁 Projects** - Create, manage, collaborate on circuit projects
- **🔧 Components** - Browse, search, favorite electronic components
- **🏆 Challenges** - Daily/weekly challenges with XP rewards
- **🔔 Notifications** - Real-time notification system
- **🎮 Gamification** - XP, levels, achievements, leaderboards, streaks

### Key Highlights
- JWT-based authentication with Redis session storage
- Multi-provider OAuth2 (Google, GitHub, Apple Sign In)
- Clean architecture (Repository → Service → Handler)
- Auto database migrations
- Swagger API documentation

---

## 🛠 Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.21+ |
| Framework | Gin |
| ORM | GORM |
| Database | PostgreSQL 15+ |
| Cache | Redis |
| Authentication | JWT + OAuth2 |
| API Docs | Swagger (swaggo) |

---

## 📁 Project Structure

```
nexfi-backend/
├── api/
│   ├── handlers/           # HTTP request handlers
│   │   ├── auth.handler.go
│   │   ├── oauth.handler.go
│   │   ├── user.handler.go
│   │   ├── project.handler.go
│   │   ├── component.handler.go
│   │   ├── challenge.handler.go
│   │   ├── notification.handler.go
│   │   └── gamification.handler.go
│   ├── middleware/         # JWT, CORS, rate limiting
│   ├── repositories/       # Data access layer
│   │   ├── base.repository.go
│   │   ├── user.repository.go
│   │   ├── project.repository.go
│   │   ├── component.repository.go
│   │   ├── challenge.repository.go
│   │   ├── notification.repository.go
│   │   └── gamification.repository.go
│   ├── routes/             # API route definitions
│   └── services/           # Business logic
│       ├── auth.services.go
│       ├── oauth.services.go
│       ├── project.service.go
│       ├── component.service.go
│       ├── challenge.service.go
│       ├── notification.service.go
│       └── gamification.service.go
├── config/                 # Configuration
│   ├── config.go
│   └── oauth.go
├── database/               # Database connection & migrations
│   └── db.go
├── dto/                    # Data Transfer Objects
│   ├── auth.dto.go
│   ├── user.dto.go
│   ├── project.dto.go
│   ├── component.dto.go
│   ├── challenge.dto.go
│   └── common.dto.go
├── models/                 # Database models
│   ├── user.model.go
│   ├── project.model.go
│   ├── component.model.go
│   ├── challenge.model.go
│   ├── notification.model.go
│   └── gamification.model.go
├── redis/                  # Redis client
├── utils/                  # Utility functions
├── .env                    # Environment variables (not in git)
├── .env_example            # Example environment file
├── main.go
├── go.mod
└── go.sum
```

---

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Redis 7+

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/nexfi-backend.git
   cd nexfi-backend
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Setup environment**
   ```bash
   cp .env_example .env
   # Edit .env with your configuration
   ```

4. **Setup database**
   ```bash
   # Create database and enable UUID extension
   psql -U postgres -c "CREATE DATABASE nexflux_db;"
   psql -U postgres -d nexflux_db -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp";'
   ```

5. **Run the server**
   ```bash
   go run main.go
   ```

   Server akan berjalan di `http://localhost:8080`

---

## 🗄 Database Migrations

Backend menggunakan **auto-migration** via GORM. Set `DB_AUTO_MIGRATE=true` di `.env` untuk mengaktifkan.

Migrations dijalankan dalam urutan:
1. Users Module (users, user_settings, user_sessions, user_streaks)
2. Components Module (component_categories, components, component_requests)
3. Projects Module (projects, project_collaborators, project_components)
4. Challenges Module (challenges, challenge_progress, daily_challenges)
5. Notifications Module (notifications)
6. Gamification Module (achievements, user_achievements, leaderboards, leaderboard_entries)

---

## 📡 API Endpoints

### Base URL: `/api/v1`

### Authentication

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/auth/register` | Register new user | ❌ |
| POST | `/auth/login` | User login | ❌ |
| GET | `/auth/providers` | List OAuth providers | ❌ |
| GET | `/auth/google` | Get Google OAuth URL | ❌ |
| GET | `/auth/google/callback` | Google OAuth callback | ❌ |
| GET | `/auth/github` | Get GitHub OAuth URL | ❌ |
| GET | `/auth/github/callback` | GitHub OAuth callback | ❌ |
| GET | `/auth/apple` | Get Apple OAuth URL | ❌ |
| POST | `/auth/apple/callback` | Apple OAuth callback | ❌ |

### Users

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/users/me` | Get current user profile | ✅ |
| PUT | `/users/me` | Update profile | ✅ |
| GET | `/users/me/stats` | Get gamification stats | ✅ |
| GET | `/users/me/settings` | Get user settings | ✅ |
| PUT | `/users/me/settings` | Update settings | ✅ |

### Projects

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/projects` | List user's projects | ✅ |
| POST | `/projects` | Create new project | ✅ |
| GET | `/projects/templates` | Get project templates | ✅ |
| GET | `/projects/:id` | Get project details | ✅ |
| PUT | `/projects/:id` | Update project | ✅ |
| DELETE | `/projects/:id` | Delete project | ✅ |
| POST | `/projects/:id/duplicate` | Duplicate project | ✅ |
| PUT | `/projects/:id/favorite` | Toggle favorite | ✅ |

### Components

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/components/categories` | List categories | ❌ |
| GET | `/components` | List components | ✅ |
| GET | `/components/search` | Search components | ✅ |
| GET | `/components/favorites` | Get favorites | ✅ |
| GET | `/components/:id` | Get component details | ✅ |
| POST | `/components/:id/favorite` | Toggle favorite | ✅ |
| POST | `/components/request` | Request new component | ✅ |
| GET | `/components/requests` | Get my requests | ✅ |

### Challenges

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/challenges` | List challenges | ✅ |
| GET | `/challenges/daily` | Get daily challenge | ✅ |
| GET | `/challenges/progress` | Get my progress | ✅ |
| GET | `/challenges/:id` | Get challenge details | ✅ |
| POST | `/challenges/:id/start` | Start challenge | ✅ |
| PUT | `/challenges/:id/progress` | Update progress | ✅ |
| POST | `/challenges/:id/submit` | Submit completion | ✅ |

### Notifications

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/notifications` | List notifications | ✅ |
| PUT | `/notifications/read-all` | Mark all as read | ✅ |
| PUT | `/notifications/:id/read` | Mark as read | ✅ |
| DELETE | `/notifications/:id` | Delete notification | ✅ |

### Gamification

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/achievements` | List all achievements | ✅ |
| GET | `/achievements/user` | Get my achievements | ✅ |
| GET | `/leaderboard` | Get leaderboard | ✅ |
| GET | `/streak` | Get streak info | ✅ |

---

## 🔐 Authentication

### JWT Token

Semua endpoint yang protected memerlukan header:
```
Authorization: Bearer <jwt_token>
```

### OAuth Setup

#### Google OAuth
1. Buka [Google Cloud Console](https://console.cloud.google.com/)
2. Create project dan enable Google+ API
3. Create OAuth 2.0 credentials
4. Set redirect URI: `{OAUTH_BASE_URL}/api/v1/auth/google/callback`

#### GitHub OAuth
1. Buka [GitHub Developer Settings](https://github.com/settings/developers)
2. Create new OAuth App
3. Set callback URL: `{OAUTH_BASE_URL}/api/v1/auth/github/callback`

#### Apple Sign In
1. Buka [Apple Developer Portal](https://developer.apple.com/account/)
2. Create App ID dengan Sign In with Apple capability
3. Create Service ID untuk web authentication
4. Create Key untuk Sign In with Apple
5. Set redirect URI: `{OAUTH_BASE_URL}/api/v1/auth/apple/callback`

---

## 📝 Response Format

### Success Response
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... }
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Email is required",
    "details": { ... }
  }
}
```

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

*Built with ❤️ for NexFlux Virtual Lab*