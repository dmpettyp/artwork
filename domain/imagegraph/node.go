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
	Inputs  []Input
	Outputs []Output
}

func NewNode(
	id NodeID,
	nodeType NodeType,
	name string,
	// inputNames []string,
	// outputNames []string,
) (*Node, error) {
	conf, ok := nodeConfigs[nodeType]

	if !ok {
		return nil, fmt.Errorf("node type %q does not have config", nodeType)
	}

	n := &Node{
		ID:      id,
		Version: 0,
		Type:    nodeType,
		Name:    name,
	}

	for _, inputName := range conf.inputNames {
		n.Inputs = append(n.Inputs, Input{Name: inputName})
	}

	for _, outputName := range conf.outputNames {
		n.Outputs = append(n.Outputs, Output{Name: outputName})
	}

	return n, nil
}

// func (n *Node) addEvent(event NodeEvent) {
// 	event.SetNodeID(n.ID)
// 	event.SetNodeVersion(n.Version.Next())
// 	n.Events = append(n.Events, event)
// }
//
// func (n *Node) GetEvents() []NodeEvent {
// 	events := n.Events
// 	n.Events = nil
// 	return events
// }

// func (n *Node) SetInputSource(name string, source OutputPort) error {
// 	for i := range n.Inputs {
// 		if n.Inputs[i].Name == name {
// 			n.Inputs[i].Source = &source
// 			return nil
// 		}
// 	}
//
// 	return fmt.Errorf(
// 		"can't set input for node %s: no such input port %s", n.ID, name,
// 	)
// }
//
// func (n *Node) UnsetInputSource(name string) error {
// 	for i := range n.Inputs {
// 		if n.Inputs[i].Name == name {
// 			n.Inputs[i].Source = nil
// 			return nil
// 		}
// 	}
//
// 	return fmt.Errorf(
// 		"can't unset input for node %s: no such input port %s", n.ID, name,
// 	)
// }
//
// func (n *Node) GetOutputPort(name string) (OutputPort, error) {
// 	for i := range n.Outputs {
// 		if n.Outputs[i].Name == name {
// 			return n.Outputs[i], nil
// 		}
// 	}
//
// 	return OutputPort{}, fmt.Errorf(
// 		"node %q doesn't have output port %q", n.ID, name,
// 	)
// }
