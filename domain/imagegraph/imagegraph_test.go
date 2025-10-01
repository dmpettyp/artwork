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
		config := `{}`

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
		config := `{}`

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

	t.Run("validates required fields for NodeTypeScale", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeScale, "node", `{"factor": 2.0}`)

		node, _ := ig.Nodes.Get(nodeID)

		// Missing required field
		err := node.SetConfig(`{}`)
		if err == nil {
			t.Fatal("expected error for missing required field, got nil")
		}

		// Valid config
		err = node.SetConfig(`{"factor": 2.5}`)
		if err != nil {
			t.Fatalf("expected no error for valid config, got %v", err)
		}
	})

	t.Run("validates field types for NodeTypeScale", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeScale, "node", `{"factor": 2.0}`)

		node, _ := ig.Nodes.Get(nodeID)

		// Wrong type - string instead of float
		err := node.SetConfig(`{"factor": "2.0"}`)
		if err == nil {
			t.Fatal("expected error for wrong field type, got nil")
		}

		// Valid float
		err = node.SetConfig(`{"factor": 1.5}`)
		if err != nil {
			t.Fatalf("expected no error for valid float, got %v", err)
		}

		// Valid integer (also acceptable as float)
		err = node.SetConfig(`{"factor": 2}`)
		if err != nil {
			t.Fatalf("expected no error for integer as float, got %v", err)
		}
	})

	t.Run("rejects unknown fields", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeScale, "node", `{"factor": 2.0}`)

		node, _ := ig.Nodes.Get(nodeID)

		err := node.SetConfig(`{"factor": 2.0, "unknown": "value"}`)
		if err == nil {
			t.Fatal("expected error for unknown field, got nil")
		}
	})

	t.Run("allows empty config for NodeTypeInput", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")

		node, _ := ig.Nodes.Get(nodeID)

		err := node.SetConfig(`{}`)
		if err != nil {
			t.Fatalf("expected no error for empty config on NodeTypeInput, got %v", err)
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

func TestImageGraph_RemoveNode(t *testing.T) {
	t.Run("removes node from graph", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")

		err := ig.RemoveNode(nodeID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(ig.Nodes) != 0 {
			t.Errorf("expected 0 nodes, got %d", len(ig.Nodes))
		}

		_, exists := ig.Nodes.Get(nodeID)
		if exists {
			t.Error("expected node to not exist after removal")
		}
	})

	t.Run("returns error for non-existent node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()

		err := ig.RemoveNode(nodeID)

		if err == nil {
			t.Fatal("expected error for non-existent node, got nil")
		}
	})

	t.Run("emits NodeRemoved event", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")
		ig.ResetEvents()

		err := ig.RemoveNode(nodeID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := ig.GetEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		if _, ok := events[0].(*imagegraph.NodeRemovedEvent); !ok {
			t.Errorf("expected NodeRemovedEvent, got %T", events[0])
		}
	})

	t.Run("increments version", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")
		initialVersion := ig.Version

		err := ig.RemoveNode(nodeID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if ig.Version != initialVersion+1 {
			t.Errorf("expected version %v, got %v", initialVersion+1, ig.Version)
		}
	})

	t.Run("disconnects upstream connections", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		scaleID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input", "{}")
		ig.AddNode(scaleID, imagegraph.NodeTypeScale, "scale", `{"factor": 2.0}`)

		// Connect input → scale
		ig.ConnectNodes(inputID, "original", scaleID, "original")

		// Remove scale node
		ig.ResetEvents()
		err := ig.RemoveNode(scaleID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify upstream node's output is disconnected
		inputNode, _ := ig.Nodes.Get(inputID)
		if len(inputNode.Outputs["original"].Connections) != 0 {
			t.Error("expected upstream output to be disconnected")
		}

		// Verify disconnection event was emitted
		events := ig.GetEvents()
		foundDisconnectEvent := false
		for _, event := range events {
			if _, ok := event.(*imagegraph.NodeOutputDisconnectedEvent); ok {
				foundDisconnectEvent = true
				break
			}
		}
		if !foundDisconnectEvent {
			t.Error("expected NodeOutputDisconnectedEvent to be emitted")
		}
	})

	t.Run("disconnects downstream connections", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		scaleID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input", "{}")
		ig.AddNode(scaleID, imagegraph.NodeTypeScale, "scale", `{"factor": 2.0}`)

		// Connect input → scale
		ig.ConnectNodes(inputID, "original", scaleID, "original")

		// Set an image on the connection to verify it gets unset
		imageID := imagegraph.MustNewImageID()
		ig.SetNodeOutputImage(inputID, "original", imageID)

		// Remove input node
		ig.ResetEvents()
		err := ig.RemoveNode(inputID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify downstream node's input is disconnected
		scaleNode, _ := ig.Nodes.Get(scaleID)
		if scaleNode.Inputs["original"].Connected {
			t.Error("expected downstream input to be disconnected")
		}

		// Verify downstream node's input image is unset
		if scaleNode.Inputs["original"].HasImage() {
			t.Error("expected downstream input image to be unset")
		}

		// Verify disconnection events were emitted
		events := ig.GetEvents()
		foundInputDisconnect := false
		foundImageUnset := false
		for _, event := range events {
			if _, ok := event.(*imagegraph.NodeInputDisconnectedEvent); ok {
				foundInputDisconnect = true
			}
			if _, ok := event.(*imagegraph.NodeInputImageUnsetEvent); ok {
				foundImageUnset = true
			}
		}
		if !foundInputDisconnect {
			t.Error("expected NodeInputDisconnectedEvent to be emitted")
		}
		if !foundImageUnset {
			t.Error("expected NodeInputImageUnsetEvent to be emitted")
		}
	})

	t.Run("handles node with no connections", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node", "{}")

		err := ig.RemoveNode(nodeID)

		if err != nil {
			t.Fatalf("expected no error for standalone node, got %v", err)
		}

		if len(ig.Nodes) != 0 {
			t.Errorf("expected 0 nodes, got %d", len(ig.Nodes))
		}
	})

	t.Run("can remove multiple nodes", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeAID := imagegraph.MustNewNodeID()
		nodeBID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeAID, imagegraph.NodeTypeInput, "nodeA", "{}")
		ig.AddNode(nodeBID, imagegraph.NodeTypeInput, "nodeB", "{}")

		err := ig.RemoveNode(nodeAID)
		if err != nil {
			t.Fatalf("expected no error removing nodeA, got %v", err)
		}

		err = ig.RemoveNode(nodeBID)
		if err != nil {
			t.Fatalf("expected no error removing nodeB, got %v", err)
		}

		if len(ig.Nodes) != 0 {
			t.Errorf("expected 0 nodes, got %d", len(ig.Nodes))
		}
	})
}

func TestImageGraph_ConnectNodes(t *testing.T) {
	t.Run("connects nodes successfully", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		scaleID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input", "{}")
		ig.AddNode(scaleID, imagegraph.NodeTypeScale, "scale", `{"factor": 2.0}`)

		err := ig.ConnectNodes(inputID, "original", scaleID, "original")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify output connection
		inputNode, _ := ig.Nodes.Get(inputID)
		output := inputNode.Outputs["original"]
		if len(output.Connections) != 1 {
			t.Errorf("expected 1 output connection, got %d", len(output.Connections))
		}

		// Verify input connection
		scaleNode, _ := ig.Nodes.Get(scaleID)
		input := scaleNode.Inputs["original"]
		if !input.Connected {
			t.Error("expected input to be connected")
		}
		if input.InputConnection.NodeID != inputID {
			t.Errorf("expected input connected to %v, got %v", inputID, input.InputConnection.NodeID)
		}
	})

	t.Run("returns error for self-connection", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeScale, "node", `{"factor": 2.0}`)

		err := ig.ConnectNodes(nodeID, "scaled", nodeID, "original")

		if err == nil {
			t.Fatal("expected error for self-connection, got nil")
		}
	})

	t.Run("returns error for cycle", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		node1ID := imagegraph.MustNewNodeID()
		node2ID := imagegraph.MustNewNodeID()
		ig.AddNode(node1ID, imagegraph.NodeTypeScale, "node1", `{"factor": 2.0}`)
		ig.AddNode(node2ID, imagegraph.NodeTypeScale, "node2", `{"factor": 2.0}`)

		// Create A → B
		ig.ConnectNodes(node1ID, "scaled", node2ID, "original")

		// Try B → A (would create cycle)
		err := ig.ConnectNodes(node2ID, "scaled", node1ID, "original")

		if err == nil {
			t.Fatal("expected error for cycle, got nil")
		}
	})

	t.Run("returns error for non-existent from node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		scaleID := imagegraph.MustNewNodeID()
		ig.AddNode(scaleID, imagegraph.NodeTypeScale, "scale", `{"factor": 2.0}`)

		fakeID := imagegraph.MustNewNodeID()
		err := ig.ConnectNodes(fakeID, "original", scaleID, "original")

		if err == nil {
			t.Fatal("expected error for non-existent from node, got nil")
		}
	})

	t.Run("returns error for non-existent to node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input", "{}")

		fakeID := imagegraph.MustNewNodeID()
		err := ig.ConnectNodes(inputID, "original", fakeID, "original")

		if err == nil {
			t.Fatal("expected error for non-existent to node, got nil")
		}
	})

	t.Run("returns error for invalid output name", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		scaleID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input", "{}")
		ig.AddNode(scaleID, imagegraph.NodeTypeScale, "scale", `{"factor": 2.0}`)

		err := ig.ConnectNodes(inputID, "invalid", scaleID, "original")

		if err == nil {
			t.Fatal("expected error for invalid output name, got nil")
		}
	})

	t.Run("returns error for invalid input name", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		scaleID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input", "{}")
		ig.AddNode(scaleID, imagegraph.NodeTypeScale, "scale", `{"factor": 2.0}`)

		err := ig.ConnectNodes(inputID, "original", scaleID, "invalid")

		if err == nil {
			t.Fatal("expected error for invalid input name, got nil")
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		scaleID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input", "{}")
		ig.AddNode(scaleID, imagegraph.NodeTypeScale, "scale", `{"factor": 2.0}`)

		err := ig.ConnectNodes(inputID, "original", scaleID, "original")
		if err != nil {
			t.Fatalf("expected no error on first connect, got %v", err)
		}

		ig.ResetEvents()
		err = ig.ConnectNodes(inputID, "original", scaleID, "original")
		if err != nil {
			t.Fatalf("expected no error on second connect, got %v", err)
		}

		events := ig.GetEvents()
		if len(events) != 0 {
			t.Errorf("expected 0 events on duplicate connect, got %d", len(events))
		}
	})

	t.Run("replaces existing connection", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		input1ID := imagegraph.MustNewNodeID()
		input2ID := imagegraph.MustNewNodeID()
		scaleID := imagegraph.MustNewNodeID()
		ig.AddNode(input1ID, imagegraph.NodeTypeInput, "input1", "{}")
		ig.AddNode(input2ID, imagegraph.NodeTypeInput, "input2", "{}")
		ig.AddNode(scaleID, imagegraph.NodeTypeScale, "scale", `{"factor": 2.0}`)

		// Connect input1 → scale
		ig.ConnectNodes(input1ID, "original", scaleID, "original")

		// Connect input2 → scale (should disconnect input1)
		err := ig.ConnectNodes(input2ID, "original", scaleID, "original")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify input1 is disconnected
		input1Node, _ := ig.Nodes.Get(input1ID)
		if len(input1Node.Outputs["original"].Connections) != 0 {
			t.Error("expected input1 to be disconnected")
		}

		// Verify input2 is connected
		scaleNode, _ := ig.Nodes.Get(scaleID)
		if scaleNode.Inputs["original"].InputConnection.NodeID != input2ID {
			t.Error("expected scale to be connected to input2")
		}
	})

	t.Run("emits connection events", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		scaleID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input", "{}")
		ig.AddNode(scaleID, imagegraph.NodeTypeScale, "scale", `{"factor": 2.0}`)
		ig.ResetEvents()

		err := ig.ConnectNodes(inputID, "original", scaleID, "original")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := ig.GetEvents()
		if len(events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(events))
		}

		if _, ok := events[0].(*imagegraph.NodeOutputConnectedEvent); !ok {
			t.Errorf("expected first event to be NodeOutputConnectedEvent, got %T", events[0])
		}

		if _, ok := events[1].(*imagegraph.NodeInputConnectedEvent); !ok {
			t.Errorf("expected second event to be NodeInputConnectedEvent, got %T", events[1])
		}
	})

	t.Run("emits disconnection events when replacing", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		input1ID := imagegraph.MustNewNodeID()
		input2ID := imagegraph.MustNewNodeID()
		scaleID := imagegraph.MustNewNodeID()
		ig.AddNode(input1ID, imagegraph.NodeTypeInput, "input1", "{}")
		ig.AddNode(input2ID, imagegraph.NodeTypeInput, "input2", "{}")
		ig.AddNode(scaleID, imagegraph.NodeTypeScale, "scale", `{"factor": 2.0}`)

		ig.ConnectNodes(input1ID, "original", scaleID, "original")
		ig.ResetEvents()

		ig.ConnectNodes(input2ID, "original", scaleID, "original")

		events := ig.GetEvents()
		// Should have: InputDisconnected, OutputDisconnected, OutputConnected, InputConnected
		if len(events) != 4 {
			t.Fatalf("expected 4 events, got %d", len(events))
		}
	})
}
