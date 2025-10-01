package imagegraph

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/dmpettyp/id"
)

// NodeID is the type that represents node IDs
type NodeID struct{ id.ID }

var NewNodeID, MustNewNodeID, ParseNodeID = id.Intitalizers(
	func(id id.ID) NodeID { return NodeID{ID: id} },
)

// NodeVersion models the version of a node. Node versions are incremented
// every time an event is emitted by the node.
type NodeVersion int

func (v *NodeVersion) Next() NodeVersion {
	*v = *v + 1
	return *v
}

// Node represents a node in the ImageGraph that define the image pipeline.
// Node are connected to upstream nodes through thier inputs, and to their
// downstream nodes through their outputs.
type Node struct {
	// The globally unique identifier for the Node
	ID NodeID

	// The curent version of the Node. Every time the node emits a new event
	// the node's version is incremented by one
	Version NodeVersion

	// The type of the node, representing the kind of transformations it
	// performs on its inputs to generate its outputs
	Type NodeType

	// The name assigned to the node, chosen by the ImageGraph author
	Name string

	// The configuration for the node. The configuration is a string containing
	// json that is provided to the image processor.
	Config string

	// The inputs that provide images to the node that are processed and
	// then set as outputs
	Inputs map[InputName]*Input

	// The outputs of the node that are passed to downstream nodes as inputs to
	// be processed.
	Outputs map[OutputName]*Output

	// addEvent is a function that can be used by the node to add an event
	// to its ImageGraph parent
	addEvent func(Event)
}

func NewNode(
	eventAdder func(Event),
	id NodeID,
	nodeType NodeType,
	name string,
	config string,
) (
	*Node,
	error,
) {
	nodeConfig, ok := nodeConfigs[nodeType]

	if !ok {
		return nil, fmt.Errorf("node type %q does not have config", nodeType)
	}

	n := &Node{
		ID:       id,
		addEvent: eventAdder,
		Version:  0,
		Type:     nodeType,
		Name:     name,
		Inputs:   make(map[InputName]*Input),
		Outputs:  make(map[OutputName]*Output),
	}

	err := n.SetConfig(config)

	if err != nil {
		return nil, fmt.Errorf("could not create node: %w", err)
	}

	for _, inputName := range nodeConfig.inputNames {
		if _, ok := n.Inputs[inputName]; ok {
			return nil, fmt.Errorf("node already has an input named %q", inputName)
		}
		input := MakeInput(inputName)
		n.Inputs[inputName] = &input
	}

	for _, outputName := range nodeConfig.outputNames {
		if _, ok := n.Outputs[outputName]; ok {
			return nil, fmt.Errorf("node already has an output named %q", outputName)
		}
		output := MakeOutput(outputName)
		n.Outputs[outputName] = &output
	}

	n.addEvent(NewNodeCreatedEvent(n))

	return n, nil
}

func (n *Node) SetEventAdder(eventAdder func(Event)) {
	n.addEvent = eventAdder
}

func (n *Node) SetConfig(config string) error {
	// Empty config is allowed
	if config == "" {
		n.Config = ""
		return nil
	}

	// Validate that config is valid JSON
	if !json.Valid([]byte(config)) {
		return fmt.Errorf("config must be valid JSON")
	}

	n.Config = config
	n.addEvent(NewNodeConfigSetEvent(n, config))

	return nil
}

func (n *Node) HasOutput(outputName OutputName) bool {
	_, ok := n.Outputs[outputName]
	return ok
}

func (n *Node) IsOutputConnectedTo(
	outputName OutputName,
	toNodeID NodeID,
	inputName InputName,
) (
	bool,
	error,
) {
	output, ok := n.Outputs[outputName]

	if !ok {
		return false, fmt.Errorf("no output named %q exists", outputName)
	}

	return output.IsConnected(toNodeID, inputName), nil
}

// SetOutputImage updates a node's output to the provided ImageID. If the
// ImageID is nil, the image is considered to be unset.
func (n *Node) SetOutputImage(
	outputName OutputName,
	imageID ImageID,
) (
	[]OutputConnection,
	error,
) {
	output, ok := n.Outputs[outputName]

	if !ok {
		return nil, fmt.Errorf("no output named %q exists", outputName)
	}

	output.SetImage(imageID)

	if !imageID.IsNil() {
		n.addEvent(NewOutputImageSetEvent(n, outputName, imageID))
	} else {
		n.addEvent(NewOutputImageUnsetEvent(n, outputName))
	}

	return slices.Collect(maps.Keys(output.Connections)), nil
}

func (n *Node) ConnectOutputTo(
	outputName OutputName,
	toNodeID NodeID,
	inputName InputName,
) error {
	output, ok := n.Outputs[outputName]

	if !ok {
		return fmt.Errorf("no output named %q exists", outputName)
	}

	err := output.Connect(toNodeID, inputName)

	if err != nil {
		return err
	}

	n.addEvent(NewOutputConnectedEvent(n, outputName, toNodeID, inputName))

	return nil
}

func (n *Node) DisconnectOutput(
	outputName OutputName,
	toNodeID NodeID,
	inputName InputName,
) error {
	output, ok := n.Outputs[outputName]

	if !ok {
		return fmt.Errorf("no output named %q exists", outputName)
	}

	err := output.Disconnect(toNodeID, inputName)

	if err != nil {
		return err
	}

	n.addEvent(
		NewOutputDisconnectedEvent(
			n,
			outputName,
			toNodeID,
			inputName,
		),
	)

	return nil
}

func (n *Node) HasInput(inputName InputName) bool {
	_, ok := n.Inputs[inputName]
	return ok
}

func (n *Node) ConnectInputFrom(
	inputName InputName,
	fromNodeID NodeID,
	outputName OutputName,
) error {
	input, ok := n.Inputs[inputName]

	if !ok {
		return fmt.Errorf("no input named %q exists", inputName)
	}

	err := input.Connect(fromNodeID, outputName)

	if err != nil {
		return err
	}

	n.addEvent(
		NewInputConnectedEvent(n, inputName, fromNodeID, outputName),
	)

	return nil
}

func (n *Node) IsInputConnected(inputName InputName) (
	bool,
	error,
) {
	input, ok := n.Inputs[inputName]

	if !ok {
		return false, fmt.Errorf("no input named %q exists", inputName)
	}

	return input.Connected, nil
}

func (n *Node) DisconnectInput(inputName InputName) (
	InputConnection,
	error,
) {
	input, ok := n.Inputs[inputName]

	if !ok {
		return InputConnection{}, fmt.Errorf("no input named %q exists", inputName)
	}

	inputConnection := input.InputConnection

	err := input.Disconnect()

	if err != nil {
		return inputConnection, err
	}

	n.addEvent(
		NewInputDisconnectedEvent(
			n,
			inputName,
			inputConnection.NodeID,
			inputConnection.OutputName,
		),
	)

	//
	// If the input has an image set, reset it and emit an appropriate event
	//
	if input.HasImage() {
		input.ResetImage()
		n.addEvent(NewInputImageUnsetEvent(n, inputName))
	}

	return inputConnection, nil
}

// SetInputImage updates an node's input to the provided ImageID. If the
// ImageID is nil, the image is considered to be unset.
func (n *Node) SetInputImage(
	inputName InputName,
	imageID ImageID,
) error {
	input, ok := n.Inputs[inputName]

	if !ok {
		return fmt.Errorf("no input named %q exists", inputName)
	}

	input.SetImage(imageID)

	if !imageID.IsNil() {
		n.addEvent(NewInputImageSetEvent(n, inputName, imageID))
	} else {
		n.addEvent(NewInputImageUnsetEvent(n, inputName))
	}

	return nil
}
