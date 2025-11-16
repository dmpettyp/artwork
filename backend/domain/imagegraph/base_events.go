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

func (e *ImageGraphEvent) GetAggregateVersion() int64 {
	return int64(e.ImageGraphVersion)
}

type Event interface {
	dorky.Event
	applyImageGraph(ig *ImageGraph)
}

// Base event type that all Node-specific ImageGraph domain events extend
type NodeEvent struct {
	ImageGraphEvent
	NodeID      NodeID
	NodeState   NodeState
	NodeVersion NodeVersion
	NodeType    NodeType
}

func (e *NodeEvent) applyNode(n *Node) {
	e.NodeID = n.ID
	e.NodeType = n.Type
	e.NodeState = n.State.Get()
	e.NodeVersion = n.Version.Next()
}
