package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmpettyp/artwork/backend/domain/imagegraph"
	"github.com/dmpettyp/artwork/backend/domain/ui"
)

// LayoutRepository implements application.LayoutRepository using PostgreSQL
type LayoutRepository struct {
	tx       *sql.Tx
	modified []*ui.Layout // Track modified aggregates for event collection
}

// Get retrieves a Layout by graph ID with SELECT FOR UPDATE row locking
func (r *LayoutRepository) Get(graphID imagegraph.ImageGraphID) (*ui.Layout, error) {
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
	r.modified = append(r.modified, layout)

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
	r.modified = append(r.modified, layout)

	return nil
}
