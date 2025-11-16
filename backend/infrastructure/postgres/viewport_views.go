package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
)

// ViewportViews provides read-only queries for Viewports
type ViewportViews struct {
	db *sql.DB
}

// Get retrieves a Viewport by graph ID (read-only, no locking)
func (v *ViewportViews) Get(graphID imagegraph.ImageGraphID) (*ui.Viewport, error) {
	ctx := context.Background()

	var row viewportRow
	err := v.db.QueryRowContext(ctx, `
		SELECT graph_id, data, updated_at
		FROM viewports
		WHERE graph_id = $1
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

	return viewport, nil
}
