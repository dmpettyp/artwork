package inmem

import (
	"context"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
)

// ViewportViews implements application.ViewportViews using the viewport repository
type ViewportViews struct {
	repo *ViewportRepository
}

// NewViewportViews creates a new viewport views instance
func NewViewportViews(repo *ViewportRepository) *ViewportViews {
	return &ViewportViews{
		repo: repo,
	}
}

// Get retrieves a viewport by graph ID
func (v *ViewportViews) Get(ctx context.Context, graphID imagegraph.ImageGraphID) (*ui.Viewport, error) {
	return v.repo.Get(graphID)
}
