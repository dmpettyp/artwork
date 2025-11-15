-- Initial schema for Artwork application using JSONB for aggregate storage

-- Image graphs table - stores complete ImageGraph aggregate as JSONB
CREATE TABLE image_graphs (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    data JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for listing/searching graphs by name
CREATE INDEX idx_image_graphs_name ON image_graphs(name);

-- Layouts table - stores UI layout data as JSONB
CREATE TABLE layouts (
    graph_id UUID PRIMARY KEY REFERENCES image_graphs(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Viewports table - stores UI viewport data as JSONB
CREATE TABLE viewports (
    graph_id UUID PRIMARY KEY REFERENCES image_graphs(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Events table - stores domain events for audit trail and debugging
CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB NOT NULL,
    aggregate_version BIGINT,
    timestamp TIMESTAMP NOT NULL
);

-- Indexes for querying events
CREATE INDEX idx_events_aggregate ON events(aggregate_id, timestamp);
CREATE INDEX idx_events_type ON events(event_type);
CREATE INDEX idx_events_timestamp ON events(timestamp DESC);
