# Documentation Module API Specification

Dokumen ini menjelaskan endpoint API yang dibutuhkan untuk module Documentation di NexFlux Platform.

---

## Database Schema

### Table: `doc_categories`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| name | VARCHAR(100) | NOT NULL | Category name |
| slug | VARCHAR(100) | UNIQUE, NOT NULL | URL-friendly identifier |
| description | TEXT | | Category description |
| icon | VARCHAR(50) | | Icon identifier |
| color | VARCHAR(50) | | Gradient color class |
| order | INT | DEFAULT 0 | Display order |
| is_active | BOOLEAN | DEFAULT true | |
| created_at | TIMESTAMP | DEFAULT NOW() | |

### Table: `doc_articles`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| category_id | UUID | FOREIGN KEY | Reference to doc_categories |
| title | VARCHAR(300) | NOT NULL | Article title |
| slug | VARCHAR(300) | UNIQUE, NOT NULL | URL-friendly identifier |
| excerpt | TEXT | | Short description/summary |
| content | TEXT | NOT NULL | Full markdown content |
| author_id | UUID | FOREIGN KEY | Reference to users |
| read_time_minutes | INT | DEFAULT 5 | Estimated read time |
| difficulty | ENUM | DEFAULT 'beginner' | 'beginner', 'intermediate', 'advanced' |
| tags | VARCHAR(50)[] | | Array of tags |
| views | INT | DEFAULT 0 | View count |
| is_featured | BOOLEAN | DEFAULT false | Featured article |
| is_published | BOOLEAN | DEFAULT true | |
| published_at | TIMESTAMP | | Publication date |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | | |

### Table: `doc_videos`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| category_id | UUID | FOREIGN KEY | Reference to doc_categories |
| title | VARCHAR(300) | NOT NULL | Video title |
| description | TEXT | | Video description |
| video_url | VARCHAR(500) | NOT NULL | YouTube/Vimeo URL |
| thumbnail_url | VARCHAR(500) | | Thumbnail image |
| duration_seconds | INT | | Duration in seconds |
| difficulty | ENUM | DEFAULT 'beginner' | 'beginner', 'intermediate', 'advanced' |
| views | INT | DEFAULT 0 | View count |
| is_featured | BOOLEAN | DEFAULT false | |
| is_published | BOOLEAN | DEFAULT true | |
| created_at | TIMESTAMP | DEFAULT NOW() | |

---

## API Endpoints

### Documentation

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/docs/categories` | List all documentation categories |
| GET | `/docs/categories/:slug` | Get category details with articles |
| GET | `/docs/articles` | List all articles (with filters) |
| GET | `/docs/articles/:slug` | Get single article by slug |
| GET | `/docs/articles/popular` | Get popular articles |
| GET | `/docs/articles/featured` | Get featured articles |
| GET | `/docs/videos` | List all video tutorials |
| GET | `/docs/videos/:id` | Get single video details |
| GET | `/docs/search` | Search articles and videos |
| POST | `/docs/articles/:slug/view` | Increment article view count |

---

## Request/Response Payloads

### GET /docs/categories

**Response (200):**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Getting Started",
      "slug": "getting-started",
      "description": "Pelajari dasar-dasar NexFlux dan mulai perjalananmu",
      "icon": "rocket",
      "color": "from-brand-500 to-neon-500",
      "articles_count": 12,
      "order": 1
    },
    {
      "id": "uuid",
      "name": "Hardware Guides",
      "slug": "hardware",
      "description": "Panduan lengkap untuk berbagai komponen hardware",
      "icon": "microchip",
      "color": "from-blue-500 to-cyan-500",
      "articles_count": 48,
      "order": 2
    }
  ]
}
```

---

### GET /docs/categories/:slug

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Getting Started",
    "slug": "getting-started",
    "description": "Pelajari dasar-dasar NexFlux dan mulai perjalananmu",
    "icon": "rocket",
    "color": "from-brand-500 to-neon-500",
    "articles": [
      {
        "id": "uuid",
        "title": "Cara Memulai dengan Arduino Uno",
        "slug": "memulai-arduino-uno",
        "excerpt": "Panduan lengkap untuk pemula...",
        "read_time_minutes": 10,
        "difficulty": "beginner",
        "tags": ["arduino", "beginner", "tutorial"],
        "views": 15420,
        "is_featured": true,
        "published_at": "2024-12-20T10:00:00Z",
        "author": {
          "id": "uuid",
          "name": "Admin NexFlux",
          "avatar_url": "https://..."
        }
      }
    ],
    "articles_count": 12
  }
}
```

---

### GET /docs/articles

**Query Parameters:**
- `page` (int): Page number
- `limit` (int): Items per page  
- `category` (string): Category slug
- `difficulty` (string): 'beginner', 'intermediate', 'advanced'
- `search` (string): Search query
- `sort` (string): 'newest', 'popular', 'title'

**Response (200):**
```json
{
  "success": true,
  "data": {
    "articles": [
      {
        "id": "uuid",
        "title": "Cara Memulai dengan Arduino Uno",
        "slug": "memulai-arduino-uno",
        "excerpt": "Panduan lengkap untuk pemula...",
        "category": {
          "id": "uuid",
          "name": "Getting Started",
          "slug": "getting-started"
        },
        "read_time_minutes": 10,
        "difficulty": "beginner",
        "tags": ["arduino", "beginner"],
        "views": 15420,
        "is_featured": true,
        "published_at": "2024-12-20T10:00:00Z",
        "author": {
          "id": "uuid",
          "name": "Admin NexFlux",
          "avatar_url": "https://..."
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 159,
      "total_pages": 8
    }
  }
}
```

---

### GET /docs/articles/:slug

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "title": "Cara Memulai dengan Arduino Uno: Panduan Lengkap untuk Pemula",
    "slug": "memulai-arduino-uno",
    "excerpt": "Panduan lengkap untuk pemula...",
    "content": "# Pendahuluan\n\nArduino Uno adalah...\n\n## Persiapan\n\n1. Download Arduino IDE...",
    "category": {
      "id": "uuid",
      "name": "Getting Started", 
      "slug": "getting-started"
    },
    "author": {
      "id": "uuid",
      "name": "Admin NexFlux",
      "avatar_url": "https://..."
    },
    "read_time_minutes": 10,
    "difficulty": "beginner",
    "tags": ["arduino", "beginner", "tutorial", "uno"],
    "views": 15420,
    "is_featured": true,
    "published_at": "2024-12-20T10:00:00Z",
    "updated_at": "2024-12-21T08:00:00Z",
    "related_articles": [
      {
        "id": "uuid",
        "title": "Mengenal Pin Arduino Uno",
        "slug": "mengenal-pin-arduino-uno",
        "excerpt": "Penjelasan lengkap tentang pin...",
        "read_time_minutes": 8,
        "difficulty": "beginner"
      }
    ],
    "table_of_contents": [
      { "id": "pendahuluan", "title": "Pendahuluan", "level": 1 },
      { "id": "persiapan", "title": "Persiapan", "level": 2 },
      { "id": "instalasi", "title": "Instalasi", "level": 2 }
    ]
  }
}
```

---

### GET /docs/articles/popular

**Query Parameters:**
- `limit` (int): Number of articles (default: 10)

**Response (200):**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "title": "Cara Memulai dengan Arduino Uno",
      "slug": "memulai-arduino-uno",
      "category": {
        "name": "Getting Started",
        "slug": "getting-started"
      },
      "read_time_minutes": 10,
      "views": 15420,
      "is_new": true
    }
  ]
}
```

---

### GET /docs/videos

**Query Parameters:**
- `page` (int): Page number
- `limit` (int): Items per page
- `category` (string): Category slug
- `difficulty` (string): Difficulty filter

**Response (200):**
```json
{
  "success": true,
  "data": {
    "videos": [
      {
        "id": "uuid",
        "title": "Belajar Arduino dalam 30 Menit",
        "description": "Tutorial lengkap untuk pemula...",
        "video_url": "https://youtube.com/watch?v=xxxxx",
        "thumbnail_url": "https://img.youtube.com/...",
        "duration_seconds": 1935,
        "difficulty": "beginner",
        "views": 8500,
        "category": {
          "name": "Getting Started",
          "slug": "getting-started"
        },
        "created_at": "2024-12-15T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45,
      "total_pages": 3
    }
  }
}
```

---

### GET /docs/search

**Query Parameters:**
- `q` (string): Search query (required)
- `type` (string): 'all', 'articles', 'videos' (default: 'all')
- `limit` (int): Results per type (default: 10)

**Response (200):**
```json
{
  "success": true,
  "data": {
    "articles": [
      {
        "id": "uuid",
        "title": "Cara Memulai dengan Arduino",
        "slug": "memulai-arduino",
        "excerpt": "...",
        "category": { "name": "Getting Started", "slug": "getting-started" },
        "read_time_minutes": 10,
        "views": 15420
      }
    ],
    "videos": [
      {
        "id": "uuid",
        "title": "Tutorial Arduino",
        "thumbnail_url": "...",
        "duration_seconds": 1200,
        "difficulty": "beginner"
      }
    ],
    "total_results": 25
  }
}
```

---

### POST /docs/articles/:slug/view

Increment view count for an article (called when user opens article).

**Response (200):**
```json
{
  "success": true,
  "message": "View recorded",
  "data": {
    "views": 15421
  }
}
```

---

## Frontend Routes Required

| Route | Component | Description |
|-------|-----------|-------------|
| `/docs` | `Docs.tsx` | Main documentation page (exists) |
| `/docs/:categorySlug` | `DocCategory.tsx` | Category listing page |
| `/docs/:categorySlug/:articleSlug` | `DocArticle.tsx` | Article detail page |
| `/docs/video/:id` | `DocVideo.tsx` | Video player page |
| `/docs/search` | `DocSearch.tsx` | Search results page |

---

## Notes

1. Article content uses **Markdown** format
2. Views are tracked via POST request when article is opened
3. Related articles are auto-suggested based on tags and category
4. Table of contents is generated from markdown headings
5. Videos support YouTube and Vimeo embeds
