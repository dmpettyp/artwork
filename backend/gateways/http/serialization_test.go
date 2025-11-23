package http

import (
	"testing"

	"github.com/dmpettyp/artwork/domain/imagegraph"
)

func TestNodeTypeMapperIsComplete(t *testing.T) {
	for nodeType := range imagegraph.NodeTypeDefs {
		str := imagegraph.NodeTypeMapper.FromWithDefault(nodeType, "MISSING")
		if str == "MISSING" {
			t.Fatalf("NodeType %v not in mapper", nodeType)
		}

		t.Run(str, func(t *testing.T) {
			roundtrip, err := imagegraph.NodeTypeMapper.To(str)
			if err != nil {
				t.Fatalf("Failed to round-trip %v: %v", nodeType, err)
			}
			if roundtrip != nodeType {
				t.Errorf("Round-trip failed: got %v, want %v", roundtrip, nodeType)
			}
		})
	}
}

func TestNodeStateMapperIsComplete(t *testing.T) {
	for _, nodeState := range imagegraph.AllNodeStates() {
		str := imagegraph.NodeStateMapper.FromWithDefault(nodeState, "MISSING")
		if str == "MISSING" {
			t.Fatalf("NodeState %v not in mapper", nodeState)
		}

		roundtrip, err := imagegraph.NodeStateMapper.To(str)
		if err != nil {
			t.Fatalf("Failed to round-trip %v: %v", nodeState, err)
		}
		if roundtrip != nodeState {
			t.Errorf("Round-trip failed: got %v, want %v", roundtrip, nodeState)
		}
	}
}
