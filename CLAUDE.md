# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Artwork is an image processing pipeline application that uses a node-based graph system for non-destructive image manipulation. It consists of a Go backend implementing Domain-Driven Design (DDD) principles and a vanilla JavaScript frontend with an SVG-based graph editor.

The core concept is an **ImageGraph**: a directed acyclic graph (DAG) of connected nodes where each node performs a specific image transformation. Images flow through the pipeline from input nodes through transformation nodes to output nodes.

## Quick Start

```bash
# Start Postgres with expected defaults (user: postgres, pass: foofoofoo, db: artwork)
docker run --name artwork-postgres -e POSTGRES_PASSWORD=foofoofoo -e POSTGRES_USER=postgres \
  -e POSTGRES_DB=artwork -p 5432:5432 -d postgres:16

cd backend
make run  # serves API/WS + static frontend on http://localhost:8080
```

Images are stored under `backend/uploads/`. Ensure that directory exists and is writable before starting.

## Build and Run Commands

### Backend (Go)
```bash
# Run the application (from backend directory)
make run

# Build the application
make build

# Run tests
go test ./...

# Run a specific test
go test -run TestName ./path/to/package
```

The backend runs on port 8080 by default. Set `LOG_LEVEL=debug` environment variable for detailed logging.

### Frontend
The frontend is static HTML/CSS/JavaScript served by the Go backend. Simply run the backend and navigate to `http://localhost:8080`.

## Switching Infrastructure

- **Postgres (default):** `cmd/artwork/main.go` wires `postgres.NewUnitOfWork(postgres.NewDB(postgres.DefaultConfig()))`.
- **In-memory:** swap in the commented `inmem.NewUnitOfWork()` block in `cmd/artwork/main.go` and remove Postgres wiring when you want a no-deps local/dev run. Tests already use in-memory repos and mock storage.

## Optional Bootstrap

`cmd/artwork/bootstrap.go` seeds a demo graph. To enable, uncomment the `bootstrap(...)` call near the bottom of `cmd/artwork/main.go`. It expects an image in storage matching the hardcoded ImageID; update the ImageID or upload a matching file before enabling.

## Architecture

### Domain-Driven Design Structure

The codebase follows DDD with clear separation of concerns:

**Domain Layer** (`backend/domain/`):
- `imagegraph/`: Core business logic for the ImageGraph aggregate root
  - `imagegraph.go`: The ImageGraph aggregate that maintains nodes and connections
  - `node.go`: Node entity with inputs, outputs, state, and configuration
  - `node_type.go`: Node type definitions and configuration validation
  - `events.go`: Domain events emitted by the aggregate
  - `node_state.go`: State machine for nodes (Waiting → Generating → Generated)
- `ui/`: UI metadata aggregates (Layout, Viewport) for node positioning
- No dependencies on infrastructure or application layers

**Application Layer** (`backend/application/`):
- Command handlers: Process commands and orchestrate domain operations
- Event handlers: React to domain events, trigger image generation, handle side effects
- `unit_of_work.go`: Transaction boundary for repository operations
- `node_output_setter.go`: Service for setting node outputs via commands

**Infrastructure Layer** (`backend/infrastructure/`):
- `inmem/`: In-memory repositories and unit of work implementation
- `filestorage/`: File system-based image storage
- `imagegen/`: Image generation service that performs actual transformations

**Gateways Layer** (`backend/gateways/`):
- `http/`: HTTP API handlers, WebSocket notifications, serialization

### HTTP & WebSocket Surfaces

- API (under `/api`): node type schemas, list/create graphs, get graph, add/delete/patch nodes, connect/disconnect nodes, upload outputs, layout/viewport get+put, fetch images. Routes are defined in `backend/gateways/http/server.go`.
- WebSocket: `/api/imagegraphs/{id}/ws` streams graph/layout/viewport change notifications for a single graph.

### Event-Driven Architecture

The system uses event sourcing patterns:
1. Domain operations emit events (e.g., `NodeAddedEvent`, `NodeNeedsOutputsEvent`)
2. Events are processed by the message bus (`dorky.MessageBus`)
3. Event handlers trigger side effects (image generation, WebSocket notifications)
4. All state changes propagate through events

### Image Processing Pipeline

**Node State Machine:**
- **Waiting**: Node is waiting for inputs or configuration
- **Generating**: All inputs are ready, image generation is in progress
- **Generated**: Output images have been created

**Image Flow:**
1. Input nodes receive uploaded images
2. When a node has all inputs set, it transitions to Generating state
3. `NodeNeedsOutputsEvent` triggers the ImageGen service
4. ImageGen processes inputs based on node type and configuration
5. Generated images are saved and set as node outputs via `SetNodeOutputImage`
6. Output images propagate downstream to connected nodes
7. When outputs change, downstream nodes reset and regenerate if needed

Sequence in practice:
1. HTTP handler → command → domain change → domain events
2. Message bus fan-out → event handlers
3. Handlers run side effects (imagegen, storage cleanup) and notifier pushes updates over WebSocket

### Node Types

Defined in `backend/domain/imagegraph/node_type.go`:
- **Input**: Upload/provide source images
- **Output**: Terminal nodes with named outputs
- **Crop**: Crop with optional aspect ratio constraints
- **Blur**: Gaussian blur with configurable radius
- **Resize**: Resize to specific dimensions with interpolation options
- **ResizeMatch**: Resize to match another image's dimensions
- **PixelInflate**: Pixel art scaling with grid lines
- **PaletteExtract**: Extract color palette using k-means clustering
- **PaletteApply**: Apply palette to remap image colors

Each node type has:
- Defined inputs and outputs
- Configuration schema with validation
- Optional custom validation logic

### Frontend Architecture

Located in `frontend/`:
- Vanilla JavaScript (no framework)
- SVG-based interactive graph canvas
- WebSocket connection for real-time updates
- Modular structure: `js/api/`, `js/graph/`, `js/ui/`

## Key Domain Concepts

### ImageGraph Operations

The ImageGraph aggregate (`backend/domain/imagegraph/imagegraph.go`) provides these operations:
- `AddNode`: Add a node with type, name, and configuration
- `RemoveNode`: Remove a node and all its connections
- `ConnectNodes`: Create a connection from one node's output to another's input
- `DisconnectNodes`: Remove a connection
- `SetNodeOutputImage`: Set output image and propagate to downstream nodes
- `UnsetNodeOutputImage`: Clear output image
- `SetNodeConfig`: Update node configuration (triggers regeneration)
- `SetNodeName`: Update node name

**Important invariants:**
- Connections cannot create cycles (DAG constraint enforced in `wouldCreateCycle`)
- Each input can only connect to one output
- Outputs can connect to multiple inputs
- Disconnecting/changing upstream nodes triggers downstream regeneration

### Node Configuration

Each node type has a typed config struct in `node_type_config.go` that implements the `NodeConfig` interface:
- `Validate()` - validates the config values
- `NodeType()` - returns the associated node type
- `Schema()` - returns field schema for API responses

The `NodeTypeConfigs` map in `node_type.go` associates each node type with its inputs, outputs, and config factory.

## Development Patterns

### Adding a New Node Type

When adding a new node type, update ALL of the following locations (exhaustiveness tests will fail if you miss any):

1. **Domain - node_type.go** (`backend/domain/imagegraph/node_type.go`):
   - Add the node type constant (e.g., `NodeTypeMyNewType`)
   - Add entry to `NodeTypeConfigs` map with `Inputs`, `Outputs`, `NameRequired`, and `NewConfig` factory

2. **Domain - node_type_config.go** (`backend/domain/imagegraph/node_type_config.go`):
   - Create config struct (e.g., `NodeConfigMyNewType`)
   - Add constructor (e.g., `NewNodeConfigMyNewType()`)
   - Implement `Validate()`, `NodeType()`, and `Schema()` methods

3. **Domain - mappers.go** (`backend/domain/imagegraph/mappers.go`):
   - Add mapping to `NodeTypeMapper` (e.g., `"my_new_type", NodeTypeMyNewType`)

4. **HTTP - serialization.go** (`backend/gateways/http/serialization.go`):
   - Add entry to `nodeTypeMetadata` slice with node type, API name, display name, and category

5. **Image Generation** (`backend/infrastructure/imagegen/imagegen.go`):
   - Implement generation logic in `ImageGen.GenerateNodeOutput()` switch statement

6. **Frontend Schema** (`frontend/js/schemas/`):
   - Add schema file for the new node type (e.g., `my_new_type.js`)

7. **Verify**: Run tests to ensure mapper completeness:
   ```bash
   go test ./...
   ```

The exhaustiveness tests (`TestNodeTypeMapperIsComplete`) will fail if the mapper is incomplete.

### Command/Event Flow

Commands → Domain → Events → Event Handlers → Side Effects

Example: Adding a node
1. HTTP handler receives request → `AddImageGraphNodeCommand`
2. Command handler loads ImageGraph → `ig.AddNode()`
3. Domain emits `NodeAddedEvent`
4. Event handler broadcasts WebSocket notification

### Working with the UnitOfWork

All repository operations must use the UnitOfWork pattern:

```go
return h.uow.Run(ctx, func(repos *Repos) error {
    ig, err := repos.ImageGraphRepository.Get(imageGraphID)
    if err != nil {
        return err
    }

    // Perform domain operations
    err = ig.AddNode(...)

    // Changes are automatically saved when function returns nil
    return nil
})
```

## Testing

- Domain logic tests in `backend/domain/imagegraph/imagegraph_test.go`.
- HTTP handler tests in `backend/gateways/http/http_test.go` (in-memory UoW + mock storage; no Postgres needed).
- `go test ./...` from `backend` is the main entrypoint.
- Use table-driven tests for validation logic.
- Test state transitions and event emission.

## Common Gotchas

1. **Node configurations are maps**: When accessing config values, type assert carefully and handle missing keys
2. **ImageIDs are value objects**: Use `.IsNil()` to check for empty ImageID, not `== nil`
3. **Events must set entity info**: All events need `SetEntity()` called with entity type and ID
4. **State transitions are validated**: Invalid state transitions return errors from the state machine
5. **Preview images are separate from outputs**: Nodes have both preview images (for UI thumbnails) and output images (for pipeline)
6. **Output propagation is automatic**: Setting a node's output image automatically sets it as input for all connected downstream nodes
7. **Uploads directory**: `backend/uploads/` must exist and be writable; missing or read-only will break image saving.
8. **Interpolation names**: Resize/ResizeMatch only accept keys in `resizeInterpolationFunctions` (`NearestNeighbor`, `Bilinear`, `Bicubic`, `MitchellNetravali`, `Lanczos2`, `Lanczos3`).

## Code Style

- **Avoid unnecessary comments**: Don't add comments that merely restate what the code does
- **Self-documenting code**: Use clear variable and function names instead of comments
- **Comment only when needed**: Add comments for non-obvious logic, important invariants, gotchas, or complex business rules
- **Keep it clean**: Code should be straightforward and speak for itself

## Image Storage

Images are stored in the `backend/uploads/` directory with ImageID as the filename. The system supports PNG and JPEG formats. Images are automatically cleaned up when no longer referenced by any node output.
