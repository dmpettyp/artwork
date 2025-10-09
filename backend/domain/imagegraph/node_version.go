package imagegraph

// NodeVersion models the version of a node. Node versions are incremented
// every time an event is emitted by the node.
type NodeVersion int

func (v *NodeVersion) Next() NodeVersion {
	*v = *v + 1
	return *v
}
