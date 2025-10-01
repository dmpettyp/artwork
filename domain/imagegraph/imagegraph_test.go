package imagegraph_test

import (
	"testing"

	"github.com/dmpettyp/artwork/domain/imagegraph"
)

func TestNewImageGraph(t *testing.T) {
	t.Run("creates image graph with valid parameters", func(t *testing.T) {
		id := imagegraph.MustNewImageGraphID()
		name := "test-graph"

		ig, err := imagegraph.NewImageGraph(id, name)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if ig == nil {
			t.Fatal("expected image graph, got nil")
		}

		if ig.ID != id {
			t.Errorf("expected ID %v, got %v", id, ig.ID)
		}

		if ig.Name != name {
			t.Errorf("expected name %q, got %q", name, ig.Name)
		}

		if ig.Version != 1 {
			t.Errorf("expected version 1, got %v", ig.Version)
		}

		if ig.Nodes == nil {
			t.Error("expected nodes to be initialized")
		}

		if len(ig.Nodes) != 0 {
			t.Errorf("expected 0 nodes, got %d", len(ig.Nodes))
		}
	})

	t.Run("emits created event", func(t *testing.T) {
		id := imagegraph.MustNewImageGraphID()
		name := "test-graph"

		ig, err := imagegraph.NewImageGraph(id, name)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := ig.GetEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		createdEvent, ok := events[0].(*imagegraph.CreatedEvent)

		if !ok {
			t.Fatalf("expected CreatedEvent")
		}

		if createdEvent.EventType != "Created" {
			t.Fatalf("expected Created event, got %s", createdEvent.EventType)
		}
	})

	t.Run("increments version on creation", func(t *testing.T) {
		id := imagegraph.MustNewImageGraphID()

		ig, err := imagegraph.NewImageGraph(id, "test")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if ig.Version != 1 {
			t.Errorf("expected version 1 after creation event, got %v", ig.Version)
		}
	})

	t.Run("returns error for empty name", func(t *testing.T) {
		id := imagegraph.MustNewImageGraphID()

		ig, err := imagegraph.NewImageGraph(id, "")

		if err == nil {
			t.Fatal("expected error for empty name, got nil")
		}

		if ig != nil {
			t.Errorf("expected nil image graph, got %v", ig)
		}
	})
}

func TestImageGraph_AddNode(t *testing.T) {
	t.Run("adds node with valid parameters", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()

		err := ig.AddNode(nodeID, imagegraph.NodeTypeInput, "input-node", "{}")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(ig.Nodes) != 1 {
			t.Errorf("expected 1 node, got %d", len(ig.Nodes))
		}

		node, exists := ig.Nodes.Get(nodeID)
		if !exists {
			t.Fatal("expected node to exist")
		}

		if node.ID != nodeID {
			t.Errorf("expected node ID %v, got %v", nodeID, node.ID)
		}

		if node.Name != "input-node" {
			t.Errorf("expected node name %q, got %q", "input-node", node.Name)
		}

		if node.Type != imagegraph.NodeTypeInput {
			t.Errorf("expected node type %v, got %v", imagegraph.NodeTypeInput, node.Type)
		}
	})

	t.Run("increments version when node added", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		initialVersion := ig.Version

		err := ig.AddNode(imagegraph.MustNewNodeID(), imagegraph.NodeTypeInput, "node", "{}")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if ig.Version != initialVersion+3 {
			t.Errorf("expected version %v, got %v", initialVersion+3, ig.Version)
		}
	})

	t.Run("emits NodeAdded event", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		ig.ResetEvents()
		nodeID := imagegraph.MustNewNodeID()

		err := ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := ig.GetEvents()
		// Should emit NodeCreatedEvent, NodeConfigSetEvent, and NodeAddedEvent
		if len(events) != 3 {
			t.Fatalf("expected 3 events, got %d", len(events))
		}

		if _, ok := events[0].(*imagegraph.NodeCreatedEvent); !ok {
			t.Errorf("expected first event to be NodeCreatedEvent, got %T", events[0])
		}

		if _, ok := events[1].(*imagegraph.NodeConfigSetEvent); !ok {
			t.Errorf("expected second event to be NodeConfigSetEvent, got %T", events[1])
		}

		if _, ok := events[2].(*imagegraph.NodeAddedEvent); !ok {
			t.Errorf("expected third event to be NodeAddedEvent, got %T", events[2])
		}
	})

	t.Run("returns error for duplicate node ID", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()

		err := ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node1", "{}")
		if err != nil {
			t.Fatalf("expected no error on first add, got %v", err)
		}

		err = ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node2", "{}")
		if err == nil {
			t.Fatal("expected error for duplicate node ID, got nil")
		}
	})

	t.Run("returns error for invalid node type", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")

		err := ig.AddNode(imagegraph.MustNewNodeID(), imagegraph.NodeTypeNone, "node", "{}")

		if err == nil {
			t.Fatal("expected error for invalid node type, got nil")
		}
	})

	t.Run("can add multiple nodes", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")

		err := ig.AddNode(imagegraph.MustNewNodeID(), imagegraph.NodeTypeInput, "node1", "{}")
		if err != nil {
			t.Fatalf("expected no error adding node1, got %v", err)
		}

		err = ig.AddNode(imagegraph.MustNewNodeID(), imagegraph.NodeTypeInput, "node2", "{}")
		if err != nil {
			t.Fatalf("expected no error adding node2, got %v", err)
		}

		if len(ig.Nodes) != 2 {
			t.Errorf("expected 2 nodes, got %d", len(ig.Nodes))
		}
	})
}

func TestNode_SetConfig(t *testing.T) {
	t.Run("accepts valid JSON object", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")

		node, _ := ig.Nodes.Get(nodeID)
		config := `{"key": "value", "number": 42}`

		err := node.SetConfig(config)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if node.Config != config {
			t.Errorf("expected config %q, got %q", config, node.Config)
		}
	})

	t.Run("returns error for JSON array", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")

		node, _ := ig.Nodes.Get(nodeID)
		config := `[1, 2, 3]`

		err := node.SetConfig(config)

		if err == nil {
			t.Fatal("expected error for JSON array, got nil")
		}
	})

	t.Run("returns error for JSON primitives", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")

		node, _ := ig.Nodes.Get(nodeID)

		testCases := []string{`"string"`, `42`, `true`, `null`}

		for _, config := range testCases {
			err := node.SetConfig(config)
			if err == nil {
				t.Errorf("expected error for primitive %q, got nil", config)
			}
		}
	})

	t.Run("returns error for empty string", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")

		node, _ := ig.Nodes.Get(nodeID)

		err := node.SetConfig("")

		if err == nil {
			t.Fatal("expected error for empty config, got nil")
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")

		node, _ := ig.Nodes.Get(nodeID)

		invalidConfigs := []string{
			`{invalid}`,
			`{"unclosed": `,
			`not json at all`,
			`{"key": undefined}`,
		}

		for _, config := range invalidConfigs {
			err := node.SetConfig(config)
			if err == nil {
				t.Errorf("expected error for invalid config %q, got nil", config)
			}
		}
	})

	t.Run("emits NodeConfigSet event", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")
		ig.ResetEvents()

		node, _ := ig.Nodes.Get(nodeID)
		config := `{"key": "value"}`

		err := node.SetConfig(config)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := ig.GetEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		if _, ok := events[0].(*imagegraph.NodeConfigSetEvent); !ok {
			t.Errorf("expected NodeConfigSetEvent, got %T", events[0])
		}
	})

}

func TestImageGraph_SetNodePreview(t *testing.T) {
	t.Run("sets preview image for existing node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")

		imageID := imagegraph.MustNewImageID()

		err := ig.SetNodePreview(nodeID, imageID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		node, _ := ig.Nodes.Get(nodeID)
		if node.Preview != imageID {
			t.Errorf("expected preview %v, got %v", imageID, node.Preview)
		}
	})

	t.Run("returns error for non-existent node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		imageID := imagegraph.MustNewImageID()

		err := ig.SetNodePreview(nodeID, imageID)

		if err == nil {
			t.Fatal("expected error for non-existent node, got nil")
		}
	})

	t.Run("emits NodePreviewSet event", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")
		ig.ResetEvents()

		imageID := imagegraph.MustNewImageID()

		err := ig.SetNodePreview(nodeID, imageID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := ig.GetEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		if _, ok := events[0].(*imagegraph.NodePreviewSetEvent); !ok {
			t.Errorf("expected NodePreviewSetEvent, got %T", events[0])
		}
	})

	t.Run("can update preview to different image", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")

		imageID1 := imagegraph.MustNewImageID()
		imageID2 := imagegraph.MustNewImageID()

		ig.SetNodePreview(nodeID, imageID1)
		err := ig.SetNodePreview(nodeID, imageID2)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		node, _ := ig.Nodes.Get(nodeID)
		if node.Preview != imageID2 {
			t.Errorf("expected preview %v, got %v", imageID2, node.Preview)
		}
	})

	t.Run("UnsetNodePreview clears preview", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")

		imageID := imagegraph.MustNewImageID()
		ig.SetNodePreview(nodeID, imageID)

		err := ig.UnsetNodePreview(nodeID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		node, _ := ig.Nodes.Get(nodeID)
		if !node.Preview.IsNil() {
			t.Errorf("expected nil preview, got %v", node.Preview)
		}
	})

	t.Run("emits NodePreviewUnset event when unsetting", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")

		imageID := imagegraph.MustNewImageID()
		ig.SetNodePreview(nodeID, imageID)
		ig.ResetEvents()

		err := ig.UnsetNodePreview(nodeID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := ig.GetEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		if _, ok := events[0].(*imagegraph.NodePreviewUnsetEvent); !ok {
			t.Errorf("expected NodePreviewUnsetEvent, got %T", events[0])
		}
	})
}

func TestImageGraph_UnsetNodePreview(t *testing.T) {
	t.Run("unsets preview for existing node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")

		imageID := imagegraph.MustNewImageID()
		ig.SetNodePreview(nodeID, imageID)

		err := ig.UnsetNodePreview(nodeID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		node, _ := ig.Nodes.Get(nodeID)
		if !node.Preview.IsNil() {
			t.Errorf("expected nil preview, got %v", node.Preview)
		}
	})

	t.Run("returns error for non-existent node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()

		err := ig.UnsetNodePreview(nodeID)

		if err == nil {
			t.Fatal("expected error for non-existent node, got nil")
		}
	})

	t.Run("emits NodePreviewUnset event", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")

		imageID := imagegraph.MustNewImageID()
		ig.SetNodePreview(nodeID, imageID)
		ig.ResetEvents()

		err := ig.UnsetNodePreview(nodeID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := ig.GetEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		if _, ok := events[0].(*imagegraph.NodePreviewUnsetEvent); !ok {
			t.Errorf("expected NodePreviewUnsetEvent, got %T", events[0])
		}
	})

	t.Run("works on node without preview set", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")

		err := ig.UnsetNodePreview(nodeID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		node, _ := ig.Nodes.Get(nodeID)
		if !node.Preview.IsNil() {
			t.Errorf("expected nil preview, got %v", node.Preview)
		}
	})
}
