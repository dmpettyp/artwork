package imagegraph

import (
	"fmt"
	"slices"
)

type Nodes []*Node

func NewNodes() Nodes {
	return nil
}

func (nodes *Nodes) Add(node *Node) error {
	if _, ok := nodes.Get(node.ID); ok {
		return fmt.Errorf(
			"cannot add node: node with ID %q already exists", node.ID,
		)
	}

	*nodes = append(*nodes, node)

	return nil
}

func (nodes *Nodes) Remove(id NodeID) (*Node, error) {
	n, ok := nodes.Get(id)

	if !ok {
		return nil, fmt.Errorf(
			"cannot remove node: node with ID %q does not exist", id,
		)
	}

	*nodes = slices.DeleteFunc(
		*nodes,
		func(n *Node) bool { return n.ID == id },
	)

	return n, nil
}

func (ns *Nodes) Get(id NodeID) (*Node, bool) {
	i := slices.IndexFunc(
		*ns,
		func(n *Node) bool { return n.ID == id },
	)

	if i == -1 {
		return nil, false
	}

	return (*ns)[i], true
}
