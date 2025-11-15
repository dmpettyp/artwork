package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
)

// LayoutRepository implements application.LayoutRepository using PostgreSQL
type LayoutRepository struct {
	tx       *sql.Tx
	modified map[imagegraph.ImageGraphID]*ui.Layout // Track modified aggregates for event collection
}

// newLayoutRepository creates a new repository with initialized maps
func newLayoutRepository(tx *sql.Tx) *LayoutRepository {
	return &LayoutRepository{
		tx:       tx,
		modified: make(map[imagegraph.ImageGraphID]*ui.Layout),
	}
}

// Get retrieves a Layout by graph ID with SELECT FOR UPDATE row locking
func (r *LayoutRepository) Get(graphID imagegraph.ImageGraphID) (*ui.Layout, error) {
	// Check if already loaded in this transaction (identity map pattern)
	if layout, ok := r.modified[graphID]; ok {
		return layout, nil
	}

	ctx := context.Background()

	var row layoutRow
	err := r.tx.QueryRowContext(ctx, `
		SELECT graph_id, data, updated_at
		FROM layouts
		WHERE graph_id = $1
		FOR UPDATE
	`, graphID.ID).Scan(
		&row.GraphID,
		&row.Data,
		&row.UpdatedAt,
	)

	if err != nil {
		return nil, wrapLayoutNotFound(err)
	}

	layout, err := deserializeLayout(row)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize layout: %w", err)
	}

	// Track for event collection
	r.modified[layout.GraphID] = layout

	return layout, nil
}

// Add inserts a new Layout
func (r *LayoutRepository) Add(layout *ui.Layout) error {
	ctx := context.Background()

	row, err := serializeLayout(layout)
	if err != nil {
		return fmt.Errorf("failed to serialize layout: %w", err)
	}

	_, err = r.tx.ExecContext(ctx, `
		INSERT INTO layouts (graph_id, data)
		VALUES ($1, $2)
		ON CONFLICT (graph_id) DO UPDATE
		SET data = EXCLUDED.data, updated_at = NOW()
	`, row.GraphID, row.Data)

	if err != nil {
		return fmt.Errorf("failed to insert/update layout: %w", err)
	}

	// Track for event collection
	r.modified[layout.GraphID] = layout

	return nil
}

// save persists a Layout using UPSERT (called by UnitOfWork on commit)
func (r *LayoutRepository) save(layout *ui.Layout) error {
	ctx := context.Background()

	row, err := serializeLayout(layout)
	if err != nil {
		return fmt.Errorf("failed to serialize layout: %w", err)
	}

	_, err = r.tx.ExecContext(ctx, `
		INSERT INTO layouts (graph_id, data)
		VALUES ($1, $2)
		ON CONFLICT (graph_id) DO UPDATE
		SET data = EXCLUDED.data, updated_at = NOW()
	`, row.GraphID, row.Data)

	if err != nil {
		return fmt.Errorf("failed to save layout: %w", err)
	}

	return nil
}
