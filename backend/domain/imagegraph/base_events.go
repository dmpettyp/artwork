package imagegraph

import (
	"github.com/dmpettyp/dorky/messages"
)

// Base event type that all ImageGraph domain events extend
type ImageGraphEvent struct {
	messages.BaseEvent
	ImageGraphID      ImageGraphID      `json:"image_graph_id"`
	ImageGraphVersion ImageGraphVersion `json:"image_graph_version"`
}

func (e *ImageGraphEvent) applyImageGraph(ig *ImageGraph) {
	e.ImageGraphID = ig.ID
	e.ImageGraphVersion = ig.Version.Next()
}

func (e *ImageGraphEvent) GetAggregateVersion() int64 {
	return int64(e.ImageGraphVersion)
}

type Event interface {
	messages.Event
	applyImageGraph(ig *ImageGraph)
}

// Base event type that all Node-specific ImageGraph domain events extend
type NodeEvent struct {
	ImageGraphEvent
	NodeID      NodeID      `json:"node_id"`
	NodeState   NodeState   `json:"node_state"`
	NodeVersion NodeVersion `json:"node_version"`
	NodeType    NodeType    `json:"node_type"`
}

func (e *NodeEvent) applyNode(n *Node) {
	e.NodeID = n.ID
	e.NodeType = n.Type
	e.NodeState = n.State.Get()
	e.NodeVersion = n.Version.Next()
}
