package ui

import (
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky/aggregate"
)

// NodePosition represents the 2D position of a node on the canvas
type NodePosition struct {
	NodeID imagegraph.NodeID
	X      float64
	Y      float64
}

// Layout represents the node positioning layout for an ImageGraph
// This is an aggregate root identified by GraphID
type Layout struct {
	aggregate.Aggregate

	// The ImageGraph this layout belongs to (serves as the aggregate ID)
	GraphID imagegraph.ImageGraphID

	// Node positions on the canvas
	NodePositions []NodePosition
}

// NewLayout creates a new Layout with empty node positions
func NewLayout(
	graphID imagegraph.ImageGraphID,
) (*Layout, error) {
	if graphID.IsNil() {
		return nil, fmt.Errorf("cannot create Layout with nil GraphID")
	}

	return &Layout{
		GraphID:       graphID,
		NodePositions: []NodePosition{},
	}, nil
}

// SetNodePositions replaces all node positions and emits a LayoutUpdatedEvent
func (l *Layout) SetNodePositions(nodePositions []NodePosition) {
	l.NodePositions = nodePositions
	l.AddEvent(NewLayoutUpdatedEvent(l))
}

// Clone creates a deep copy of the Layout
func (l *Layout) Clone() *Layout {
	clone := &Layout{
		GraphID:       l.GraphID,
		NodePositions: make([]NodePosition, len(l.NodePositions)),
	}

	copy(clone.NodePositions, l.NodePositions)

	return clone
}
