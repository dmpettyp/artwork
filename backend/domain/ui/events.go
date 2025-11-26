package ui

import (
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky/messages"
)

// LayoutEvent is the base event for Layout domain events
type LayoutEvent struct {
	messages.BaseEvent
	GraphID imagegraph.ImageGraphID
}

// LayoutUpdatedEvent is emitted when node positions are updated
type LayoutUpdatedEvent struct {
	LayoutEvent
	NodePositions []NodePosition
}

func NewLayoutUpdatedEvent(layout *Layout) *LayoutUpdatedEvent {
	e := &LayoutUpdatedEvent{
		LayoutEvent: LayoutEvent{
			GraphID: layout.GraphID,
		},
		NodePositions: append([]NodePosition{}, layout.NodePositions...),
	}
	e.Init("LayoutUpdated")
	return e
}

// ViewportEvent is the base event for Viewport domain events
type ViewportEvent struct {
	messages.BaseEvent
	GraphID imagegraph.ImageGraphID
}

// ViewportUpdatedEvent is emitted when viewport is updated
type ViewportUpdatedEvent struct {
	ViewportEvent
	Zoom float64
	PanX float64
	PanY float64
}

func NewViewportUpdatedEvent(viewport *Viewport) *ViewportUpdatedEvent {
	e := &ViewportUpdatedEvent{
		ViewportEvent: ViewportEvent{
			GraphID: viewport.GraphID,
		},
		Zoom: viewport.Zoom,
		PanX: viewport.PanX,
		PanY: viewport.PanY,
	}
	e.Init("ViewportUpdated")
	return e
}
