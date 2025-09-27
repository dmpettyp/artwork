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

type NodeInputConnectedEvent struct {
	NodeEvent
	InputName      InputName
	FromNodeID     NodeID
	FromOutputName OutputName
}

func NewInputConnectedEvent(
	ig *ImageGraph,
	n *Node,
	inputName InputName,
	fromNodeID NodeID,
	fromOutputName OutputName,
) *NodeInputConnectedEvent {
	e := &NodeInputConnectedEvent{
		InputName:      inputName,
		FromNodeID:     fromNodeID,
		FromOutputName: fromOutputName,
	}
	e.Init("NodeInputConnected", "ImageGraph", ig.ID.ID)
	e.applyImageGraph(ig)
	e.applyNode(n)
	return e
}

type NodeInputDisconnectedEvent struct {
	NodeEvent
	InputName      InputName
	FromNodeID     NodeID
	FromOutputName OutputName
}

func NewInputDisconnectedEvent(
	ig *ImageGraph,
	n *Node,
	inputName InputName,
	fromNodeID NodeID,
	fromOutputName OutputName,
) *NodeInputDisconnectedEvent {
	e := &NodeInputDisconnectedEvent{
		InputName:      inputName,
		FromNodeID:     fromNodeID,
		FromOutputName: fromOutputName,
	}
	e.Init("NodeInputDisconnected", "ImageGraph", ig.ID.ID)
	e.applyImageGraph(ig)
	e.applyNode(n)
	return e
}

type NodeOutputConnectedEvent struct {
	NodeEvent
	OutputName  OutputName
	ToNodeID    NodeID
	ToInputName InputName
}

func NewOutputConnectedEvent(
	ig *ImageGraph,
	n *Node,
	outputName OutputName,
	toNodeID NodeID,
	toInputName InputName,
) *NodeOutputConnectedEvent {
	e := &NodeOutputConnectedEvent{
		OutputName:  outputName,
		ToNodeID:    toNodeID,
		ToInputName: toInputName,
	}
	e.Init("NodeOutputConnected", "ImageGraph", ig.ID.ID)
	e.applyImageGraph(ig)
	e.applyNode(n)
	return e
}

type NodeOutputDisconnectedEvent struct {
	NodeEvent
	OutputName  OutputName
	ToNodeID    NodeID
	ToInputName InputName
}

func NewOutputDisconnectedEvent(
	ig *ImageGraph,
	n *Node,
	outputName OutputName,
	toNodeID NodeID,
	toInputName InputName,
) *NodeOutputDisconnectedEvent {
	e := &NodeOutputDisconnectedEvent{
		OutputName:  outputName,
		ToNodeID:    toNodeID,
		ToInputName: toInputName,
	}
	e.Init("NodeOutputDisconnected", "ImageGraph", ig.ID.ID)
	e.applyImageGraph(ig)
	e.applyNode(n)
	return e
}
