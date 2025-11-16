package imagegraph

import "fmt"

type CreatedEvent struct {
	ImageGraphEvent
	Name string `json:"name"`
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
	NodeID NodeID `json:"node_id"`
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
	NodeID NodeID `json:"node_id"`
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
	NodeType NodeType `json:"node_type"`
	NodeName string   `json:"node_name"`
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
	InputName      InputName  `json:"input_name"`
	FromNodeID     NodeID     `json:"from_node_id"`
	FromOutputName OutputName `json:"from_output_name"`
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
	InputName      InputName  `json:"input_name"`
	FromNodeID     NodeID     `json:"from_node_id"`
	FromOutputName OutputName `json:"from_output_name"`
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
	OutputName  OutputName `json:"output_name"`
	ToNodeID    NodeID     `json:"to_node_id"`
	ToInputName InputName  `json:"to_input_name"`
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
	OutputName  OutputName `json:"output_name"`
	ToNodeID    NodeID     `json:"to_node_id"`
	ToInputName InputName  `json:"to_input_name"`
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
	OutputName OutputName `json:"output_name"`
	ImageID    ImageID    `json:"image_id"`
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

type NodeOutputImageUnsetEvent struct {
	NodeEvent
	OutputName OutputName `json:"output_name"`
	ImageID    ImageID    `json:"image_id"`
}

func NewOutputImageUnsetEvent(
	n *Node,
	outputName OutputName,
	imageID ImageID,
) *NodeOutputImageUnsetEvent {
	e := &NodeOutputImageUnsetEvent{
		OutputName: outputName,
		ImageID:    imageID,
	}
	e.Init("NodeOutputImageUnset")
	e.applyNode(n)
	return e
}

type NodeInputImageSetEvent struct {
	NodeEvent
	InputName InputName `json:"input_name"`
	ImageID   ImageID   `json:"image_id"`
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

type NodeInputImageUnsetEvent struct {
	NodeEvent
	InputName InputName `json:"input_name"`
}

func NewInputImageUnsetEvent(
	n *Node,
	inputName InputName,
) *NodeInputImageUnsetEvent {
	e := &NodeInputImageUnsetEvent{
		InputName: inputName,
	}
	e.Init("NodeInputImageUnset")
	e.applyNode(n)
	return e
}

type NodeConfigSetEvent struct {
	NodeEvent
	Config NodeConfig `json:"config"`
}

func NewNodeConfigSetEvent(n *Node) *NodeConfigSetEvent {
	e := &NodeConfigSetEvent{
		Config: n.Config,
	}
	e.Init("NodeConfigSet")
	e.applyNode(n)
	return e
}

type NodeNameSetEvent struct {
	NodeEvent
	Name string `json:"name"`
}

func NewNodeNameSetEvent(n *Node) *NodeNameSetEvent {
	e := &NodeNameSetEvent{
		Name: n.Name,
	}
	e.Init("NodeNameSet")
	e.applyNode(n)
	return e
}

type NodePreviewSetEvent struct {
	NodeEvent
	ImageID ImageID `json:"image_id"`
}

func NewNodePreviewSetEvent(n *Node) *NodePreviewSetEvent {
	e := &NodePreviewSetEvent{
		ImageID: n.Preview,
	}
	e.Init("NodePreviewSet")
	e.applyNode(n)
	return e
}

type NodePreviewUnsetEvent struct {
	NodeEvent
}

func NewNodePreviewUnsetEvent(n *Node) *NodePreviewUnsetEvent {
	e := &NodePreviewUnsetEvent{}
	e.Init("NodePreviewUnset")
	e.applyNode(n)
	return e
}

type nodeInput struct {
	Name    InputName `json:"name"`
	ImageID ImageID   `json:"image_id"`
}

type NodeNeedsOutputsEvent struct {
	NodeEvent
	NodeConfig NodeConfig  `json:"node_config"`
	Inputs     []nodeInput `json:"inputs"`
}

func NewNodeNeedsOutputsEvent(n *Node) *NodeNeedsOutputsEvent {
	e := &NodeNeedsOutputsEvent{
		NodeConfig: n.Config,
	}
	e.Init("NodeNeedsOutputs")
	e.applyNode(n)

	for name, input := range n.Inputs {
		e.Inputs = append(
			e.Inputs,
			nodeInput{
				Name:    name,
				ImageID: input.ImageID,
			},
		)
	}
	return e
}

// GetInput retrieves an input image by name, returning an error if not found or nil
func (e *NodeNeedsOutputsEvent) GetInput(name InputName) (ImageID, error) {
	for _, input := range e.Inputs {
		if input.Name == name {
			if input.ImageID.IsNil() {
				return ImageID{}, fmt.Errorf("input %q has nil image", name)
			}
			return input.ImageID, nil
		}
	}
	return ImageID{}, fmt.Errorf("input %q not found", name)
}
