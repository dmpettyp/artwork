package ui

import (
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky"
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
	// The ImageGraph this layout belongs to (serves as the aggregate ID)
	GraphID imagegraph.ImageGraphID

	// Node positions on the canvas
	NodePositions []NodePosition

	dorky.Entity
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

// SetNodePosition sets the position for a specific node
func (l *Layout) SetNodePosition(nodeID imagegraph.NodeID, x, y float64) error {
	if nodeID.IsNil() {
		return fmt.Errorf("cannot set position for nil NodeID")
	}

	// Find existing position and update it
	for i := range l.NodePositions {
		if l.NodePositions[i].NodeID == nodeID {
			l.NodePositions[i].X = x
			l.NodePositions[i].Y = y
			return nil
		}
	}

	// Not found, add new position
	l.NodePositions = append(l.NodePositions, NodePosition{
		NodeID: nodeID,
		X:      x,
		Y:      y,
	})

	return nil
}

// GetNodePosition retrieves the position for a specific node
func (l *Layout) GetNodePosition(nodeID imagegraph.NodeID) (NodePosition, bool) {
	for _, pos := range l.NodePositions {
		if pos.NodeID == nodeID {
			return pos, true
		}
	}
	return NodePosition{}, false
}

// RemoveNodePosition removes the position for a specific node
func (l *Layout) RemoveNodePosition(nodeID imagegraph.NodeID) {
	for i, pos := range l.NodePositions {
		if pos.NodeID == nodeID {
			// Remove by replacing with last element and truncating
			l.NodePositions[i] = l.NodePositions[len(l.NodePositions)-1]
			l.NodePositions = l.NodePositions[:len(l.NodePositions)-1]
			return
		}
	}
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
