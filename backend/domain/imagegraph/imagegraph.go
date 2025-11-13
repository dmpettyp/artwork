package imagegraph

import (
	"fmt"

	"github.com/dmpettyp/dorky"
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
	if id.IsNil() {
		return nil, fmt.Errorf("cannot create ImageGraph with nil ID")
	}

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

func (ig *ImageGraph) Clone() *ImageGraph {
	clone := *ig

	for nodeID, n := range ig.Nodes {
		c := &(*n)
		c.SetEventAdder(clone.addEvent)
		clone.Nodes[nodeID] = c
	}

	return &clone
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
	config NodeConfig,
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
	if id.IsNil() {
		return fmt.Errorf("cannot remove node with nil ID from ImageGraph %q", ig.ID)
	}

	removeNodeError := fmt.Sprintf(
		"could not remove node %q from ImageGraph %q", id, ig.ID,
	)

	node, err := ig.Nodes.Remove(id)

	if err != nil {
		return fmt.Errorf("%s: %w", removeNodeError, err)
	}

	ig.addEvent(NewNodeRemovedEvent(ig, node))

	//
	// Disconnect each upstream node's output that connects to this node
	//
	for _, input := range node.Inputs {
		if !input.Connected {
			continue
		}

		err := ig.Nodes.WithNode(input.InputConnection.NodeID, func(n *Node) error {
			return n.DisconnectOutput(
				input.InputConnection.OutputName, node.ID, input.Name,
			)
		})

		if err != nil {
			return fmt.Errorf("%s: %w", removeNodeError, err)
		}
	}

	//
	// Disconnect each downstream node's input that is connected to this node's
	// output
	//
	for _, output := range node.Outputs {
		for outputConnection := range output.Connections {
			err := ig.Nodes.WithNode(outputConnection.NodeID, func(n *Node) error {
				_, err := n.DisconnectInput(
					outputConnection.InputName,
				)
				return err
			})

			if err != nil {
				return fmt.Errorf("%s: %w", removeNodeError, err)
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
	if fromNodeID.IsNil() {
		return fmt.Errorf("cannot connect from node with nil ID in ImageGraph %q", ig.ID)
	}

	if toNodeID.IsNil() {
		return fmt.Errorf("cannot connect to node with nil ID in ImageGraph %q", ig.ID)
	}

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
		err = ig.Nodes.WithNode(inputConnection.NodeID, func(n *Node) error {
			return n.DisconnectOutput(
				inputConnection.OutputName,
				toNodeID,
				inputName,
			)
		})

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

	imageID, err := fromNode.GetOutputImage(outputName)

	if err != nil {
		return fmt.Errorf(
			"%s: couldn't get output image: %w", baseError, err,
		)
	}

	if imageID.IsNil() {
		return nil
	}

	err = toNode.SetInputImage(inputName, imageID)

	if err != nil {
		return fmt.Errorf(
			"%s: couldn't set input image: %w", baseError, err,
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
	if fromNodeID.IsNil() {
		return fmt.Errorf("cannot disconnect from node with nil ID in ImageGraph %q", ig.ID)
	}

	if toNodeID.IsNil() {
		return fmt.Errorf("cannot disconnect to node with nil ID in ImageGraph %q", ig.ID)
	}

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

// SetNodeOutputImage sets the image for a specific node's output.
// Downstream propagation is handled by event handlers.
func (ig *ImageGraph) SetNodeOutputImage(
	nodeID NodeID,
	outputName OutputName,
	imageID ImageID,
) error {
	if nodeID.IsNil() {
		return fmt.Errorf("cannot set output image for node with nil ID in ImageGraph %q", ig.ID)
	}

	node, exists := ig.Nodes.Get(nodeID)

	if !exists {
		return fmt.Errorf(
			"couldn't set node %q output image: node doesn't exist",
			nodeID,
		)
	}

	err := node.SetOutputImage(outputName, imageID)

	if err != nil {
		return fmt.Errorf("couldn't set node %q output image: %w", nodeID, err)
	}

	return nil
}

// PropagateOutputImageToConnections propagates an output image to all
// downstream nodes connected to this output
func (ig *ImageGraph) PropagateOutputImageToConnections(
	nodeID NodeID,
	outputName OutputName,
	imageID ImageID,
) error {
	if nodeID.IsNil() {
		return fmt.Errorf("cannot propagate output for node with nil ID in ImageGraph %q", ig.ID)
	}

	node, exists := ig.Nodes.Get(nodeID)

	if !exists {
		return fmt.Errorf(
			"couldn't propagate node %q output image: node doesn't exist",
			nodeID,
		)
	}

	connections, err := node.OutputConnections(outputName)

	if err != nil {
		return fmt.Errorf("couldn't propagate node %q output image: %w", nodeID, err)
	}

	//
	// Set each downstream node's input to the provided ImageID
	//
	for _, connection := range connections {
		err := ig.Nodes.WithNode(connection.NodeID, func(n *Node) error {
			return n.SetInputImage(connection.InputName, imageID)
		})

		if err != nil {
			return fmt.Errorf(
				"could not propagate node %q output image to %q: %w",
				nodeID, connection.NodeID, err,
			)
		}
	}

	return nil
}

// UnsetNodeOutputImage unsets the image for a specific node's output.
// Downstream propagation is handled by event handlers.
func (ig *ImageGraph) UnsetNodeOutputImage(
	nodeID NodeID,
	outputName OutputName,
) error {
	if nodeID.IsNil() {
		return fmt.Errorf("cannot unset output image for node with nil ID in ImageGraph %q", ig.ID)
	}

	node, exists := ig.Nodes.Get(nodeID)

	if !exists {
		return fmt.Errorf(
			"couldn't unset node %q output image: node doesn't exist",
			nodeID,
		)
	}

	err := node.UnsetOutputImage(outputName)

	if err != nil {
		return fmt.Errorf("couldn't unset node %q output image: %w", nodeID, err)
	}

	return nil
}

func (ig *ImageGraph) UnsetNodeOutputConnections(
	nodeID NodeID,
	outputName OutputName,
) error {
	if nodeID.IsNil() {
		return fmt.Errorf("cannot unset output image for node with nil ID in ImageGraph %q", ig.ID)
	}

	node, exists := ig.Nodes.Get(nodeID)

	if !exists {
		return fmt.Errorf(
			"couldn't unset node %q output image: node doesn't exist",
			nodeID,
		)
	}

	connections, err := node.OutputConnections(outputName)

	if err != nil {
		return fmt.Errorf("couldn't set node %q output image: %w", nodeID, err)
	}

	//
	// Unset each downstream node's input
	//
	for _, connection := range connections {
		err := ig.Nodes.WithNode(connection.NodeID, func(n *Node) error {
			return n.UnsetInputImage(connection.InputName)
		})

		if err != nil {
			return fmt.Errorf(
				"could not unset node %q output image: %w", nodeID, err,
			)
		}
	}

	return nil
}

// SetNodePreview sets the preview image for a specific node
func (ig *ImageGraph) SetNodePreview(
	nodeID NodeID,
	imageID ImageID,
) error {
	if nodeID.IsNil() {
		return fmt.Errorf("cannot set preview for node with nil ID in ImageGraph %q", ig.ID)
	}

	err := ig.Nodes.WithNode(nodeID, func(n *Node) error {
		return n.SetPreview(imageID)
	})

	if err != nil {
		return fmt.Errorf(
			"couldn't set node %q preview image to %q: %w",
			nodeID, imageID, err,
		)
	}

	return nil
}

// UnsetNodePreview unsets the preview image for a specific node
func (ig *ImageGraph) UnsetNodePreview(
	nodeID NodeID,
) error {
	if nodeID.IsNil() {
		return fmt.Errorf("cannot unset preview for node with nil ID in ImageGraph %q", ig.ID)
	}

	return ig.SetNodePreview(nodeID, ImageID{})
}

// SetNodeConfig sets the configuration for a specific node
func (ig *ImageGraph) SetNodeConfig(
	nodeID NodeID,
	config NodeConfig,
) error {
	if nodeID.IsNil() {
		return fmt.Errorf("cannot set config for node with nil ID in ImageGraph %q", ig.ID)
	}

	err := ig.Nodes.WithNode(nodeID, func(n *Node) error {
		return n.SetConfig(config)
	})

	if err != nil {
		return fmt.Errorf(
			"couldn't set node %q config: %w",
			nodeID, err,
		)
	}

	return nil
}

// SetNodeName sets the name for a specific node
func (ig *ImageGraph) SetNodeName(
	nodeID NodeID,
	name string,
) error {
	if nodeID.IsNil() {
		return fmt.Errorf("cannot set name for node with nil ID in ImageGraph %q", ig.ID)
	}

	err := ig.Nodes.WithNode(nodeID, func(n *Node) error {
		return n.SetName(name)
	})

	if err != nil {
		return fmt.Errorf(
			"couldn't set node %q name: %w",
			nodeID, err,
		)
	}

	return nil
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
