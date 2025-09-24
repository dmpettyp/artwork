package imagegraph

type CreatedEvent struct {
	ImageGraphEvent
	Name string
}

func NewCreatedEvent(ig *ImageGraph) *CreatedEvent {
	e := &CreatedEvent{
		Name: ig.Name,
	}
	e.Init("Created", "ImageGraph", ig.ID.ID)
	e.applyImageGraph(ig)
	return e
}

type NodeAddedEvent struct {
	ImageGraphEvent
	NodeID NodeID
}

func NewNodeAddedEvent(ig *ImageGraph, n *Node) *NodeAddedEvent {
	e := &NodeAddedEvent{
		NodeID: n.ID,
	}
	e.Init("NodeAdded", "ImageGraph", ig.ID.ID)
	e.applyImageGraph(ig)
	return e
}

type NodeRemovedEvent struct {
	ImageGraphEvent
	NodeID NodeID
}

func NewNodeRemovedEvent(ig *ImageGraph, n *Node) *NodeRemovedEvent {
	e := &NodeRemovedEvent{
		NodeID: n.ID,
	}
	e.Init("NodeRemoved", "ImageGraph", ig.ID.ID)
	e.applyImageGraph(ig)
	return e
}

type NodeCreatedEvent struct {
	NodeEvent
	NodeType NodeType
	NodeName string
}

func NewNodeCreatedEvent(ig *ImageGraph, n *Node) *NodeCreatedEvent {
	e := &NodeCreatedEvent{
		NodeType: n.Type,
		NodeName: n.Name,
	}
	e.Init("NodeCreated", "ImageGraph", ig.ID.ID)
	e.applyImageGraph(ig)
	e.applyNode(n)
	return e
}

type NodeConnectedEvent struct {
	NodeEvent
	// InputNodeID  NodeID
	// InputName    string
	// OutputNodeID NodeID
	// OutputName   string
}

type NodeDisconnectedEvent struct {
	NodeEvent
	// Event
	// NodeID NodeID
	// InputNodeID NodeID
	// InputName   string
}

type NodeRunnableEvent struct {
	NodeEvent
	// Event
	// NodeID NodeID
	// Inputs []Port
}
