# Vue 3 + TypeScript Migration Plan

Migrate the frontend from Alpine.js (CDN, no build step) to a Vue 3 + TypeScript SPA built with Vite. The Go backend and all `/api/*` routes are unchanged.

---

## M1 — Project setup

- [ ] Scaffold `frontend/` with `npm create vue@latest` (TypeScript, Vue Router, Pinia)
- [ ] Configure Tailwind CSS via PostCSS in Vite; remove CDN links from templates
- [ ] Configure `vite.config.ts`: output to `../dist`, proxy `/api` and `/ws` to `localhost:8083`
- [ ] Add `npm run build` to `push.sh` before the Docker build
- [ ] Add catch-all route in `cmd/root.go` serving `dist/index.html`; serve `dist/assets/` statically
- [ ] Remove goview wiring from `cmd/root.go` and `go.mod`
- [ ] Verify: `npm run dev` loads blank page, proxied `/api/store` returns data

---

## M2 — TypeScript types

Define shared types before writing any components or stores.

- [ ] Create `src/api/types.ts` with:
  - `EntityMetadata` — quantity, owners, tags, islabeled
  - `Entity` — id, name, description, location, artifacts, metadata
  - `Artifact` — id, path, image
  - `FullState` — entities record, artifacts record, storeversion
- [ ] Verify types match actual API responses from `/api/store`

---

## M3 — API module

Replace all XHR calls (currently in the `api` Alpine store, ~236 lines) with a typed `fetch`-based module. The synchronous XHR sequencing (upload artifact → create entity) becomes `await` chains.

- [ ] Create `src/api/index.ts` with functions:
  - `getFullState(): Promise<FullState>`
  - `createEntity(entity: Partial<Entity>): Promise<number>`
  - `updateEntity(id: number, patch: Partial<Entity>): Promise<void>`
  - `deleteEntity(id: number): Promise<void>`
  - `moveEntity(id: number, location: number): Promise<void>`
  - `uploadArtifact(file: File): Promise<number>`
  - `deleteArtifact(id: number): Promise<void>`
  - `firstFreeId(): Promise<number>`
  - `firstAvailableId(): Promise<number>`

---

## M4 — Pinia stores

Migrate each Alpine store. Stores with global lifecycle (entities, camera, clip) become Pinia stores. Dialog state that is tightly scoped becomes component-local `ref`s.

| Alpine store | Target |
|---|---|
| `entities` (~245 lines) | `useEntitiesStore` — fullstate, load/reload, search, location tree, WebSocket reconnect |
| `camera` (~253 lines) | `useCameraStore` — stream, capture, rotate, confirm, callback |
| `clip` (~195 lines) | `useClipStore` — model loading, encode, search, score map |
| `toasts` (~15 lines) | `useToastsStore` |
| `newEntityDialog` | Component-local state in `NewEntityDialog.vue` |
| `editEntityDialog` | Component-local state in `EntityCard.vue` |
| `moveEntityDialog` | Component-local state in `MoveEntityDialog.vue` |
| `api` | Deleted — replaced by `src/api/index.ts` and store actions |
| `isLoading` | Derived from `useEntitiesStore` |

Tasks:
- [ ] `useEntitiesStore`: fullstate fetching, `load()`, `reload()`, `hasChildren()`, `selectImages()`, `readname()`, `listChildLocations()`, search/filter logic, WebSocket connection
- [ ] `useCameraStore`: port camera logic (~253 lines) including orientation tracking, capture, rotate, retake, confirm, `open(callback)` interface
- [ ] `useClipStore`: port CLIP logic including lazy model load, per-entity encoding, text search, score merging
- [ ] `useToastsStore`: trivial port

---

## M5 — Router

- [ ] Install Vue Router, configure hash mode to preserve existing `#42` URL format
- [ ] Define route: `/:entityId(\\d+)?` → `EntityView.vue`
- [ ] Replace `setCurrentEntity(id)` calls with `router.push({ params: { entityId: id } })`
- [ ] On app mount, read `route.params.entityId` to set initial entity

---

## M6 — Components

Build bottom-up so each component is testable before the next depends on it.

### 6a — `CameraModal.vue`
Port `camera.html` (53 lines) + `useCameraStore`.
- [ ] Full-screen overlay using `<Teleport to="body">`
- [ ] Live viewfinder, capture button, preview, rotate, retake, confirm ("Use") buttons
- [ ] Orientation-aware button rotation
- [ ] No entity dependencies — driven entirely by `useCameraStore.open(callback)`

### 6b — `EntityCard.vue`
Port `card.html` (122 lines).
- [ ] View mode: name, id, quantity, description, child list, images, action buttons
- [ ] Edit mode: inline inputs, image deletion, camera quick-capture, save, cancel
- [ ] Quick-capture button (view mode) calls `api.uploadArtifact` + `api.updateEntity` directly
- [ ] CLIP score badge and text-match badge
- [ ] Emit nothing — mutations go directly through API module + store invalidation

### 6c — `NewEntityDialog.vue`
Port `newEntity.html` (86 lines).
- [ ] Props: `visible`, `location`
- [ ] File input + camera button (opens `useCameraStore`)
- [ ] ID preview (firstFreeId / firstAvailableId based on islabeled)
- [ ] Submit calls `api.createEntity` then emits `created`

### 6d — `MoveEntityDialog.vue`
Port `moveEntity.html` (66 lines).
- [ ] Searchable entity selector
- [ ] Calls `api.moveEntity` on confirm

### 6e — `BreadcrumbNav.vue`
- [ ] Reads `locationTree` from `useEntitiesStore`
- [ ] Each crumb calls `router.push`

### 6f — `SearchBar.vue`
- [ ] Debounced input bound to `useEntitiesStore.searchtext`
- [ ] Checkboxes for filterworld, searchdescription, CLIP enabled
- [ ] CLIP loading spinner

### 6g — `QuickCaptureCard.vue`
Port the blank camera card from `body.html`.
- [ ] Dashed border placeholder card with camera icon
- [ ] On click: upload artifact → create entity at current location → reload

### 6h — `ToastContainer.vue`
Port toast markup from `body.html`.
- [ ] Reads from `useToastsStore`

### 6i — `EntityView.vue` + `App.vue`
Assemble everything.
- [ ] `EntityView.vue`: entity grid, quick-capture card, empty state, search bar, breadcrumb
- [ ] `App.vue`: router-view, `CameraModal`, `NewEntityDialog`, `MoveEntityDialog`, `ToastContainer`

---

## M7 — Cleanup

- [ ] Delete `views/` directory
- [ ] Delete `assets/scripts.js`
- [ ] Delete `components/buttonRound.html`
- [ ] Delete `cmd/components.go`
- [ ] Remove `foolin/goview` from `go.mod` / `go.sum`
- [ ] Remove Alpine.js and Tailwind CDN links (already gone after M1)
- [ ] Update `Dockerfile` to include `frontend/` build in the image build

---

## Known risks

**Async sequencing** — The current `quickCapture` and `newEntity` flows rely on synchronous XHR to guarantee artifact upload completes before entity creation. The async rewrite must preserve this ordering with explicit `await`.

**CLIP worker boundary** — If the CLIP model runs in a Web Worker, Vue reactivity won't cross that boundary. The store will need explicit `postMessage` / `onmessage` wiring, the same as Alpine does today.

**Camera modal stacking** — Use `<Teleport to="body">` to avoid z-index issues from component tree stacking contexts.

**`foolin/goview` removal** — The template functions `componentButtonRound` and `unescapeHTML` are used throughout the current templates. Once removed from Go, ensure no references remain in any served HTML.
