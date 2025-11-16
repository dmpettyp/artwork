package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
)

// ImageGraphViews provides read-only queries for ImageGraphs
type ImageGraphViews struct {
	db *sql.DB
}

// Get retrieves an ImageGraph by ID (read-only, no locking)
func (v *ImageGraphViews) Get(id imagegraph.ImageGraphID) (*imagegraph.ImageGraph, error) {
	ctx := context.Background()

	var row imageGraphRow
	err := v.db.QueryRowContext(ctx, `
		SELECT id, name, version, data, created_at, updated_at
		FROM image_graphs
		WHERE id = $1
	`, id.ID).Scan(
		&row.ID,
		&row.Name,
		&row.Version,
		&row.Data,
		&row.CreatedAt,
		&row.UpdatedAt,
	)

	if err != nil {
		return nil, wrapImageGraphNotFound(err)
	}

	ig, err := deserializeImageGraph(row)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize image graph: %w", err)
	}

	return ig, nil
}

// List retrieves all ImageGraphs (read-only)
func (v *ImageGraphViews) List() ([]*imagegraph.ImageGraph, error) {
	ctx := context.Background()

	rows, err := v.db.QueryContext(ctx, `
		SELECT id, name, version, data, created_at, updated_at
		FROM image_graphs
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query image graphs: %w", err)
	}
	defer rows.Close()

	var graphs []*imagegraph.ImageGraph
	for rows.Next() {
		var row imageGraphRow
		if err := rows.Scan(
			&row.ID,
			&row.Name,
			&row.Version,
			&row.Data,
			&row.CreatedAt,
			&row.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan image graph row: %w", err)
		}

		ig, err := deserializeImageGraph(row)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize image graph: %w", err)
		}

		graphs = append(graphs, ig)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating image graph rows: %w", err)
	}

	return graphs, nil
}
