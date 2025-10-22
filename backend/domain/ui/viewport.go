package ui

import (
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky"
)

// Viewport represents the canvas viewport state (zoom and pan) for an ImageGraph
// This is an aggregate root identified by GraphID
type Viewport struct {
	// The ImageGraph this viewport belongs to (serves as the aggregate ID)
	GraphID imagegraph.ImageGraphID

	// Zoom level (must be > 0)
	Zoom float64

	// Pan offset X
	PanX float64

	// Pan offset Y
	PanY float64

	dorky.Entity
}

// NewViewport creates a new Viewport with default settings
func NewViewport(
	graphID imagegraph.ImageGraphID,
) (*Viewport, error) {
	if graphID.IsNil() {
		return nil, fmt.Errorf("cannot create Viewport with nil GraphID")
	}

	return &Viewport{
		GraphID: graphID,
		Zoom:    1.0,
		PanX:    0,
		PanY:    0,
	}, nil
}

// Set updates all viewport properties at once and emits a ViewportUpdatedEvent
func (v *Viewport) Set(zoom, panX, panY float64) error {
	if zoom <= 0 {
		return fmt.Errorf("zoom must be greater than 0, got %f", zoom)
	}

	v.Zoom = zoom
	v.PanX = panX
	v.PanY = panY

	event := NewViewportUpdatedEvent(v)
	event.SetEntity("Viewport", v.GraphID.ID)
	v.AddEvent(event)

	return nil
}

// Clone creates a copy of the Viewport
func (v *Viewport) Clone() *Viewport {
	return &Viewport{
		GraphID: v.GraphID,
		Zoom:    v.Zoom,
		PanX:    v.PanX,
		PanY:    v.PanY,
	}
}
