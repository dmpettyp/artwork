package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmpettyp/dorky"

	"github.com/dmpettyp/artwork/domain/imagegraph"
)

// ImageGraphRepository implements application.ImageGraphRepository using PostgreSQL
type ImageGraphRepository struct {
	tx       *sql.Tx
	modified map[imagegraph.ImageGraphID]*imagegraph.ImageGraph // Track all modified aggregates
}

// newImageGraphRepository creates a new repository with initialized maps
func newImageGraphRepository(tx *sql.Tx) *ImageGraphRepository {
	return &ImageGraphRepository{
		tx:       tx,
		modified: make(map[imagegraph.ImageGraphID]*imagegraph.ImageGraph),
	}
}

// Get retrieves an ImageGraph by ID with SELECT FOR UPDATE row locking
func (r *ImageGraphRepository) Get(id imagegraph.ImageGraphID) (*imagegraph.ImageGraph, error) {
	// Check if already loaded in this transaction (identity map pattern)
	if ig, ok := r.modified[id]; ok {
		return ig, nil
	}

	ctx := context.Background()

	var row imageGraphRow
	err := r.tx.QueryRowContext(ctx, `
		SELECT id, name, version, data, created_at, updated_at
		FROM image_graphs
		WHERE id = $1
		FOR UPDATE
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

	// Track for event collection and saving
	r.modified[ig.ID] = ig

	return ig, nil
}

// Add inserts a new ImageGraph
func (r *ImageGraphRepository) Add(ig *imagegraph.ImageGraph) error {
	ctx := context.Background()

	row, err := serializeImageGraph(ig)
	if err != nil {
		return fmt.Errorf("failed to serialize image graph: %w", err)
	}

	_, err = r.tx.ExecContext(ctx, `
		INSERT INTO image_graphs (id, name, version, data)
		VALUES ($1, $2, $3, $4)
	`, row.ID, row.Name, row.Version, row.Data)

	if err != nil {
		return fmt.Errorf("failed to insert image graph: %w", err)
	}

	r.modified[ig.ID] = ig

	return nil
}

// SaveAll persists all modified ImageGraphs back to the database
func (r *ImageGraphRepository) SaveAll() error {
	ctx := context.Background()

	for _, ig := range r.modified {
		row, err := serializeImageGraph(ig)
		if err != nil {
			return fmt.Errorf("failed to serialize image graph: %w", err)
		}

		result, err := r.tx.ExecContext(ctx, `
			UPDATE image_graphs
			SET name = $2, version = $3, data = $4, updated_at = NOW()
			WHERE id = $1
		`, row.ID, row.Name, row.Version, row.Data)

		if err != nil {
			return fmt.Errorf("failed to update image graph: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("image graph not found for update: %s", ig.ID.ID)
		}
	}

	return nil
}

// CollectEvents retrieves and clears events from all modified ImageGraphs
func (r *ImageGraphRepository) CollectEvents() []dorky.Event {
	var events []dorky.Event

	for _, ig := range r.modified {
		events = append(events, ig.GetEvents()...)
		ig.ResetEvents()
	}

	return events
}
