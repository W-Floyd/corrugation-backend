# CLAUDE.md

This file provides guidance to Claude Code when working with the Corrugation frontend.

## Overview

Corrugation frontend is a Vue 3 + TypeScript single-page application for entity management with image capture. Built with Vite, Pinia state management, Vue Router, and TailwindCSS 4. Primary use case: hierarchical organization of records (entities) with photo artifacts.

## Technology Stack

- **Framework**: Vue 3.5.32 (Composition API)
- **Build Tool**: Vite 8
- **Language**: TypeScript 6.0 (noUncheckedIndexedAccess enabled)
- **State Management**: Pinia 3.0
- **Routing**: Vue Router 5.0
- **Styling**: TailwindCSS 4 + PostCSS
- **Camera**: WebRTC (navigator.mediaDevices.getUserMedia)
- **Realtime**: WebSocket for live record updates

## Project Structure

```
frontend/src/
├── api/           # API client, types, record/entity mappings
├── assets/        # Static assets (CSS, images, favicon)
├── components/    # Reusable Vue components
│   ├── icons/    # Vue Material Design Icons
│   └── *.vue     # UI components
├── router/        # Vue Router configuration
├── stores/        # Pinia stores (auth, entities, camera, toasts)
├── utils/         # Utility functions
└── views/         # Page components (login, callback, entity)
```

## Key Stores

### auth.ts
- OAuth2 PKCE flow with Authentik
- Handles token lifecycle (localStorage + httpOnly cookie)
- `fetchConfig()` loads auth config from `/api/auth/config`
- `startLogin()` initiates OAuth redirect
- `handleCallback()` exchanges auth code for token

### entities.ts
- Master store for all records (`allRecords`)
- Entity/record indexing (`entityMap`, `nameMap`, `recordById`)
- Location tree building (`buildLocationTree`)
- Search with image/text embedding support
- WebSocket connection for live updates
- Debounced search triggering API calls

### camera.ts
- WebRTC camera stream management
- Device enumeration and selection
- Capture/rotate/preview workflow
- Device orientation detection (iOS)
- Responsive portrait/landscape handling

### toasts.ts
- Notification system
- `add()` / `update()` / `remove()` / `finalize()`
- Auto-dismiss after 5 seconds (unless persistent)

## Routes

| Path | Component | Auth Required |
|------|-----------|---------------|
| `/login` | LoginView | Optional |
| `/callback` | CallbackView | N/A (OAuth redirect) |
| `/` | EntityView | Required if auth enabled |

## Core Features

### Entity Management
- Hierarchical location tree (parent-child relationships)
- Search by text/embedding with scoring
- Filter by image presence (`filter:missing-image`, `filter:only-image`)
- Create/move/delete operations via API

### Camera Capture
- Fullscreen camera modal overlay
- Live preview via `<video>` element
- Capture → rotate → confirm workflow
- Device selection dropdown
- Keyboard shortcuts (Enter, R, C, Escape)

### API Endpoints
```typescript
GET  /api/v2/records?{params}     // List records
POST /api/v2/record               // Create record
POST /api/v2/record/{id}          // Update record
DELETE /api/v2/record/{id}        // Delete record
POST /api/v2/record/{id}          // Move record (ParentID)
GET  /api/v2/records?search={q}   // Search with embeddings
POST /api/v2/artifact             // Upload image
DELETE /api/v2/artifact/{id}      // Delete artifact
GET  /api/v2/embeddings/search-progress  // Indexing progress
```

## Build Configuration

- **OutDir**: `../dist` (serves from backend)
- **Proxy**: `/api` → `http://localhost:8083`, `/ws` → WebSocket
- **Env**: `DEBUG` defined based on build mode
- **Icons**: Auto-generated via `vite-plugin-favicon-generator`

## TypeScript Setup

- Extends `@vue/tsconfig/tsconfig.dom.json`
- Path alias: `@/` → `src/`
- `noUncheckedIndexedAccess: true` for safer array/object access

## Component Patterns

- **EntityCard**: Displays individual entity with metadata
- **SearchBar**: Debounced search input with filter toggles
- **CameraModal**: Fullscreen camera overlay
- **CommandDialog**: Keyboard command palette
- **ToastContainer**: Notification display area
- **BreadcrumbNav**: Hierarchical navigation

## Dependencies Summary

**Production**:
- `vue`, `vue-router`, `pinia`
- `@mdi/js`, `vue-material-design-icons`
- `drag-drop-touch`

**Development**:
- `vite`, `vue-tsc`, `tailwindcss`
- `vite-plugin-vue-devtools`
- `npm-run-all2` for script orchestration

## Notes for Development

1. WebSocket auto-reconnects on close
2. Auth config fetched once at app startup
3. Search results may be partial during indexing
4. Camera device ID persists in localStorage
5. Toasts auto-dismiss unless `persistent: true`
