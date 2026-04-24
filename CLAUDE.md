# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

**Build:**
```bash
go build -ldflags="-extldflags -static" -o main .
```

**Run:**
```bash
./main [--port 8083] [--data ./data] [--auth false] [--username user] [--password pass] [--jwt-secret XXX]
```

**Frontend dev server (proxies `/api` and `/ws` to localhost:8083):**
```bash
cd frontend && npm run dev
```

**Frontend build:**
```bash
cd frontend && npm run build
```

**Test (integration — requires running server):**
```bash
./test.sh
```

**Deploy to VPS:**
```bash
./push.sh user@host
```

**Docker:**
```bash
docker compose stop; docker compose up -d --build
```
Using the Firefox MCP, run the docker command and connect to `http://localhost:8083/` to inspect.

## Architecture

**Backend:** Go + Echo v4 HTTP framework. All routes are registered in `cmd/root.go`. Business logic lives in `cmd/entity.go`, `cmd/store.go`, `cmd/artifacts.go`, and `cmd/auth.go`. The backend serves the compiled Vue SPA from `dist/index.html` via a catch-all route, and `dist/assets/` statically.

**Storage:** A single `Store` struct (entities + artifacts + version counter) is serialized to disk via diskv. There is no database — reads and writes go through `loadStore()` / `updateStore()` in `cmd/store.go`. `StoreVersion` is incremented on every write; `updateStore()` also calls `hub.broadcast()` to push change notifications to all connected WebSocket clients.

**Frontend:** Vue 3 + TypeScript SPA built with Vite, located in `frontend/`. State management via Pinia. Routing via Vue Router (hash mode, URL format `/#/<entityId>`). Styled with Tailwind CSS v4 via PostCSS. Built output goes to `../dist/`.

## Frontend Structure

### API Layer
- `frontend/src/api/types.ts` — shared TypeScript types (`Entity`, `Artifact`, `FullState`, `Metadata`, `EntityCreate`, `EntityUpdate`)
- `frontend/src/api/index.ts` — typed fetch-based API client with authentication handling
  - Core methods: `getFullState()`, `createEntity()`, `updateEntity()`, `patchEntity()`, `deleteEntity()`, `moveEntity()`, `uploadArtifact()`, `deleteArtifact()`
  - Auth tokens stored in `localStorage`, automatically attached to requests

### Stores (Pinia)
- `frontend/src/stores/entities.ts` — main store managing entities state, search, location tree
  - Key methods: `loadFullState()`, `setCurrentEntity()`, `load()`, `listChildLocations()`, `listChildLocationsDeep()`, `readname()`, `selectImages()`
  - WebSocket connection for real-time updates via `connectWS()`
  - Search functionality with filters (`filterworld`, `filterToMissingImage`, `filterToOnlyImage`)

- `frontend/src/stores/camera.ts` — camera capture state management
  - Device orientation detection for mobile
  - Image capture, rotation, and preview handling
  - Full-screen camera view with Teleport

- `frontend/src/stores/clip.ts` — CLIP-based visual search
  - Lazy loading of transformer models from CDN (`Xenova/clip-vit-base-patch32`)
  - Embedding cache in `localStorage` with base64 encoding
  - Concurrent encoding of images from artifacts
  - Merges text search results with visual search scores

- `frontend/src/stores/toasts.ts` — toast notifications
  - Simple message-based toasts with 5-second auto-dismiss
  - Add/remove methods for programmatic control

### Components
- `frontend/src/components/EntityCard.vue` — main entity display/edit component
  - Edit mode for inline entity editing
  - Quick actions: delete, move, edit, capture, create child
  - Artifact image grid with deletion support
  - Move entity dialog with search

- `frontend/src/components/SearchBar.vue` — search interface with filters
  - World-only filter, description filter, CLIP visual search toggle
  - 500ms debounced search input
  - Keyboard shortcuts (Enter to focus, Esc to reset)

- `frontend/src/components/BreadcrumbNav.vue` — hierarchical navigation
  - Renders entity path from World to current entity
  - Click to navigate to any level in the hierarchy

- `frontend/src/components/NewEntityDialog.vue` — entity creation modal
  - Form fields: name, description, quantity, isLabeled checkbox
  - Image upload via file input or camera
  - Auto-assigns free ID or labeled ID based on isLabeled flag

- `frontend/src/components/QuickCaptureCard.vue` — quick camera capture entry point
  - Click-to-capture entity creation at current location
  - Opens camera, uploads artifact, creates child entity

- `frontend/src/components/CameraModal.vue` — camera interface
  - Live viewfinder for capture
  - Preview with rotate/retake/confirm/cancel controls
  - Keyboard shortcuts: Enter (capture/confirm), R (rotate), C (retake), Esc (cancel)

- `frontend/src/components/KbdHint.vue` — keyboard shortcut hints
  - Displays short key names on buttons when showShortcuts is true

### Views
- `frontend/src/views/EntityView.vue` — main entity listing view
  - Breadcrumb navigation
  - Search bar integration
  - Entity grid using `EntityCard` components
  - Filtered entity display based on current search state

- `frontend/src/views/LoginView.vue` — login page (referenced in routes)

- `frontend/src/views/HomeView.vue` — home view (referenced in routes)

### App Root
- `frontend/src/App.vue` — root application component
  - Main layout with header, entity grid, FABs (new, camera)
  - Dialog state management: new entity, move entity, command palette
  - Key handler for keyboard navigation (arrow keys, delete, enter, escape)
  - Grid navigation for keyboard-only entity browsing

### Routing
- `frontend/src/router/index.ts` — Vue Router configuration
  - Hash-based routing (`createWebHashHistory()`)
  - Routes: `/login`, `/:entityId?` (entity view)
  - Auth guard: redirect from login if token exists

### Entity System
- Every `Entity` has:
  - `id`: integer identifier
  - `name`: string name (nullable)
  - `description`: optional description (nullable)
  - `artifacts`: array of `ArtifactID` references (nullable)
  - `location`: parent entity ID (default 0 = World)
  - `metadata`: contains quantity, owners, tags, islabeled, lastModified, lastModifiedBy

- Entity `0` ("World") is the implicit root
- Navigation uses Vue Router hash mode: `router.push({ params: { entityId: id } })`
- URL persistence via hash: `#42` shows entity with ID 42

### Real-time Updates
- WebSocket connection opened to `/ws` endpoint
- Client reconnects automatically after 3 seconds on disconnect
- Server broadcasts `"update"` via `wsHub` whenever `updateStore()` is called
- Client reloads only if `StoreVersion` changed to avoid unnecessary fetches

### Artifact System
- Images uploaded to `/api/artifact` are auto-converted to WebP
- Stored under `./data/artifacts/`
- Artifacts linked to entities via `[]ArtifactID` in `Entity.artifacts`
- CLIP store indexes image artifacts for visual search

### CLIP Integration
- Uses `@huggingface/transformers` from CDN
- Lazy loads `clip-vit-base-patch32` model
- Encodes images into embeddings, stores in `localStorage` with base64 encoding
- Searches artifacts in current entity subtree
- Merges text and visual results with different scoring weights

### Shortcuts Reference
- `N` — Create new entity
- `C` — Open camera capture
- `/` — Focus search
- `Esc` — Reset search
- `Del` — Delete entity
- `M` — Move entity
- `Enter` Edit entity
- `P` — Capture photo to entity
- `⇧C` — Capture photo to new child
- `⇧N` — New child entity
- Arrow keys — Navigate entity grid
- `R` — Rotate (in camera)
- `C` — Retake (in camera)
- `?` — Show shortcuts
- `Esc` — Close dialogs

### Authentication
- JWT-based, optional
- Disabled by default in Docker (`--auth false`)
- POST `/login` issues a 72-hour token
- Token stored in `localStorage`, attached to requests via `Authorization: Bearer` header
- Protected routes check for token, redirect to `/login` on 401
