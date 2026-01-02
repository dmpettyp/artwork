# Project Summary (Artwork)

Artwork is an image-processing pipeline app built around an ImageGraph
(a DAG of nodes). Each node transforms images, passes outputs to downstream
nodes, and can expose previews. The backend is Go (DDD style), the frontend
is vanilla JS with an SVG graph editor, and the backend serves the UI and API.

## Quick Start

- Postgres (optional):
  - docker run --name artwork-postgres -e POSTGRES_PASSWORD=foofoofoo -e POSTGRES_USER=postgres -e POSTGRES_DB=artwork -p 5432:5432 -d postgres:16
- Run backend:
  - cd backend
  - go run ./cmd/artwork -store=postgres
  - or use -store=inmem for no DB
  - optional demo graph: -bootstrap
- UI: open http://localhost:8080
- Images: stored under backend/uploads/ (must exist and be writable)

## Repository Map

- backend/
  - cmd/artwork/         app entrypoint, flags, optional bootstrap
  - domain/              core ImageGraph model + UI metadata
  - application/         command/event handlers, unit of work, output setting
  - infrastructure/      image generation, storage, in-memory repos
  - gateways/http/       HTTP + WebSocket API, serialization
- frontend/
  - index.html, css/
  - js/                  app state, graph editor, modals, schema usage
- doc/                   palettes and notes
- scripts/               utility scripts (if any)

## Backend Architecture

DDD layers:
- Domain: backend/domain/imagegraph
  - ImageGraph aggregate, Node entity, typed configs, events
  - Enforces DAG constraints and input/output invariants
- Application: backend/application
  - Commands, handlers, message bus orchestration
  - Event handlers trigger image generation and WebSocket updates
- Infrastructure: backend/infrastructure
  - imagegen: transforms images by node type
  - filestorage + inmem repos
- Gateways: backend/gateways/http
  - REST API + WS notifications

UnitOfWork pattern:
- All state changes are wrapped in a UoW transaction.
- Changes are persisted when the UoW function returns nil.

## Core Concepts

ImageGraph:
- A graph of nodes with inputs/outputs and typed configs.
- Operations: add/remove nodes, connect/disconnect, set outputs, set preview,
  update config/name, set layout/viewport.

Node types:
- Input, Output, Crop, Blur, Resize, ResizeMatch, PixelInflate,
  PaletteExtract, PaletteApply.
- Each node type defines inputs, outputs, and a typed config schema.
- /api/node-types is the frontend source of truth for config shapes.

Image versioning:
- Each node tracks ImageVersion for preview/outputs.
- Writes must include node_version; stale writes are ignored.
- Events include image_version and generation logs include node_version.

## Data and Event Flow

Typical flow (create/update):
1. HTTP handler receives request and builds a command.
2. Command handler loads ImageGraph via UnitOfWork.
3. Domain operation emits events.
4. Message bus dispatches events to handlers.
5. Side effects run (image generation, storage updates, WS notify).

Image generation flow:
- NodeNeedsOutputsEvent triggers imagegen.
- Imagegen saves preview/output images, then sets them on the node via
  commands that carry node_version.
- Outputs propagate to downstream nodes; state updates push over WS.

WebSocket:
- /api/imagegraphs/{id}/ws sends graph/layout/viewport updates in real time.

## HTTP API (high level)

- GET /api/node-types
- GET/POST /api/imagegraphs
- GET /api/imagegraphs/{id}
- POST /api/imagegraphs/{id}/nodes
- PATCH /api/imagegraphs/{id}/nodes/{node_id}
- DELETE /api/imagegraphs/{id}/nodes/{node_id}
- PUT /api/imagegraphs/{id}/connectNodes
- PUT /api/imagegraphs/{id}/disconnectNodes
- PUT /api/imagegraphs/{id}/nodes/{node_id}/outputs/{output_name} (multipart)
- GET /api/images/{image_id}
- GET/PUT /api/imagegraphs/{id}/layout
- GET/PUT /api/imagegraphs/{id}/viewport

## Frontend Architecture

- Vanilla JS + SVG graph editor.
- State is hydrated from API and kept in sync via WebSocket updates.
- Uses /api/node-types to build config forms and validate inputs.
- Modals handle editing, image previews, and node JSON viewing.

Key folders:
- frontend/js/graph: graph model + rendering
- frontend/js/api: API calls
- frontend/js/modals: modal controllers
- frontend/js/form-builder.js: node config UI

## Adding a New Node Type (Checklist)

Update all:
- backend/domain/imagegraph/node_type.go
- backend/domain/imagegraph/node_type_config.go
- backend/domain/imagegraph/mappers.go
- backend/gateways/http/serialization.go (metadata)
- backend/application/node_output_generators.go
- backend/infrastructure/imagegen/ (implementation)
- frontend/js/schemas/ (schema file)

Run: go test ./... (from backend)

## Testing

- go test ./... (backend)
- Domain tests: backend/domain/imagegraph/imagegraph_test.go
- HTTP tests: backend/gateways/http/http_test.go

## Common Gotchas

- backend/uploads/ must exist and be writable.
- Use -store=inmem if Postgres is not available.
- Typed configs reject invalid JSON; rely on /api/node-types schemas.
- Resize interpolation must use supported names.
- Image versioning requires node_version on preview/output writes.
