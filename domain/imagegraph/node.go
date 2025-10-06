package imagegraph

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/dmpettyp/id"
	"github.com/dmpettyp/state"
)

// NodeID is the type that represents node IDs
type NodeID struct{ id.ID }

var NewNodeID, MustNewNodeID, ParseNodeID = id.Inititalizers(
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

	State state.State[State]

	// The configuration for the node. The configuration is a string containing
	// json that is provided to the image processor.
	Config string

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
	config string,
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

	nodeConfig, ok := nodeConfigs[nodeType]

	if !ok {
		return nil, fmt.Errorf("node type %q does not have config", nodeType)
	}

	initState, err := state.NewState(WaitingForInputs)

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

	for _, inputName := range nodeConfig.inputs {
		if _, ok := n.Inputs[inputName]; ok {
			return nil, fmt.Errorf("node already has an input named %q", inputName)
		}
		input := MakeInput(inputName)
		n.Inputs[inputName] = &input
	}

	for _, outputName := range nodeConfig.outputs {
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

func (n *Node) SetConfig(config string) error {
	if config == "" {
		return fmt.Errorf("config cannot be empty")
	}

	if config == "null" {
		return fmt.Errorf("config cannot be null")
	}

	// Validate that config is valid JSON
	if !json.Valid([]byte(config)) {
		return fmt.Errorf("config must be valid JSON")
	}

	// Validate that config is a JSON object
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(config), &obj); err != nil {
		return fmt.Errorf("config must be a JSON object")
	}

	// Get node type configuration
	nodeConfig, ok := nodeConfigs[n.Type]
	if !ok {
		return fmt.Errorf("node type %q does not have config", n.Type)
	}

	// Validate required fields are present
	for fieldName, fieldDef := range nodeConfig.fields {
		if fieldDef.required {
			if _, exists := obj[fieldName]; !exists {
				return fmt.Errorf("required field %q is missing", fieldName)
			}
		}
	}

	// Validate field types and reject unknown fields
	for key, value := range obj {
		fieldDef, exists := nodeConfig.fields[key]
		if !exists {
			return fmt.Errorf("unknown field %q", key)
		}

		// Validate field type
		switch fieldDef.fieldType {
		case NodeConfigTypeString:
			if _, ok := value.(string); !ok {
				return fmt.Errorf("field %q must be a string", key)
			}
		case NodeConfigTypeInt:
			// JSON numbers are float64, check if it's a whole number
			if num, ok := value.(float64); !ok {
				return fmt.Errorf("field %q must be an integer", key)
			} else if num != float64(int(num)) {
				return fmt.Errorf("field %q must be an integer", key)
			}
		case NodeConfigTypeFloat:
			if _, ok := value.(float64); !ok {
				return fmt.Errorf("field %q must be a number", key)
			}
		case NodeConfigTypeBool:
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("field %q must be a boolean", key)
			}
		}
	}

	n.Config = config

	if n.State.Get() == OutputsGenerated {
		err := n.State.Transition(GeneratingOutputs)
		if err != nil {
			return fmt.Errorf("could not set config for node %q: %w", n.ID, err)
		}
		n.addEvent(NewNodeNeedsOutputsEvent(n))
	}

	n.addEvent(NewNodeConfigSetEvent(n))

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

	n.addEvent(NewOutputImageSetEvent(n, outputName, imageID))

	return slices.Collect(maps.Keys(output.Connections)), nil
}

// UnsetOutputImage updates a node's output ImageID to nil
func (n *Node) UnsetOutputImage(outputName OutputName) (
	[]OutputConnection,
	error,
) {
	output, ok := n.Outputs[outputName]

	if !ok {
		return nil, fmt.Errorf("no output named %q exists", outputName)
	}

	output.SetImage(ImageID{})

	n.addEvent(NewOutputImageUnsetEvent(n, outputName))

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
	if imageID.IsNil() {
		return fmt.Errorf("cannot set input %q for node %q to nil", inputName, n.ID)
	}

	input, ok := n.Inputs[inputName]

	if !ok {
		return fmt.Errorf("no input named %q exists", inputName)
	}

	input.SetImage(imageID)

	n.addEvent(NewInputImageSetEvent(n, inputName, imageID))

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

	input.SetImage(ImageID{})

	n.addEvent(NewInputImageUnsetEvent(n, inputName))

	return nil
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

// if n.State.Get() == OutputsGenerated {
// 	err := n.State.Transition(GeneratingOutputs)
// 	if err != nil {
// 		return fmt.Errorf("could not set config for node %q: %w", n.ID, err)
// 	}
// 	n.addEvent(NewNodeNeedsOutputsEvent(n))
// }
