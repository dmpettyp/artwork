package imagegraph

type CreatedEvent struct {
	ImageGraphEvent
	Name string
}

func NewCreatedEvent(ig *ImageGraph) *CreatedEvent {
	e := &CreatedEvent{
		Name: ig.Name,
	}
	e.Init("Created")
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
	e.Init("NodeAdded")
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
	e.Init("NodeRemoved")
	return e
}

type NodeCreatedEvent struct {
	NodeEvent
	NodeType NodeType
	NodeName string
}

func NewNodeCreatedEvent(n *Node) *NodeCreatedEvent {
	e := &NodeCreatedEvent{
		NodeType: n.Type,
		NodeName: n.Name,
	}
	e.Init("NodeCreated")
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
	e.Init("NodeInputConnected")
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
	e.Init("NodeInputDisconnected")
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
	e.Init("NodeOutputConnected")
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
	e.Init("NodeOutputDisconnected")
	e.applyNode(n)
	return e
}

type NodeOutputImageSetEvent struct {
	NodeEvent
	OutputName OutputName
	ImageID    ImageID
}

func NewOutputImageSetEvent(
	n *Node,
	outputName OutputName,
	imageID ImageID,
) *NodeOutputImageSetEvent {
	e := &NodeOutputImageSetEvent{
		OutputName: outputName,
		ImageID:    imageID,
	}
	e.Init("NodeOutputImageSet")
	e.applyNode(n)
	return e
}

type NodeInputImageSetEvent struct {
	NodeEvent
	InputName InputName
	ImageID   ImageID
}

func NewInputImageSetEvent(
	n *Node,
	inputName InputName,
	imageID ImageID,
) *NodeInputImageSetEvent {
	e := &NodeInputImageSetEvent{
		InputName: inputName,
		ImageID:   imageID,
	}
	e.Init("NodeInputImageSet")
	e.applyNode(n)
	return e
}
