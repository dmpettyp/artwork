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

// NewImageGraph creates and initializes a new ImageGraph
func NewImageGraph(
	id ImageGraphID,
	name string,
) (
	*ImageGraph,
	error,
) {
	if len(name) == 0 {
		return nil, fmt.Errorf("cannot create ImageGraph with empty name")
	}

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
	config string,
) error {
	n, err := NewNode(ig.addEvent, id, nodeType, name, config)

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
		if !input.Connected {
			continue
		}

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
	// Ensure that we aren't connecting the node to itself
	//
	if fromNodeID == toNodeID {
		return fmt.Errorf("%s: cannot connect node to itself", baseError)
	}

	//
	// Determine if the connection would create a cycle, cycles are not allowed
	// in the imagegraph
	//
	if ig.wouldCreateCycle(fromNodeID, toNodeID) {
		return fmt.Errorf("%s: would create cycle", baseError)
	}

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
			"%s: to node %q doesn't have input %q", baseError, toNodeID, inputName,
		)
	}

	//
	// If this connection already exists, do nothing
	//
	connectionExists, err := fromNode.IsOutputConnectedTo(
		outputName,
		toNodeID,
		inputName,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", baseError, err)
	}

	if connectionExists {
		return nil
	}

	//
	// If the input is already connected to another nodes' output, disconnect it
	//
	connected, err := toNode.IsInputConnected(inputName)

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
	// Connect the target input from the sources output and emit an event
	//
	err = toNode.ConnectInputFrom(inputName, fromNodeID, outputName)

	if err != nil {
		return fmt.Errorf(
			"%s: couldn't connect input: %w", baseError, err,
		)
	}

	return nil
}

// DisconnectNodes removes a connection from one node's output to another
// node's input.
func (ig *ImageGraph) DisconnectNodes(
	fromNodeID NodeID,
	outputName OutputName,
	toNodeID NodeID,
	inputName InputName,
) error {
	baseError := fmt.Sprintf(
		"error disconnecting node %s:%s from node %s:%s in imagegraph %s",
		fromNodeID, outputName,
		toNodeID, inputName,
		ig.ID,
	)

	//
	// Ensure that the source node exists and has the output
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
	// Ensure that the target node exists and has the input
	//
	toNode, exists := ig.Nodes.Get(toNodeID)

	if !exists {
		return fmt.Errorf("%s: to node doesn't exist", baseError)
	}

	if !toNode.HasInput(inputName) {
		return fmt.Errorf(
			"%s: to node doesn't have input %q", baseError, inputName,
		)
	}

	//
	// If this connection doesn't exist, do nothing (idempotent)
	//
	connectionExists, err := fromNode.IsOutputConnectedTo(
		outputName,
		toNodeID,
		inputName,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", baseError, err)
	}

	if !connectionExists {
		return nil
	}

	//
	// Disconnect the source node's output and emit an event
	//
	err = fromNode.DisconnectOutput(outputName, toNodeID, inputName)

	if err != nil {
		return fmt.Errorf(
			"%s: couldn't disconnect output: %w", baseError, err,
		)
	}

	//
	// Disconnect the target node's input and emit an event
	//
	_, err = toNode.DisconnectInput(inputName)

	if err != nil {
		return fmt.Errorf(
			"%s: couldn't disconnect input: %w", baseError, err,
		)
	}

	return nil
}

// SetNodeOutputImage sets the image for a specific node's output and
// propagates it to all downstream nodes that have it set as an input
func (ig *ImageGraph) SetNodeOutputImage(
	nodeID NodeID,
	outputName OutputName,
	imageID ImageID,
) error {
	node, exists := ig.Nodes.Get(nodeID)

	if !exists {
		return fmt.Errorf(
			"couldn't set node %q output image: node doesn't exist",
			nodeID,
		)
	}

	connections, err := node.SetOutputImage(outputName, imageID)

	if err != nil {
		return fmt.Errorf("couldn't set node %q output image: %w", nodeID, err)
	}

	//
	// Set each downstream node's input to the provided ImageID
	//
	for _, connection := range connections {
		downstreamNode, exists := ig.Nodes.Get(connection.NodeID)

		if !exists {
			return fmt.Errorf(
				"could not set node %q output image: downstream node %q does not exist",
				nodeID, connection.NodeID,
			)
		}

		err := downstreamNode.SetInputImage(
			connection.InputName,
			imageID,
		)

		if err != nil {
			return fmt.Errorf("could not set node %q output image: %w", nodeID, err)
		}
	}

	return nil
}

// UnsetNodeOutputImage unsets the image for a specific node's output and
// propagates it to all downstream nodes that have it set as an input
func (ig *ImageGraph) UnsetNodeOutputImage(
	nodeID NodeID,
	outputName OutputName,
) error {
	return ig.SetNodeOutputImage(
		nodeID,
		outputName,
		ImageID{},
	)
}

// SetNodePreview sets the preview image for a specific node
func (ig *ImageGraph) SetNodePreview(
	nodeID NodeID,
	imageID ImageID,
) error {
	node, exists := ig.Nodes.Get(nodeID)

	if !exists {
		return fmt.Errorf(
			"couldn't set node %q preview: node doesn't exist",
			nodeID,
		)
	}

	err := node.SetPreview(imageID)

	if err != nil {
		return fmt.Errorf("couldn't set node %q preview: %w", nodeID, err)
	}

	return nil
}

// UnsetNodePreview unsets the preview image for a specific node
func (ig *ImageGraph) UnsetNodePreview(
	nodeID NodeID,
) error {
	return ig.SetNodePreview(nodeID, ImageID{})
}

// wouldCreateCycle checks if connecting fromNodeID to toNodeID would create a cycle
func (ig *ImageGraph) wouldCreateCycle(fromNodeID, toNodeID NodeID) bool {
	// If we connect fromNode -> toNode, check if there's already a path
	// from toNode back to fromNode (which would create a cycle)
	visited := make(map[NodeID]bool)
	return ig.hasPathBetween(toNodeID, fromNodeID, visited)
}

// hasPathBetween checks if there's a path from fromID to toID in the graph
func (ig *ImageGraph) hasPathBetween(fromID, toID NodeID, visited map[NodeID]bool) bool {
	if fromID == toID {
		return true
	}

	if visited[fromID] {
		return false
	}
	visited[fromID] = true

	fromNode, exists := ig.Nodes.Get(fromID)
	if !exists {
		return false
	}

	// Check all downstream nodes
	for _, output := range fromNode.Outputs {
		for connection := range output.Connections {
			if ig.hasPathBetween(connection.NodeID, toID, visited) {
				return true
			}
		}
	}

	return false
}
