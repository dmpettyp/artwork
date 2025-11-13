package imagegraph

import (
	"fmt"

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
	Inputs Inputs

	// The outputs of the node that are passed to downstream nodes as inputs to
	// be processed.
	Outputs Outputs

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

	if nodeType == NodeTypeNone {
		return nil, fmt.Errorf("cannot create Node of type none")
	}

	if nodeType.NameRequired() && len(name) == 0 {
		return nil, fmt.Errorf("Node requires a name")
	}

	initState, err := state.NewState(Waiting)

	if err != nil {
		return nil, fmt.Errorf("could not create node: %w", err)
	}

	inputs, err := NewInputs(nodeType.InputNames())

	if err != nil {
		return nil, fmt.Errorf("could not create node: %w", err)
	}

	outputs, err := NewOutputs(nodeType.OutputNames())

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
		Inputs:   inputs,
		Outputs:  outputs,
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
	if n.Type.NameRequired() && len(name) == 0 {
		return fmt.Errorf("cannot set node name to empty string")
	}

	n.Name = name

	n.addEvent(NewNodeNameSetEvent(n))

	return nil
}

func (n *Node) SetPreview(imageID ImageID) error {
	if imageID.IsNil() {
		return fmt.Errorf("cannot set preview to nil image, use UnsetPreview instead")
	}

	n.Preview = imageID

	n.addEvent(NewNodePreviewSetEvent(n))

	return nil
}

func (n *Node) UnsetPreview() error {
	n.Preview = ImageID{}

	n.addEvent(NewNodePreviewUnsetEvent(n))

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
	return n.Outputs.IsOutputConnectedTo(outputName, toNodeID, inputName)
}

// SetOutputImage updates a node's output to the provided ImageID.
func (n *Node) SetOutputImage(
	outputName OutputName,
	imageID ImageID,
) error {
	if err := n.Outputs.SetImage(outputName, imageID); err != nil {
		return fmt.Errorf(
			"could not set output %q for node %q: %w", outputName, n.ID, err,
		)
	}

	n.addEvent(NewOutputImageSetEvent(n, outputName, imageID))

	if n.Outputs.AllSet() {
		err := n.State.Transition(Generated)

		if err != nil {
			return fmt.Errorf(
				"could not set output %q for node %q: %w", outputName, n.ID, err,
			)
		}
	}

	return nil
}

func (n *Node) UnsetOutputImage(outputName OutputName) error {
	oldImageID, err := n.Outputs.UnsetImage(outputName)

	if err != nil {
		return fmt.Errorf("could not unset node %q output image: %w", n.ID, err)
	}

	if !oldImageID.IsNil() {
		n.addEvent(NewOutputImageUnsetEvent(n, outputName, oldImageID))
	}

	return nil
}

func (n *Node) OutputConnections(outputName OutputName) (
	[]OutputConnection,
	error,
) {
	return n.Outputs.Connections(outputName)
}

func (n *Node) ConnectOutputTo(
	outputName OutputName,
	toNodeID NodeID,
	inputName InputName,
) error {
	if err := n.Outputs.ConnectTo(outputName, toNodeID, inputName); err != nil {
		return fmt.Errorf("could not connect output for node %q: %w", n.ID, err)
	}

	n.addEvent(NewOutputConnectedEvent(n, outputName, toNodeID, inputName))

	return nil
}

func (n *Node) DisconnectOutput(
	outputName OutputName,
	toNodeID NodeID,
	inputName InputName,
) error {
	if err := n.Outputs.DisconnectFrom(outputName, toNodeID, inputName); err != nil {
		return fmt.Errorf("could not disconnect output for node %q: %w", n.ID, err)
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
	return n.Inputs.Exists(inputName)
}

func (n *Node) ConnectInputFrom(
	inputName InputName,
	fromNodeID NodeID,
	outputName OutputName,
) error {
	if err := n.Inputs.ConnectFrom(inputName, fromNodeID, outputName); err != nil {
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
	return n.Inputs.IsConnected(inputName)
}

func (n *Node) DisconnectInput(inputName InputName) (
	InputConnection,
	error,
) {
	wasAllSet := n.Inputs.AllSet()

	inputConnection, hadImage, err := n.Inputs.Disconnect(inputName)

	if err != nil {
		return InputConnection{}, fmt.Errorf(
			"could not disconnect input for node %q: %w", n.ID, err,
		)
	}

	n.addEvent(
		NewInputDisconnectedEvent(
			n,
			inputName,
			inputConnection.NodeID,
			inputConnection.OutputName,
		),
	)

	// If input didn't have an image, we're done
	if !hadImage {
		return inputConnection, nil
	}

	n.addEvent(NewInputImageUnsetEvent(n, inputName))

	if wasAllSet {
		n.Preview = ImageID{}

		err := n.State.Transition(Waiting)

		if err != nil {
			return inputConnection, fmt.Errorf(
				"could not disconnect input %q from node %q: %w", inputName, n.ID, err,
			)
		}
	}

	n.resetOutputImages()

	return inputConnection, nil
}

// SetInputImage updates an node's input to the provided ImageID. If the
// ImageID is nil, the image is considered to be unset.
func (n *Node) SetInputImage(
	inputName InputName,
	imageID ImageID,
) error {
	err := n.Inputs.SetImage(inputName, imageID)

	if err != nil {
		return fmt.Errorf("could not set input image for node %q: %w", n.ID, err)
	}

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
	wasAllSet := n.Inputs.AllSet()

	err := n.Inputs.UnsetImage(inputName)

	if err != nil {
		return fmt.Errorf("could not unset input image: %w", err)
	}

	n.addEvent(NewInputImageUnsetEvent(n, inputName))

	if wasAllSet {
		n.Preview = ImageID{}

		err := n.State.Transition(Waiting)

		if err != nil {
			return fmt.Errorf(
				"could not unset input %q for node %q: %w", inputName, n.ID, err,
			)
		}
	}

	n.resetOutputImages()

	return nil
}

func (n *Node) triggerOutputsIfReady() error {
	if !n.Inputs.AllSet() {
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
	_ = n.Outputs.Each(func(output *Output) error {
		if output.ImageID.IsNil() {
			return nil
		}

		n.addEvent(NewOutputImageUnsetEvent(n, output.Name, output.ImageID))

		output.ResetImage()

		return nil
	})
}

func (n *Node) GetOutputImage(
	outputName OutputName,
) (
	ImageID,
	error,
) {
	return n.Outputs.GetImage(outputName)
}
