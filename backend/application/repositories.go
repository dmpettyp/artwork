package application

import (
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
)

type Repos struct {
	ImageGraphRepository ImageGraphRepository
	LayoutRepository     LayoutRepository
	ViewportRepository   ViewportRepository
}

type ImageGraphRepository interface {
	Add(*imagegraph.ImageGraph) error
	Get(imagegraph.ImageGraphID) (*imagegraph.ImageGraph, error)
}

type LayoutRepository interface {
	Get(graphID imagegraph.ImageGraphID) (*ui.Layout, error)
	Add(layout *ui.Layout) error
}

type ViewportRepository interface {
	Get(graphID imagegraph.ImageGraphID) (*ui.Viewport, error)
	Add(viewport *ui.Viewport) error
}
