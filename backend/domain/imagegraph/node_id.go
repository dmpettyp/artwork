package imagegraph

import "github.com/dmpettyp/dorky/id"

// NodeID is the type that represents node IDs
type NodeID struct{ id.ID }

var NewNodeID, MustNewNodeID, ParseNodeID = id.Create(
	func(id id.ID) NodeID { return NodeID{ID: id} },
)
