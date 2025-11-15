package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmpettyp/dorky"

	"github.com/dmpettyp/artwork/backend/domain/imagegraph"
	"github.com/dmpettyp/artwork/backend/domain/ui"
)

// ViewportRepository implements application.ViewportRepository using PostgreSQL
type ViewportRepository struct {
	tx       *sql.Tx
	modified map[imagegraph.ImageGraphID]*ui.Viewport // Track modified aggregates for event collection
}

// newViewportRepository creates a new repository with initialized maps
func newViewportRepository(tx *sql.Tx) *ViewportRepository {
	return &ViewportRepository{
		tx:       tx,
		modified: make(map[imagegraph.ImageGraphID]*ui.Viewport),
	}
}

// Get retrieves a Viewport by graph ID with SELECT FOR UPDATE row locking
func (r *ViewportRepository) Get(graphID imagegraph.ImageGraphID) (*ui.Viewport, error) {
	// Check if already loaded in this transaction (identity map pattern)
	if viewport, ok := r.modified[graphID]; ok {
		return viewport, nil
	}

	ctx := context.Background()

	var row viewportRow
	err := r.tx.QueryRowContext(ctx, `
		SELECT graph_id, data, updated_at
		FROM viewports
		WHERE graph_id = $1
		FOR UPDATE
	`, graphID.ID).Scan(
		&row.GraphID,
		&row.Data,
		&row.UpdatedAt,
	)

	if err != nil {
		return nil, wrapViewportNotFound(err)
	}

	viewport, err := deserializeViewport(row)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize viewport: %w", err)
	}

	// Track for event collection
	r.modified[viewport.GraphID] = viewport

	return viewport, nil
}

// Add inserts a new Viewport
func (r *ViewportRepository) Add(viewport *ui.Viewport) error {
	ctx := context.Background()

	row, err := serializeViewport(viewport)
	if err != nil {
		return fmt.Errorf("failed to serialize viewport: %w", err)
	}

	_, err = r.tx.ExecContext(ctx, `
		INSERT INTO viewports (graph_id, data)
		VALUES ($1, $2)
		ON CONFLICT (graph_id) DO UPDATE
		SET data = EXCLUDED.data, updated_at = NOW()
	`, row.GraphID, row.Data)

	if err != nil {
		return fmt.Errorf("failed to insert/update viewport: %w", err)
	}

	// Track for event collection
	r.modified[viewport.GraphID] = viewport

	return nil
}

// SaveAll persists all modified Viewports back to the database
func (r *ViewportRepository) SaveAll() error {
	ctx := context.Background()

	for _, viewport := range r.modified {
		row, err := serializeViewport(viewport)
		if err != nil {
			return fmt.Errorf("failed to serialize viewport: %w", err)
		}

		_, err = r.tx.ExecContext(ctx, `
			INSERT INTO viewports (graph_id, data)
			VALUES ($1, $2)
			ON CONFLICT (graph_id) DO UPDATE
			SET data = EXCLUDED.data, updated_at = NOW()
		`, row.GraphID, row.Data)

		if err != nil {
			return fmt.Errorf("failed to save viewport: %w", err)
		}
	}

	return nil
}

// CollectEvents retrieves and clears events from all modified Viewports
func (r *ViewportRepository) CollectEvents() []dorky.Event {
	var events []dorky.Event

	for _, viewport := range r.modified {
		events = append(events, viewport.GetEvents()...)
		viewport.ResetEvents()
	}

	return events
}
