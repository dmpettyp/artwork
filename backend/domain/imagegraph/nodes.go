package imagegraph

import (
	"fmt"
)

type Nodes map[NodeID]*Node

func NewNodes() Nodes {
	return make(map[NodeID]*Node)
}

func (nodes Nodes) Add(node *Node) error {
	if _, ok := nodes[node.ID]; ok {
		return fmt.Errorf(
			"cannot add node: node with ID %q already exists", node.ID,
		)
	}

	nodes[node.ID] = node

	return nil
}

func (nodes Nodes) Remove(id NodeID) (*Node, error) {
	node, ok := nodes[id]

	if !ok {
		return nil, fmt.Errorf("cannot remove node: node with ID %q does not exist", id)
	}

	delete(nodes, id)

	return node, nil
}

func (nodes Nodes) Get(id NodeID) (*Node, bool) {
	node, ok := nodes[id]
	return node, ok
}

func (nodes Nodes) WithNode(id NodeID, f func(*Node) error) error {
	if f == nil {
		return fmt.Errorf(
			"could not apply function to node %q: nil function provided", id,
		)
	}

	node, ok := nodes[id]

	if !ok {
		return fmt.Errorf("could not apply function to node %q: does not exist", id)
	}

	if err := f(node); err != nil {
		return fmt.Errorf("could not apply function to node %q: %w", id, err)
	}

	return nil
}

// HasPathBetween checks if there's a directed path from one node to another
func (nodes Nodes) HasPathBetween(fromID, toID NodeID) bool {
	visited := make(map[NodeID]bool)
	return nodes.hasPathBetweenRecursive(fromID, toID, visited)
}

func (nodes Nodes) hasPathBetweenRecursive(fromID, toID NodeID, visited map[NodeID]bool) bool {
	if fromID == toID {
		return true
	}

	if visited[fromID] {
		return false
	}
	visited[fromID] = true

	fromNode, exists := nodes.Get(fromID)
	if !exists {
		return false
	}

	// Check all downstream nodes
	for _, output := range fromNode.Outputs {
		for connection := range output.Connections {
			if nodes.hasPathBetweenRecursive(connection.NodeID, toID, visited) {
				return true
			}
		}
	}

	return false
}
