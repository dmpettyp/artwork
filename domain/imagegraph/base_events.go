package imagegraph

import (
	"github.com/dmpettyp/dorky"
)

// Base event type that all ImageGraph domain events extend
type ImageGraphEvent struct {
	dorky.BaseEvent
	ImageGraphID      ImageGraphID
	ImageGraphVersion ImageGraphVersion
}

func (e *ImageGraphEvent) applyImageGraph(ig *ImageGraph) {
	e.ImageGraphID = ig.ID
	e.ImageGraphVersion = ig.Version.Next()
}

// Base event type that all Node-specific ImageGraph domain events extend
type NodeEvent struct {
	ImageGraphEvent
	NodeID      NodeID
	NodeVersion NodeVersion
}

func (e *NodeEvent) applyNode(n *Node) {
	e.NodeID = n.ID
	e.NodeVersion = n.Version.Next()
}
