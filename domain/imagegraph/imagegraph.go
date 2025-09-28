package imagegraph

import (
	"fmt"

	"github.com/dmpettyp/dorky"
	"github.com/dmpettyp/id"
)

type ImageGraphID struct{ id.ID }

var NewImageGraphID, MustNewImageGraphID, ParseImageGraphID = id.Intitalizers(
	func(id id.ID) ImageGraphID { return ImageGraphID{ID: id} },
)

// A ImageGraph models an graph that consists of Nodes connected together to
// drive image creation
type ImageGraph struct {
	dorky.Entity

	// Unique Identifier for the ImageGraph
	ID ImageGraphID

	// Author-created name for the ImageGraph
	Name string

	// The version of the ImageGraph. Every time the ImageGraph is updated its
	// version is incremented
	Version ImageGraphVersion

	// The list of transform Nodes that exist in the image graph
	Nodes Nodes
}

// NewImageGraph ✅
// AddNode ✅
// RemoveNode ✅
// ConnectNodes ✅
// DisconnectNodes
// SetNodeOutputImage
// UnsetNodeOutputImage

// NewImageGraph creates and initializes a new ImageGraph
func NewImageGraph(
	id ImageGraphID,
	name string,
) (
	*ImageGraph,
	error,
) {
	ig := &ImageGraph{
		ID:      id,
		Name:    name,
		Version: 0,
		Nodes:   NewNodes(),
	}

	ig.addEvent(NewCreatedEvent(ig))

	return ig, nil
}

func (ig *ImageGraph) addEvent(e Event) {
	e.SetEntity("ImageGraph", ig.ID.ID)
	e.applyImageGraph(ig)
	ig.AddEvent(e)
}

// AddNode adds a node to an ImageGraph
func (ig *ImageGraph) AddNode(
	id NodeID,
	nodeType NodeType,
	name string,
) error {
	n, err := NewNode(ig.addEvent, id, nodeType, name)

	if err != nil {
		return fmt.Errorf("could not create node for ImageGraph %q: %w", ig.ID, err)
	}

	err = ig.Nodes.Add(n)

	if err != nil {
		return fmt.Errorf("could not add node to ImageGraph %q: %w", ig.ID, err)
	}

	ig.addEvent(NewNodeAddedEvent(ig, n))

	return nil
}

// RemoveNode removes an existing node from the ImageGraph. All connections
// between the removed node and other nodes are removed along with the node.
func (ig *ImageGraph) RemoveNode(
	id NodeID,
) error {

	n, err := ig.Nodes.Remove(id)

	if err != nil {
		return fmt.Errorf(
			"could not remove node %q from ImageGraph %q: %w", id, ig.ID, err,
		)
	}

	ig.addEvent(NewNodeRemovedEvent(ig, n))

	//
	// Disconnect each upstream node's output that connects to this node
	//
	for _, input := range n.Inputs {
		inputConnection := input.InputConnection

		upstreamNode, exists := ig.Nodes.Get(inputConnection.NodeID)

		if !exists {
			return fmt.Errorf(
				"could not remove node %q from ImageGraph %q: input source does not exist",
				id, ig.ID,
			)
		}

		err := upstreamNode.DisconnectOutput(
			inputConnection.OutputName, n.ID, input.Name,
		)

		if err != nil {
			return fmt.Errorf(
				"could not remove node %q from ImageGraph %q: %w",
				id, ig.ID, err,
			)
		}
	}

	//
	// Disconnect each downstream node's input that is connected to this node's
	// output
	//
	for _, output := range n.Outputs {
		for outputConnection := range output.Connections {
			downstreamNode, exists := ig.Nodes.Get(outputConnection.NodeID)

			if !exists {
				return fmt.Errorf(
					"could not remove node %q from ImageGraph %q: output target does not exist",
					id, ig.ID,
				)
			}

			_, err := downstreamNode.DisconnectInput(
				outputConnection.InputName,
			)

			if err != nil {
				return fmt.Errorf(
					"could not remove node %q from ImageGraph %q: %w",
					id, ig.ID, err,
				)
			}
		}
	}

	return nil
}

// ConnectNodes creates a connection from one node's output to another node's
// input.
func (ig *ImageGraph) ConnectNodes(
	fromNodeID NodeID,
	outputName OutputName,
	toNodeID NodeID,
	inputName InputName,
) error {
	baseError := fmt.Sprintf(
		"error connecting node %s:%s to node %s:%s in imagegraph %s",
		fromNodeID, outputName,
		toNodeID, inputName,
		ig.ID,
	)

	//
	// Ensure that the source node exists and has the output to be connected from
	//
	fromNode, exists := ig.Nodes.Get(fromNodeID)

	if !exists {
		return fmt.Errorf("%s: from node doesn't exist", baseError)
	}

	if !fromNode.HasOutput(outputName) {
		return fmt.Errorf(
			"%s: from node doesn't have output %q", baseError, outputName,
		)
	}

	//
	// Ensure that the target node exists and has the input to be connected to
	//
	toNode, exists := ig.Nodes.Get(toNodeID)

	if !exists {
		return fmt.Errorf("%s: to node doesn't exist", baseError)
	}

	if !toNode.HasInput(inputName) {
		return fmt.Errorf(
			"%s: from node doesn't have output %q", baseError, outputName,
		)
	}

	//
	// If this connection already exists, do nothing
	//
	connected, err := fromNode.IsOutputConnectedTo(
		outputName,
		toNodeID,
		inputName,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", baseError, err)
	}

	if connected {
		return nil
	}

	//
	// If the input is already connected to another nodes' output, disconnect it
	//
	connected, err = toNode.IsInputConnected(inputName)

	if err != nil {
		return fmt.Errorf("%s: %w", baseError, err)
	}

	if connected {
		//
		// Disconnect the target node's input and emit an event
		//
		inputConnection, err := toNode.DisconnectInput(inputName)

		if err != nil {
			return fmt.Errorf(
				"%s: couldn't disconnect former output: %w", baseError, err,
			)
		}

		//
		// Disconnect the target node's original source output and emit an event
		//
		previousFromNode, exists := ig.Nodes.Get(inputConnection.NodeID)

		if !exists {
			return fmt.Errorf("%s: former from node doesn't exist", baseError)
		}

		err = previousFromNode.DisconnectOutput(
			inputConnection.OutputName,
			toNodeID,
			inputName,
		)

		if err != nil {
			return fmt.Errorf(
				"%s: couldn't disconnect former output: %w", baseError, err,
			)
		}
	}

	//
	// Connect the source output to the target input and emit an event
	//
	err = fromNode.ConnectOutputTo(outputName, toNodeID, inputName)

	if err != nil {
		return fmt.Errorf(
			"%s: couldn't connect output: %w", baseError, err,
		)
	}

	//
	// Connect the target input from the soruces output and emit an event
	//
	err = toNode.ConnectInputFrom(inputName, fromNodeID, outputName)

	if err != nil {
		return fmt.Errorf(
			"%s: couldn't connect input: %w", baseError, err,
		)
	}

	return nil
}

/*

// DisconnectNode removes a connection from a node's input
func (p *ImageGraph) DisconnectNode(
	inputNodeID node.ID,
	inputName string,
) error {
	inputNode, err := p.findNode(inputNodeID)

	if err != nil {
		return fmt.Errorf(
			"can't disconnect %s:%s: %w", inputNodeID, inputName, err,
		)
	}

	err = inputNode.UnsetInputSource(inputName)

	if err != nil {
		return fmt.Errorf(
			"can't disconnect %s:%s: %w", inputNodeID, inputName, err,
		)
	}

	p.addEvent(
		&NodeDisconnectedEvent{
			InputNodeID: inputNodeID,
			InputName:   inputName,
		},
	)

	return nil
}

// TriggerNode acknowledges that a node in the ImageGraph has been run and that
// its outputs should be scheduled for running
func (p *ImageGraph) TriggerNode(id node.ID) error {
	// _, err := p.findNode(id)
	//
	// if err != nil {
	// 	return fmt.Errorf("could not trigger node %s: %w", id, err)
	// }
	//
	// nodesToSchedule := mapset.NewSet[node.ID]()
	//
	// for inputPort, outputPort := range p.Connections {
	// 	if outputPort.NodeID != id {
	// 		continue
	// 	}
	//
	// 	nodesToSchedule.Add(inputPort.NodeID)
	// }
	//
	// for _, nodeID := range nodesToSchedule.ToSlice() {
	// 	err := p.scheduleNode(nodeID)
	//
	// 	if err != nil {
	// 		return fmt.Errorf("could not trigger node %s: %w", nodeID, err)
	// 	}
	// }

	return nil
}

// scheduleNode schedules a node to be run if all of its inputs are connected
// to another node's output
func (p *ImageGraph) scheduleNode(id node.ID) error {
	// n, err := p.findNode(id)
	//
	// if err != nil {
	// 	return fmt.Errorf("could not schedule node %s: %w", id, err)
	// }
	//
	// outputPorts := make([]OutputPort, 0, len(n.InputNames))
	//
	// for _, inputPort := range n.Inputs {
	// 	if inputPort.Source == nil {
	// 		return nil
	// 	}
	//
	// 	outputPorts = append(outputPorts, *inputPort.Source)
	// }
	//
	// p.addEvent(
	// 	&NodeRunnableEvent{
	// 		NodeID: id,
	// 		Inputs: outputPorts,
	// 	},
	// )

	return nil
}

// findNode retrieves a node by its ID if it has been added to the ImageGraph
func (p *ImageGraph) findNode(id node.ID) (*Node, error) {
	for i := range p.Nodes {
		if p.Nodes[i].ID == id {
			return &p.Nodes[i], nil
		}
	}

	return nil, fmt.Errorf("node %s doesn't exist", id)
}
*/
