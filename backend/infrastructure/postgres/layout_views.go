package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
)

// LayoutViews provides read-only queries for Layouts
type LayoutViews struct {
	db *sql.DB
}

func NewLayoutViews(db *sql.DB) *LayoutViews {
	return &LayoutViews{db: db}
}

// Get retrieves a Layout by graph ID (read-only, no locking)
func (v *LayoutViews) Get(ctx context.Context, graphID imagegraph.ImageGraphID) (*ui.Layout, error) {
	var row layoutRow
	err := v.db.QueryRowContext(ctx, `
		SELECT graph_id, data, updated_at
		FROM layouts
		WHERE graph_id = $1
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

	return layout, nil
}
