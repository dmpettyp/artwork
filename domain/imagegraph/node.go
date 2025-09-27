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
	ID      NodeID
	Version NodeVersion
	Type    NodeType
	Name    string
	Inputs  map[InputName]Input
	Outputs map[OutputName]Output
}

func NewNode(
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
		ID:      id,
		Version: 0,
		Type:    nodeType,
		Name:    name,
		Inputs:  make(map[InputName]Input),
		Outputs: make(map[OutputName]Output),
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

	return output.Connect(toNodeID, inputName)
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

	return output.Disconnect(toNodeID, inputName)
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

	return input.Connect(fromNodeID, outputName)
}

func (n *Node) IsInputConnected(inputName InputName) (
	bool,
	InputConnection,
	error,
) {
	input, ok := n.Inputs[inputName]

	if !ok {
		return false, InputConnection{}, fmt.Errorf("no input named %q exists", inputName)
	}

	connected, ic := input.IsConnected()

	return connected, ic, nil
}

func (n *Node) DisconnectInput(inputName InputName) error {
	input, ok := n.Inputs[inputName]

	if !ok {
		return fmt.Errorf("no input named %q exists", inputName)
	}

	return input.Disconnect()
}
