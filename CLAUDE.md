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
docker-compose up
```

## Architecture

**Backend:** Go + Echo v4 HTTP framework. All routes are registered in `cmd/root.go`. Business logic lives in `cmd/entity.go`, `cmd/store.go`, `cmd/artifacts.go`, and `cmd/auth.go`.

**Storage:** A single `Store` struct (entities + artifacts + version counter) is serialized to disk via diskv. There is no database — reads and writes go through `loadStore()` / `updateStore()` in `cmd/store.go`. `StoreVersion` is incremented on every write; `updateStore()` also calls `hub.broadcast()` to push change notifications to all connected WebSocket clients.

**Frontend:** Single-page app using Alpine.js v3 stores and Tailwind CSS (both via CDN). All JS lives in `assets/scripts.js`. HTML templates are rendered server-side with Go's `text/template` via goview, assembled from partials under `views/layouts/`. Static assets are served from `assets/`.

**Entity hierarchy:** Everything is an `Entity` with an integer `ID` and a `location` field pointing to its parent entity ID. Entity `0` ("World") is the implicit root. Navigation is purely client-side: `$store.entities.currentEntity` holds the active entity ID, and `setCurrentEntity(id)` updates it, syncs the URL hash, and re-renders children.

**URL persistence:** The current entity ID is stored in the URL hash (`#42`). On load, `init()` reads the hash and navigates directly to that entity.

**Real-time updates:** The frontend opens a WebSocket to `/ws` (`connectWS()` in `assets/scripts.js`). The server broadcasts `"update"` to all clients via `wsHub` (`cmd/ws.go`) whenever `updateStore()` is called. On receiving a message the client calls `reload()`, which re-fetches the store only if `StoreVersion` changed. The client reconnects automatically after 3 s on disconnect.

**Artifact system:** Images uploaded to an entity are auto-converted to WebP and stored under `./data/artifacts/`. Artifacts are linked to entities via `[]ArtifactID` on the entity struct.

**Icon/component system:** SVG icons and reusable button HTML are generated in `cmd/components.go` and exposed as template functions (e.g. `componentButtonRound`).

**Auth:** JWT-based, optional. Disabled by default in Docker (`--auth false`). POST `/login` issues a 72-hour token; protected routes use Echo's JWT middleware.
