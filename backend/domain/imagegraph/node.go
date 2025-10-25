package imagegraph

import (
	"fmt"
	"maps"
	"slices"

	"github.com/dmpettyp/state"
)

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

	State state.State[NodeState]

	// The configuration for the node. The configuration is a map containing
	// the node's settings that are provided to the image processor.
	Config NodeConfig

	// The preview image for the node
	Preview ImageID

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
	config NodeConfig,
) (
	*Node,
	error,
) {
	if id.IsNil() {
		return nil, fmt.Errorf("cannot create Node with nil ID")
	}

	if len(name) == 0 {
		return nil, fmt.Errorf("cannot create Node with empty name")
	}

	if nodeType == NodeTypeNone {
		return nil, fmt.Errorf("cannot create Node of type none")
	}

	initState, err := state.NewState(Waiting)

	if err != nil {
		return nil, fmt.Errorf("could not create node: %w", err)
	}

	n := &Node{
		ID:       id,
		State:    initState,
		addEvent: eventAdder,
		Version:  0,
		Type:     nodeType,
		Name:     name,
		Inputs:   make(map[InputName]*Input),
		Outputs:  make(map[OutputName]*Output),
	}

	for _, inputName := range nodeType.InputNames() {
		if _, ok := n.Inputs[inputName]; ok {
			return nil, fmt.Errorf("node already has an input named %q", inputName)
		}
		input := MakeInput(inputName)
		n.Inputs[inputName] = &input
	}

	for _, outputName := range nodeType.OutputNames() {
		if _, ok := n.Outputs[outputName]; ok {
			return nil, fmt.Errorf("node already has an output named %q", outputName)
		}
		output := MakeOutput(outputName)
		n.Outputs[outputName] = &output
	}

	n.addEvent(NewNodeCreatedEvent(n))

	err = n.SetConfig(config)

	if err != nil {
		return nil, fmt.Errorf("could not create node: %w", err)
	}

	return n, nil
}

func (n *Node) SetEventAdder(eventAdder func(Event)) {
	n.addEvent = eventAdder
}

func (n *Node) SetConfig(config NodeConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if err := n.Type.ValidateConfig(config); err != nil {
		return fmt.Errorf(
			"could not set config for node %q: %w", n.ID, err,
		)
	}

	n.Config = config

	n.addEvent(NewNodeConfigSetEvent(n))

	if err := n.triggerOutputsIfReady(); err != nil {
		return fmt.Errorf(
			"could not set config for node %q: %w", n.ID, err,
		)
	}

	return nil
}

func (n *Node) SetName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("cannot set node name to empty string")
	}

	n.Name = name

	n.addEvent(NewNodeNameSetEvent(n))

	return nil
}

func (n *Node) SetPreview(imageID ImageID) error {
	n.Preview = imageID

	if !imageID.IsNil() {
		n.addEvent(NewNodePreviewSetEvent(n))
	} else {
		n.addEvent(NewNodePreviewUnsetEvent(n))
	}

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
	if imageID.IsNil() {
		return nil, fmt.Errorf(
			"cannot set output %q for node %q to nil", outputName, n.ID,
		)
	}

	output, ok := n.Outputs[outputName]

	if !ok {
		return nil, fmt.Errorf("no output named %q exists", outputName)
	}

	output.SetImage(imageID)

	if n.allOutputsSet() {
		err := n.State.Transition(Generated)

		if err != nil {
			return nil, fmt.Errorf(
				"could not set output %q for node %q: %w", outputName, n.ID, err,
			)
		}
	}

	n.addEvent(NewOutputImageSetEvent(n, outputName, imageID))

	return slices.Collect(maps.Keys(output.Connections)), nil
}

// UnsetOutputImage updates a node's output ImageID to nil
func (n *Node) UnsetOutputImage(outputName OutputName) error {
	output, ok := n.Outputs[outputName]

	if !ok {
		return fmt.Errorf("no output named %q exists", outputName)
	}

	output.ResetImage()

	n.addEvent(NewOutputImageUnsetEvent(n, outputName))

	return nil
}

// UnsetOutputImage updates a node's output ImageID to nil
func (n *Node) OutputConnections(outputName OutputName) (
	[]OutputConnection,
	error,
) {
	output, ok := n.Outputs[outputName]

	if !ok {
		return nil, fmt.Errorf("no output named %q exists", outputName)
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

	if !input.HasImage() {
		return inputConnection, nil
	}

	input.ResetImage()

	// If the node previously had all inputs set, revert the state to
	// WaitingForInputs
	if !n.allInputsSet() {
		err := n.State.Transition(Waiting)

		if err != nil {
			return inputConnection, fmt.Errorf(
				"could not disconnect input %q from node %q: %w", inputName, n.ID, err,
			)
		}
	}

	n.addEvent(NewInputImageUnsetEvent(n, inputName))

	n.resetOutputImages()

	return inputConnection, nil
}

// SetInputImage updates an node's input to the provided ImageID. If the
// ImageID is nil, the image is considered to be unset.
func (n *Node) SetInputImage(
	inputName InputName,
	imageID ImageID,
) error {
	if imageID.IsNil() {
		return fmt.Errorf("cannot set input %q for node %q to nil", inputName, n.ID)
	}

	input, ok := n.Inputs[inputName]

	if !ok {
		return fmt.Errorf("no input named %q exists", inputName)
	}

	input.SetImage(imageID)

	n.addEvent(NewInputImageSetEvent(n, inputName, imageID))

	if err := n.triggerOutputsIfReady(); err != nil {
		return fmt.Errorf(
			"could not set input %q for node %q: %w", inputName, n.ID, err,
		)
	}

	return nil
}

// UnsetInputImage updates an node's input to be a nil ImageID.
func (n *Node) UnsetInputImage(
	inputName InputName,
) error {
	input, ok := n.Inputs[inputName]

	if !ok {
		return fmt.Errorf("no input named %q exists", inputName)
	}

	input.ResetImage()

	if !n.allInputsSet() {
		err := n.State.Transition(Waiting)

		if err != nil {
			return fmt.Errorf(
				"could not unset input %q for node %q: %w", inputName, n.ID, err,
			)
		}
	}

	n.addEvent(NewInputImageUnsetEvent(n, inputName))

	n.resetOutputImages()

	return nil
}

func (n *Node) triggerOutputsIfReady() error {
	if !n.allInputsSet() {
		return nil
	}

	err := n.State.Transition(Generating)

	if err != nil {
		return err
	}

	n.addEvent(NewNodeNeedsOutputsEvent(n))

	return nil
}

func (n *Node) resetOutputImages() {
	for outputName, output := range n.Outputs {
		if output.ImageID.IsNil() {
			continue
		}

		output.ResetImage()

		n.addEvent(NewOutputImageUnsetEvent(n, outputName))
	}
}

func (n *Node) GetOutputImage(
	outputName OutputName,
) (
	ImageID,
	error,
) {
	output, ok := n.Outputs[outputName]

	if !ok {
		return ImageID{}, fmt.Errorf("no output named %q exists", outputName)
	}

	return output.ImageID, nil
}

// Test to see that all inputs are connected and have an image set
func (n *Node) allInputsSet() bool {
	for _, input := range n.Inputs {
		if !input.Connected {
			return false
		}

		if input.ImageID.IsNil() {
			return false
		}
	}

	return true
}

func (n *Node) allOutputsSet() bool {
	for _, input := range n.Outputs {
		if input.ImageID.IsNil() {
			return false
		}
	}

	return true
}
