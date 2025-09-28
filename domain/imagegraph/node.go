package imagegraph

import (
	"fmt"

	"github.com/dmpettyp/id"
)

type NodeID struct{ id.ID }

var NewNodeID, MustNewNodeID, ParseNodeID = id.Intitalizers(
	func(id id.ID) NodeID { return NodeID{ID: id} },
)

type NodeVersion int

func (v *NodeVersion) Next() NodeVersion {
	*v = *v + 1
	return *v
}

type Node struct {
	ID       NodeID
	AddEvent func(Event)
	Version  NodeVersion
	Type     NodeType
	Name     string
	Inputs   map[InputName]Input
	Outputs  map[OutputName]Output
}

func NewNode(
	addEvent func(Event),
	id NodeID,
	nodeType NodeType,
	name string,
) (
	*Node,
	error,
) {
	conf, ok := nodeConfigs[nodeType]

	if !ok {
		return nil, fmt.Errorf("node type %q does not have config", nodeType)
	}

	n := &Node{
		ID:       id,
		AddEvent: addEvent,
		Version:  0,
		Type:     nodeType,
		Name:     name,
		Inputs:   make(map[InputName]Input),
		Outputs:  make(map[OutputName]Output),
	}

	for _, inputName := range conf.inputNames {
		if _, ok := n.Inputs[inputName]; ok {
			return nil, fmt.Errorf("node already has an input named %q", inputName)
		}
		n.Inputs[inputName] = MakeInput(inputName)
	}

	for _, outputName := range conf.outputNames {
		if _, ok := n.Outputs[outputName]; ok {
			return nil, fmt.Errorf("node already has an output named %q", outputName)
		}
		n.Outputs[outputName] = MakeOutput(outputName)
	}

	n.AddEvent(NewNodeCreatedEvent(n))

	return n, nil
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

	n.AddEvent(NewOutputConnectedEvent(n, outputName, toNodeID, inputName))

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

	n.AddEvent(
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

	n.AddEvent(
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

	n.AddEvent(
		NewInputDisconnectedEvent(
			n,
			inputName,
			inputConnection.NodeID,
			inputConnection.OutputName,
		),
	)

	return inputConnection, nil
}
