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

**Frontend structure:**
- `frontend/src/api/types.ts` — shared TypeScript types (`Entity`, `Artifact`, `FullState`, etc.)
- `frontend/src/api/index.ts` — typed fetch-based API client
- `frontend/src/stores/` — Pinia stores: `entities.ts`, `camera.ts`, `clip.ts`, `toasts.ts`
- `frontend/src/components/` — Vue components: `EntityCard.vue`, `CameraModal.vue`, `NewEntityDialog.vue`, `MoveEntityDialog.vue`, `SearchBar.vue`, `BreadcrumbNav.vue`, `QuickCaptureCard.vue`, `ToastContainer.vue`
- `frontend/src/views/EntityView.vue` — main view, assembled from components above
- `frontend/src/App.vue` — root component with router-view, modals, toasts

**Entity hierarchy:** Everything is an `Entity` with an integer `ID` and a `location` field pointing to its parent entity ID. Entity `0` ("World") is the implicit root. Navigation uses Vue Router hash mode: `router.push({ params: { entityId: id } })`. On mount, `route.params.entityId` sets the initial entity.

**URL persistence:** The current entity ID is stored in the URL hash as `/#/<entityId>` (e.g. `/#/42`). Vue Router hash mode preserves this format.

**Real-time updates:** The frontend opens a WebSocket to `/ws` (in `useEntitiesStore`). The server broadcasts `"update"` to all clients via `wsHub` (`cmd/ws.go`) whenever `updateStore()` is called. On receiving a message the client calls `reload()`, which re-fetches the store only if `StoreVersion` changed. The client reconnects automatically after 3 s on disconnect.

**Artifact system:** Images uploaded to an entity are auto-converted to WebP and stored under `./data/artifacts/`. Artifacts are linked to entities via `[]ArtifactID` on the entity struct.

**Auth:** JWT-based, optional. Disabled by default in Docker (`--auth false`). POST `/login` issues a 72-hour token; protected routes use Echo's JWT middleware.
