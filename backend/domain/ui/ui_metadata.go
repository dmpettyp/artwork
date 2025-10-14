package ui

import (
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
)

// NodePosition represents the UI position of a node in the canvas
type NodePosition struct {
	NodeID imagegraph.NodeID `json:"node_id"`
	X      float64           `json:"x"`
	Y      float64           `json:"y"`
}

// Viewport represents the canvas viewport state (zoom and pan)
type Viewport struct {
	Zoom float64 `json:"zoom"`
	PanX float64 `json:"pan_x"`
	PanY float64 `json:"pan_y"`
}

// UIMetadata represents UI-specific state for an ImageGraph
type UIMetadata struct {
	// The ImageGraph this metadata belongs to (serves as the ID)
	GraphID imagegraph.ImageGraphID

	// Canvas viewport state
	Viewport Viewport

	// Node positions
	NodePositions []NodePosition
}

// NewUIMetadata creates a new UIMetadata with default viewport settings
func NewUIMetadata(
	graphID imagegraph.ImageGraphID,
) (*UIMetadata, error) {
	if graphID.IsNil() {
		return nil, fmt.Errorf("cannot create UIMetadata with nil GraphID")
	}

	return &UIMetadata{
		GraphID: graphID,
		Viewport: Viewport{
			Zoom: 1.0,
			PanX: 0,
			PanY: 0,
		},
		NodePositions: []NodePosition{},
	}, nil
}

// SetViewport updates the viewport state
func (m *UIMetadata) SetViewport(zoom, panX, panY float64) error {
	if zoom <= 0 {
		return fmt.Errorf("zoom must be greater than 0")
	}

	m.Viewport.Zoom = zoom
	m.Viewport.PanX = panX
	m.Viewport.PanY = panY

	return nil
}

// SetNodePosition sets the position for a specific node
func (m *UIMetadata) SetNodePosition(nodeID imagegraph.NodeID, x, y float64) error {
	if nodeID.IsNil() {
		return fmt.Errorf("cannot set position for nil NodeID")
	}

	// Find existing position and update it
	for i := range m.NodePositions {
		if m.NodePositions[i].NodeID == nodeID {
			m.NodePositions[i].X = x
			m.NodePositions[i].Y = y
			return nil
		}
	}

	// Not found, add new position
	m.NodePositions = append(m.NodePositions, NodePosition{
		NodeID: nodeID,
		X:      x,
		Y:      y,
	})

	return nil
}

// GetNodePosition retrieves the position for a specific node
func (m *UIMetadata) GetNodePosition(nodeID imagegraph.NodeID) (NodePosition, bool) {
	for _, pos := range m.NodePositions {
		if pos.NodeID == nodeID {
			return pos, true
		}
	}
	return NodePosition{}, false
}

// RemoveNodePosition removes the position for a specific node
func (m *UIMetadata) RemoveNodePosition(nodeID imagegraph.NodeID) {
	for i, pos := range m.NodePositions {
		if pos.NodeID == nodeID {
			// Remove by replacing with last element and truncating
			m.NodePositions[i] = m.NodePositions[len(m.NodePositions)-1]
			m.NodePositions = m.NodePositions[:len(m.NodePositions)-1]
			return
		}
	}
}

// Clone creates a deep copy of the UIMetadata
func (m *UIMetadata) Clone() *UIMetadata {
	clone := &UIMetadata{
		GraphID:       m.GraphID,
		Viewport:      m.Viewport,
		NodePositions: make([]NodePosition, len(m.NodePositions)),
	}

	copy(clone.NodePositions, m.NodePositions)

	return clone
}
