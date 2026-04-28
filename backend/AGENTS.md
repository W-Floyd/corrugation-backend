# Backend Architecture

## Tech Stack

- **Framework**: Huma v2 + Echo for HTTP API
- **Database**: SQLite with GORM ORM
- **Authentication**: OIDC/JWT via Authentik
- **Embeddings**: Infinity server (OpenAI CLIP, BGE models)
- **Concurrency**: Goroutine workers, semaphore-controlled embedding pool

## Core Data Models

### Record
- ID, Quantity, ReferenceNumber (unique) flag
- Title, Description, ParentID (hierarchical)
- Tags (many-to-many), Artifacts (references), OwnerID
- SearchConfidenceImage, SearchConfidenceText (computed scores)
- Owner relationship (User)

### Tag
- Title (PK), Color
- Records (many-to-many)

### Artifact
- ID, Data, OriginalFilename, ContentType
- SmallPreviewID, LargePreviewID (preview images)
- RecordID (link to record)
- Supports image previews (WebP, resized)

### Embedding
- RecordID, ArtifactID (both nullable)
- EmbedModel (model identifier)
- Data (binary vector), Hash (dedup key)

### EmbeddingJob
- JobType (record/artifact), TargetID, EmbedModel
- Status (pending/processing/done/failed)
- OwnerID, Username, Source (store/search/backfill)

### User
- Username (unique), Infinity config overrides per-user

### GlobalConfig (singleton ID=1)
- LogLevel, GenerateEmbeddingsOnStart

## API Endpoints

### Records
- `GET /api/v2/record/{id}` - Single record
- `GET /api/v2/records` - List with query params (search, depth, levels)
- `POST /api/v2/record` - Create
- `POST /api/v2/record/{id}` - Update
- `DELETE /api/v2/record/{id}` - Delete

### Artifacts
- `GET /api/v2/artifacts` - List
- `POST /api/v2/artifact` - Create (multipart)
- `GET /api/v2/artifact` - Get
- `DELETE /api/v2/artifact/{id}` - Delete

### Tags
- `GET /api/v2/tags` - List
- `GET /api/v2/tags` - Get
- `POST /api/v2/tags` - Create
- `DELETE /api/v2/tags/{id}` - Delete

### Config
- `GET /api/v2/config/global` - Global settings
- `PUT /api/v2/config/global` - Update global
- `GET /api/v2/config/user` - User config
- `PUT /api/v2/config/user` - Update user config

### Auth
- `GET /api/auth/config` - OIDC config for frontend

### Legacy Store API (Entity-style)
- `GET /api/store` - Export full store
- `POST /api/store` - Create entity
- `POST /api/store/{id}` - Replace/Patch entity
- `GET /api/store/{id}` - Get entity
- `DELETE /api/store/{id}` - Delete
- `GET /api/store/{id}/qr` - QR code

### Embeddings
- `POST /api/v2/embeddings/flush` - Delete stale embeddings
- `GET /api/v2/embeddings/progress` - Check job progress
- `GET /api/v2/embeddings/search-progress` - Search progress

### Import
- `POST /api/import` - Import legacy tar.gz

### Visualization
- `GET /api/v2/records/visualize` - HTML graph view

## Embedding System

### Architecture
- Embedded worker pool (default 4 concurrency) controlled by semaphore
- Two job queues: `embeddingJobQueue` (regular), `embeddingSearchJobQueue` (search)
- Deduplication at DB level (job for same target+model+status)
- Fast-path: check existing embedding before enqueuing

### Workers
1. Process artifact embeddings via `Image.GenerateEmbeddings()`
2. Process record embeddings via `Record.GenerateEmbeddings()`
3. Broadcast progress via websocket (`embedding_progress:{type}:{id}`)

### Models
- **Infinity Image**: `openai/clip-vit-large-patch14`
- **Infinity Text**: `BAAI/bge-large-en-v1.5`
- Configurable per-user or global via `SetInfinityConfig()`

### Backfill
- Runs on startup if `generateEmbeddingsOnStart=true`
- Groups records by owner for per-user model overrides
- Skips already-embedded items via hash comparison

## Search Flow

### Text Search
1. Query → text prefix → embed
2. Dot product with record embeddings
3. Filter by `minTextScore` (default 0.9)

### Image Search
1. Query image → clip embed
2. Dot product with artifact embeddings
3. Map artifact→record via `artifactRecordMap`
4. Filter by `minImageScore` (default 0.2)

### Hybrid
- Combines text + image results
- Supports substring matching (`searchTextSubstring`)
- Respects record hierarchy (`childrenDepth`, `parentDepth`)

## Entity Store Legacy

Old entity model (still supported):
- Entity → Record wrapper
- `/api/store` endpoints mirror entity API
- `ToEntity()` converts Record→legacy format

## Auth

### OIDC Flow
- Discovery URL → JWKS cache
- Bearer token validation → username context
- Middleware guards `/api/*` (excludes `/api/auth/`)
- Token cache with 10min refresh

### Anonymous
- Empty username when auth disabled
- All requests treated as anonymous

## Config System

### Global Config
- `LogLevel`: silent/panic/error/warn/info/debug
- `GenerateEmbeddingsOnStart`: backfill flag

### Per-User Config
- Override Infinity models per user
- Custom query/document prefixes
- Falls back to global/env defaults

## Workers & Concurrency

- **Embedding workers**: Fixed pool (default 4), semaphore-gated
- **WebSocket broadcaster**: `BroadcastToUser()` for real-time progress
- **DB pool**: 10 idle/open connections, WAL mode, 64MB cache
- **Dedup**: Only one pending/processing job per target+model

## Database Schema

```sql
records: id, quantity, reference_number (unique), title, description, parent_id, owner_id, timestamps
tags: title (pk), color, timestamps
artifacts: id, data, original_filename, content_type, small_preview_id, large_preview_id, record_id, timestamps
embeddings: record_id, artifact_id, embed_model, data, hash, timestamps
embedding_jobs: job_type, target_id, owner_id, username, status, error_msg, embed_model, source, timestamps
users: id, username (unique), infinity config columns, timestamps
global_config: id=1, log_level, generate_embeddings_on_start
record_tags: record_id, tag_title (junction)
```

## Key Functions

- `ConnectDB()`: SQLite connection with WAL, pool optimization
- `InitAndMigrateDB()`: AutoMigrate all models
- `StartEmbeddingWorkers()`: Launch worker goroutines
- `BackfillEmbeddings()`: Full backfill for all records/artifacts
- `GenerateRecordEmbeddings()`, `GenerateArtifactEmbeddings()`: Per-entity embedding
- `SearchByRecord()`, `SearchByArtifact()`: Vector search
- `ImportFromReader()`: Legacy tar.gz import
- `RegisterHandlers()`: Register all Huma endpoints

## Entry Point (`main.go`)

1. Parse CLI flags (port, data path, OIDC, Infinity config)
2. Connect DB + migrate
3. Setup OIDC if configured
4. Register handlers + auth middleware
5. Start embedding workers
6. Trigger backfill if flag set
7. Listen on `:8083`

## File Structure

```
backend/
├── artifact*.go        # Artifact model, preview handling
├── auth.go             # OIDC middleware, token validation
├── backfill.go         # Backfill logic for embeddings
├── config-handler.go   # Global/user config endpoints
├── constants.go        # Infinity URLs, defaults
├── db.go               # DB connection, migrations
├── embedding*.go       # Embedding model, queue, workers
├── export.go           # Export endpoints
├── handlers.go         # Register all API routes
├── import.go           # Legacy import handler
├── infinity.go         # Infinity client calls
├── logger.go           # Structured logging
├── record*.go          # Record model, embedding gen
├── search.go           # Vector search logic
├── store.go            # Legacy entity API
├── tag*.go             # Tag model/handlers
├── users.go            # User config, cache
├── utils.go            # Helpers
└── ws.go               # WebSocket broadcaster
```

## Integration with Frontend

- Frontend fetches `/api/auth/config` to trigger OIDC flow
- Real-time embedding progress via WebSocket (`ws://.../ws`)
- Config endpoints for user preferences
- Import endpoint for bulk data load

## Design Patterns

- **Repository**: GORM generic `gorm.G[T](db)` pattern
- **Worker Pool**: Fixed goroutines + channel queue
- **Cache-Aside**: `embeddingsCache` sync.Map for vectors
- **Singleton**: GlobalConfig always ID=1
- **Facade**: Backend package abstracts DB/embedding complexity
