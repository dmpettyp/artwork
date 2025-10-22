package inmem

import (
	"context"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
)

// LayoutViews implements application.LayoutViews using the layout repository
type LayoutViews struct {
	repo *LayoutRepository
}

// NewLayoutViews creates a new layout views instance
func NewLayoutViews(repo *LayoutRepository) *LayoutViews {
	return &LayoutViews{
		repo: repo,
	}
}

// Get retrieves a layout by graph ID
func (v *LayoutViews) Get(ctx context.Context, graphID imagegraph.ImageGraphID) (*ui.Layout, error) {
	return v.repo.Get(graphID)
}
