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

		if createdEvent.Type != "Created" {
			t.Fatalf("expected Created event, got %s", createdEvent.Type)
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

	t.Run("returns error for nil ID", func(t *testing.T) {
		ig, err := imagegraph.NewImageGraph(imagegraph.ImageGraphID{}, "test-graph")

		if err == nil {
			t.Fatal("expected error for nil ID, got nil")
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

		err := ig.AddNode(nodeID, imagegraph.NodeTypeInput, "input-node")

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

		err := ig.AddNode(imagegraph.MustNewNodeID(), imagegraph.NodeTypeInput, "node")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// NodeCreatedEvent, NodeNeedsOutputsEvent (Input has no inputs), NodeAddedEvent = 3 events
		if ig.Version != initialVersion+3 {
			t.Errorf("expected version %v, got %v", initialVersion+3, ig.Version)
		}
	})

	t.Run("emits NodeAdded event", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		ig.ResetEvents()
		nodeID := imagegraph.MustNewNodeID()

		err := ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := ig.GetEvents()
		// Should emit NodeCreatedEvent, NodeNeedsOutputsEvent (Input has no inputs), and NodeAddedEvent
		if len(events) != 3 {
			t.Fatalf("expected 3 events, got %d", len(events))
		}

		if _, ok := events[0].(*imagegraph.NodeCreatedEvent); !ok {
			t.Errorf("expected first event to be NodeCreatedEvent, got %T", events[0])
		}

		if _, ok := events[1].(*imagegraph.NodeNeedsOutputsEvent); !ok {
			t.Errorf("expected second event to be NodeNeedsOutputsEvent, got %T", events[1])
		}

		if _, ok := events[2].(*imagegraph.NodeAddedEvent); !ok {
			t.Errorf("expected third event to be NodeAddedEvent, got %T", events[2])
		}
	})

	t.Run("returns error for duplicate node ID", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()

		err := ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node1")
		if err != nil {
			t.Fatalf("expected no error on first add, got %v", err)
		}

		err = ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node2")
		if err == nil {
			t.Fatal("expected error for duplicate node ID, got nil")
		}
	})

	t.Run("returns error for invalid node type", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")

		err := ig.AddNode(imagegraph.MustNewNodeID(), imagegraph.NodeTypeNone, "node")

		if err == nil {
			t.Fatal("expected error for invalid node type, got nil")
		}
	})

	t.Run("returns error for nil node ID", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")

		err := ig.AddNode(imagegraph.NodeID{}, imagegraph.NodeTypeInput, "node")

		if err == nil {
			t.Fatal("expected error for nil node ID, got nil")
		}
	})

	t.Run("returns error for empty node name", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")

		err := ig.AddNode(imagegraph.MustNewNodeID(), imagegraph.NodeTypeOutput, "")

		if err == nil {
			t.Fatal("expected error for empty node name, got nil")
		}
	})

	t.Run("can add multiple nodes", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")

		err := ig.AddNode(imagegraph.MustNewNodeID(), imagegraph.NodeTypeInput, "node1")
		if err != nil {
			t.Fatalf("expected no error adding node1, got %v", err)
		}

		err = ig.AddNode(imagegraph.MustNewNodeID(), imagegraph.NodeTypeInput, "node2")
		if err != nil {
			t.Fatalf("expected no error adding node2, got %v", err)
		}

		if len(ig.Nodes) != 2 {
			t.Errorf("expected 2 nodes, got %d", len(ig.Nodes))
		}
	})
}


func TestImageGraph_SetNodeName(t *testing.T) {
	t.Run("sets name for existing node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "old-name")

		err := ig.SetNodeName(nodeID, "new-name")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		node, _ := ig.Nodes.Get(nodeID)
		if node.Name != "new-name" {
			t.Errorf("expected name %q, got %q", "new-name", node.Name)
		}
	})

	t.Run("returns error for non-existent node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()

		err := ig.SetNodeName(nodeID, "new-name")

		if err == nil {
			t.Fatal("expected error for non-existent node, got nil")
		}
	})

	t.Run("returns error for nil node ID", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")

		err := ig.SetNodeName(imagegraph.NodeID{}, "new-name")

		if err == nil {
			t.Fatal("expected error for nil node ID, got nil")
		}
	})

	t.Run("returns error for empty name", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeOutput, "old-name")

		err := ig.SetNodeName(nodeID, "")

		if err == nil {
			t.Fatal("expected error for empty name, got nil")
		}
	})

	t.Run("emits NodeNameSet event", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "old-name")
		ig.ResetEvents()

		err := ig.SetNodeName(nodeID, "new-name")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := ig.GetEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		nameSetEvent, ok := events[0].(*imagegraph.NodeNameSetEvent)
		if !ok {
			t.Errorf("expected NodeNameSetEvent, got %T", events[0])
		}

		if nameSetEvent.Name != "new-name" {
			t.Errorf("expected event name %q, got %q", "new-name", nameSetEvent.Name)
		}
	})

	t.Run("can update name multiple times", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "name1")

		ig.SetNodeName(nodeID, "name2")
		err := ig.SetNodeName(nodeID, "name3")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		node, _ := ig.Nodes.Get(nodeID)
		if node.Name != "name3" {
			t.Errorf("expected name %q, got %q", "name3", node.Name)
		}
	})

	t.Run("increments version when name is set", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "old-name")
		initialVersion := ig.Version

		err := ig.SetNodeName(nodeID, "new-name")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if ig.Version != initialVersion+1 {
			t.Errorf("expected version %v, got %v", initialVersion+1, ig.Version)
		}
	})
}

func TestImageGraph_SetNodePreview(t *testing.T) {
	t.Run("sets preview image for existing node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node")

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
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node")
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
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node")

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
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node")

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
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node")

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

	t.Run("returns error for nil node ID", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		imageID := imagegraph.MustNewImageID()

		err := ig.SetNodePreview(imagegraph.NodeID{}, imageID)

		if err == nil {
			t.Fatal("expected error for nil node ID, got nil")
		}
	})
}

func TestImageGraph_UnsetNodePreview(t *testing.T) {
	t.Run("unsets preview for existing node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node")

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
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node")

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
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node")

		err := ig.UnsetNodePreview(nodeID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		node, _ := ig.Nodes.Get(nodeID)
		if !node.Preview.IsNil() {
			t.Errorf("expected nil preview, got %v", node.Preview)
		}
	})

	t.Run("returns error for nil node ID", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")

		err := ig.UnsetNodePreview(imagegraph.NodeID{})

		if err == nil {
			t.Fatal("expected error for nil node ID, got nil")
		}
	})
}

func TestImageGraph_RemoveNode(t *testing.T) {
	t.Run("removes node from graph", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node")

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

	t.Run("returns error for nil node ID", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")

		err := ig.RemoveNode(imagegraph.NodeID{})

		if err == nil {
			t.Fatal("expected error for nil node ID, got nil")
		}
	})

	t.Run("emits NodeRemoved event", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node")
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
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node")
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
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		// Connect input → scale
		ig.ConnectNodes(inputID, "original", resizeID, "original")

		// Remove scale node
		ig.ResetEvents()
		err := ig.RemoveNode(resizeID)

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
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		// Set an image before connecting to verify it gets unset
		imageID := imagegraph.MustNewImageID()
		ig.SetNodeOutputImage(inputID, "original", imageID)

		// Connect input → resize (this will synchronously propagate the image)
		ig.ConnectNodes(inputID, "original", resizeID, "original")

		// Remove input node
		ig.ResetEvents()
		err := ig.RemoveNode(inputID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify downstream node's input is disconnected
		resizeNode, _ := ig.Nodes.Get(resizeID)
		if resizeNode.Inputs["original"].Connected {
			t.Error("expected downstream input to be disconnected")
		}

		// Verify downstream node's input image is unset
		if resizeNode.Inputs["original"].HasImage() {
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
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "node")

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
		ig.AddNode(nodeAID, imagegraph.NodeTypeInput, "nodeA")
		ig.AddNode(nodeBID, imagegraph.NodeTypeInput, "nodeB")

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
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		err := ig.ConnectNodes(inputID, "original", resizeID, "original")

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
		resizeNode, _ := ig.Nodes.Get(resizeID)
		input := resizeNode.Inputs["original"]
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
		err := ig.AddNode(nodeID, imagegraph.NodeTypeResize, "node")
		if err != nil {
			t.Fatalf("failed to add node: %v", err)
		}

		err = ig.ConnectNodes(nodeID, "resized", nodeID, "original")

		if err == nil {
			t.Fatal("expected error for self-connection, got nil")
		}
	})

	t.Run("returns error for cycle", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		node1ID := imagegraph.MustNewNodeID()
		node2ID := imagegraph.MustNewNodeID()
		ig.AddNode(node1ID, imagegraph.NodeTypeResize, "node1")
		ig.AddNode(node2ID, imagegraph.NodeTypeResize, "node2")

		// Create A → B
		ig.ConnectNodes(node1ID, "resized", node2ID, "original")

		// Try B → A (would create cycle)
		err := ig.ConnectNodes(node2ID, "resized", node1ID, "original")

		if err == nil {
			t.Fatal("expected error for cycle, got nil")
		}
	})

	t.Run("returns error for non-existent from node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		fakeID := imagegraph.MustNewNodeID()
		err := ig.ConnectNodes(fakeID, "original", resizeID, "original")

		if err == nil {
			t.Fatal("expected error for non-existent from node, got nil")
		}
	})

	t.Run("returns error for non-existent to node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")

		fakeID := imagegraph.MustNewNodeID()
		err := ig.ConnectNodes(inputID, "original", fakeID, "original")

		if err == nil {
			t.Fatal("expected error for non-existent to node, got nil")
		}
	})

	t.Run("returns error for invalid output name", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		err := ig.ConnectNodes(inputID, "invalid", resizeID, "original")

		if err == nil {
			t.Fatal("expected error for invalid output name, got nil")
		}
	})

	t.Run("returns error for invalid input name", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		err := ig.ConnectNodes(inputID, "original", resizeID, "invalid")

		if err == nil {
			t.Fatal("expected error for invalid input name, got nil")
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		err := ig.ConnectNodes(inputID, "original", resizeID, "original")
		if err != nil {
			t.Fatalf("expected no error on first connect, got %v", err)
		}

		ig.ResetEvents()
		err = ig.ConnectNodes(inputID, "original", resizeID, "original")
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
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(input1ID, imagegraph.NodeTypeInput, "input1")
		ig.AddNode(input2ID, imagegraph.NodeTypeInput, "input2")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		// Connect input1 → scale
		ig.ConnectNodes(input1ID, "original", resizeID, "original")

		// Connect input2 → scale (should disconnect input1)
		err := ig.ConnectNodes(input2ID, "original", resizeID, "original")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify input1 is disconnected
		input1Node, _ := ig.Nodes.Get(input1ID)
		if len(input1Node.Outputs["original"].Connections) != 0 {
			t.Error("expected input1 to be disconnected")
		}

		// Verify input2 is connected
		resizeNode, _ := ig.Nodes.Get(resizeID)
		if resizeNode.Inputs["original"].InputConnection.NodeID != input2ID {
			t.Error("expected scale to be connected to input2")
		}
	})

	t.Run("emits connection events", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")
		ig.ResetEvents()

		err := ig.ConnectNodes(inputID, "original", resizeID, "original")

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
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(input1ID, imagegraph.NodeTypeInput, "input1")
		ig.AddNode(input2ID, imagegraph.NodeTypeInput, "input2")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		ig.ConnectNodes(input1ID, "original", resizeID, "original")
		ig.ResetEvents()

		ig.ConnectNodes(input2ID, "original", resizeID, "original")

		events := ig.GetEvents()
		// Should have: InputDisconnected, OutputDisconnected, OutputConnected, InputConnected
		if len(events) != 4 {
			t.Fatalf("expected 4 events, got %d", len(events))
		}
	})

	t.Run("returns error for nil fromNodeID", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeResize, "resize")

		err := ig.ConnectNodes(imagegraph.NodeID{}, "original", nodeID, "original")

		if err == nil {
			t.Fatal("expected error for nil fromNodeID, got nil")
		}
	})

	t.Run("returns error for nil toNodeID", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "input")

		err := ig.ConnectNodes(nodeID, "original", imagegraph.NodeID{}, "original")

		if err == nil {
			t.Fatal("expected error for nil toNodeID, got nil")
		}
	})
}

func TestImageGraph_DisconnectNodes(t *testing.T) {
	t.Run("disconnects nodes successfully", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		// Connect nodes first
		ig.ConnectNodes(inputID, "original", resizeID, "original")

		// Disconnect them
		err := ig.DisconnectNodes(inputID, "original", resizeID, "original")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify output connection removed
		inputNode, _ := ig.Nodes.Get(inputID)
		output := inputNode.Outputs["original"]
		if len(output.Connections) != 0 {
			t.Errorf("expected 0 output connections, got %d", len(output.Connections))
		}

		// Verify input connection removed
		resizeNode, _ := ig.Nodes.Get(resizeID)
		input := resizeNode.Inputs["original"]
		if input.Connected {
			t.Error("expected input to be disconnected")
		}
	})

	t.Run("returns error for non-existent from node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		fakeID := imagegraph.MustNewNodeID()
		err := ig.DisconnectNodes(fakeID, "original", resizeID, "original")

		if err == nil {
			t.Fatal("expected error for non-existent from node, got nil")
		}
	})

	t.Run("returns error for non-existent to node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")

		fakeID := imagegraph.MustNewNodeID()
		err := ig.DisconnectNodes(inputID, "original", fakeID, "original")

		if err == nil {
			t.Fatal("expected error for non-existent to node, got nil")
		}
	})

	t.Run("returns error for invalid output name", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		err := ig.DisconnectNodes(inputID, "invalid", resizeID, "original")

		if err == nil {
			t.Fatal("expected error for invalid output name, got nil")
		}
	})

	t.Run("returns error for invalid input name", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		err := ig.DisconnectNodes(inputID, "original", resizeID, "invalid")

		if err == nil {
			t.Fatal("expected error for invalid input name, got nil")
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		// Connect nodes
		ig.ConnectNodes(inputID, "original", resizeID, "original")

		// Disconnect once
		err := ig.DisconnectNodes(inputID, "original", resizeID, "original")
		if err != nil {
			t.Fatalf("expected no error on first disconnect, got %v", err)
		}

		// Disconnect again (should be no-op)
		ig.ResetEvents()
		err = ig.DisconnectNodes(inputID, "original", resizeID, "original")
		if err != nil {
			t.Fatalf("expected no error on second disconnect, got %v", err)
		}

		events := ig.GetEvents()
		if len(events) != 0 {
			t.Errorf("expected 0 events on duplicate disconnect, got %d", len(events))
		}
	})

	t.Run("emits disconnection events", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		// Connect nodes first
		ig.ConnectNodes(inputID, "original", resizeID, "original")
		ig.ResetEvents()

		// Disconnect them
		err := ig.DisconnectNodes(inputID, "original", resizeID, "original")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := ig.GetEvents()
		if len(events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(events))
		}

		if _, ok := events[0].(*imagegraph.NodeOutputDisconnectedEvent); !ok {
			t.Errorf("expected first event to be NodeOutputDisconnectedEvent, got %T", events[0])
		}

		if _, ok := events[1].(*imagegraph.NodeInputDisconnectedEvent); !ok {
			t.Errorf("expected second event to be NodeInputDisconnectedEvent, got %T", events[1])
		}
	})

	t.Run("unsets input image when disconnecting", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		// Set image first, then connect (ConnectNodes will synchronously propagate)
		imageID := imagegraph.MustNewImageID()
		ig.SetNodeOutputImage(inputID, "original", imageID)
		ig.ConnectNodes(inputID, "original", resizeID, "original")

		// Verify image was set
		resizeNode, _ := ig.Nodes.Get(resizeID)
		if !resizeNode.Inputs["original"].HasImage() {
			t.Fatal("expected input to have image set")
		}

		ig.ResetEvents()

		// Disconnect nodes
		err := ig.DisconnectNodes(inputID, "original", resizeID, "original")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify input image was unset
		resizeNode, _ = ig.Nodes.Get(resizeID)
		if resizeNode.Inputs["original"].HasImage() {
			t.Error("expected input image to be unset after disconnection")
		}

		// Verify image unset event was emitted
		events := ig.GetEvents()
		foundImageUnset := false
		for _, event := range events {
			if _, ok := event.(*imagegraph.NodeInputImageUnsetEvent); ok {
				foundImageUnset = true
				break
			}
		}
		if !foundImageUnset {
			t.Error("expected NodeInputImageUnsetEvent to be emitted")
		}
	})

	t.Run("handles multiple connections from same output", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resize1ID := imagegraph.MustNewNodeID()
		resize2ID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resize1ID, imagegraph.NodeTypeResize, "scale1")
		ig.AddNode(resize2ID, imagegraph.NodeTypeResize, "scale2")

		// Connect input to both scale nodes
		ig.ConnectNodes(inputID, "original", resize1ID, "original")
		ig.ConnectNodes(inputID, "original", resize2ID, "original")

		// Disconnect one connection
		err := ig.DisconnectNodes(inputID, "original", resize1ID, "original")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify scale1 is disconnected
		resize1Node, _ := ig.Nodes.Get(resize1ID)
		if resize1Node.Inputs["original"].Connected {
			t.Error("expected scale1 input to be disconnected")
		}

		// Verify scale2 is still connected
		resize2Node, _ := ig.Nodes.Get(resize2ID)
		if !resize2Node.Inputs["original"].Connected {
			t.Error("expected scale2 input to still be connected")
		}

		// Verify input node still has one connection
		inputNode, _ := ig.Nodes.Get(inputID)
		if len(inputNode.Outputs["original"].Connections) != 1 {
			t.Errorf("expected 1 output connection remaining, got %d", len(inputNode.Outputs["original"].Connections))
		}
	})

	t.Run("returns error for nil fromNodeID", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeResize, "resize")

		err := ig.DisconnectNodes(imagegraph.NodeID{}, "original", nodeID, "original")

		if err == nil {
			t.Fatal("expected error for nil fromNodeID, got nil")
		}
	})

	t.Run("returns error for nil toNodeID", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "input")

		err := ig.DisconnectNodes(nodeID, "original", imagegraph.NodeID{}, "original")

		if err == nil {
			t.Fatal("expected error for nil toNodeID, got nil")
		}
	})
}

func TestImageGraph_SetNodeOutputImage(t *testing.T) {
	t.Run("sets output image for existing node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "input")

		imageID := imagegraph.MustNewImageID()

		err := ig.SetNodeOutputImage(nodeID, "original", imageID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		node, _ := ig.Nodes.Get(nodeID)
		output := node.Outputs["original"]
		if output.ImageID != imageID {
			t.Errorf("expected output image %v, got %v", imageID, output.ImageID)
		}
	})

	t.Run("returns error for non-existent node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		fakeID := imagegraph.MustNewNodeID()
		imageID := imagegraph.MustNewImageID()

		err := ig.SetNodeOutputImage(fakeID, "original", imageID)

		if err == nil {
			t.Fatal("expected error for non-existent node, got nil")
		}
	})

	t.Run("returns error for invalid output name", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "input")

		imageID := imagegraph.MustNewImageID()

		err := ig.SetNodeOutputImage(nodeID, "invalid", imageID)

		if err == nil {
			t.Fatal("expected error for invalid output name, got nil")
		}
	})

	t.Run("does not propagate image synchronously (event-driven)", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		// Connect input → resize
		ig.ConnectNodes(inputID, "original", resizeID, "original")

		imageID := imagegraph.MustNewImageID()

		err := ig.SetNodeOutputImage(inputID, "original", imageID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify output is set on the source node
		inputNode, _ := ig.Nodes.Get(inputID)
		outputImage, _ := inputNode.GetOutputImage("original")
		if outputImage != imageID {
			t.Errorf("expected output image %v, got %v", imageID, outputImage)
		}

		// Verify downstream input does NOT have the image yet (propagation is event-driven)
		resizeNode, _ := ig.Nodes.Get(resizeID)
		input := resizeNode.Inputs["original"]
		if input.HasImage() {
			t.Error("expected downstream input to NOT have image (propagation is event-driven)")
		}
	})

	t.Run("does not propagate to multiple downstream nodes synchronously", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resize1ID := imagegraph.MustNewNodeID()
		resize2ID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resize1ID, imagegraph.NodeTypeResize, "scale1")
		ig.AddNode(resize2ID, imagegraph.NodeTypeResize, "scale2")

		// Connect input to both scale nodes
		ig.ConnectNodes(inputID, "original", resize1ID, "original")
		ig.ConnectNodes(inputID, "original", resize2ID, "original")

		imageID := imagegraph.MustNewImageID()

		err := ig.SetNodeOutputImage(inputID, "original", imageID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify output is set
		inputNode, _ := ig.Nodes.Get(inputID)
		outputImage, _ := inputNode.GetOutputImage("original")
		if outputImage != imageID {
			t.Errorf("expected output image %v, got %v", imageID, outputImage)
		}

		// Verify downstream inputs do NOT have the image yet (propagation is event-driven)
		resize1Node, _ := ig.Nodes.Get(resize1ID)
		if resize1Node.Inputs["original"].HasImage() {
			t.Error("expected scale1 input to NOT have image (propagation is event-driven)")
		}

		resize2Node, _ := ig.Nodes.Get(resize2ID)
		if resize2Node.Inputs["original"].HasImage() {
			t.Error("expected scale2 input to NOT have image (propagation is event-driven)")
		}
	})

	t.Run("emits NodeOutputImageSet event", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "input")
		ig.ResetEvents()

		imageID := imagegraph.MustNewImageID()

		err := ig.SetNodeOutputImage(nodeID, "original", imageID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := ig.GetEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		if _, ok := events[0].(*imagegraph.NodeOutputImageSetEvent); !ok {
			t.Errorf("expected NodeOutputImageSetEvent, got %T", events[0])
		}
	})

	t.Run("emits only NodeOutputImageSet event (no downstream events)", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		ig.ConnectNodes(inputID, "original", resizeID, "original")
		ig.ResetEvents()

		imageID := imagegraph.MustNewImageID()

		err := ig.SetNodeOutputImage(inputID, "original", imageID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := ig.GetEvents()
		// Should only emit NodeOutputImageSetEvent (propagation is event-driven)
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		if _, ok := events[0].(*imagegraph.NodeOutputImageSetEvent); !ok {
			t.Errorf("expected NodeOutputImageSetEvent, got %T", events[0])
		}
	})

	t.Run("can update output image to different image", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "input")

		imageID1 := imagegraph.MustNewImageID()
		imageID2 := imagegraph.MustNewImageID()

		ig.SetNodeOutputImage(nodeID, "original", imageID1)
		err := ig.SetNodeOutputImage(nodeID, "original", imageID2)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		node, _ := ig.Nodes.Get(nodeID)
		if node.Outputs["original"].ImageID != imageID2 {
			t.Errorf("expected output image %v, got %v", imageID2, node.Outputs["original"].ImageID)
		}
	})

	t.Run("does not affect unconnected nodes", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		// Do NOT connect the nodes

		imageID := imagegraph.MustNewImageID()

		err := ig.SetNodeOutputImage(inputID, "original", imageID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify unconnected node doesn't have the image
		resizeNode, _ := ig.Nodes.Get(resizeID)
		if resizeNode.Inputs["original"].HasImage() {
			t.Error("expected unconnected input to not have image")
		}
	})

	t.Run("returns error for nil node ID", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		imageID := imagegraph.MustNewImageID()

		err := ig.SetNodeOutputImage(imagegraph.NodeID{}, "original", imageID)

		if err == nil {
			t.Fatal("expected error for nil node ID, got nil")
		}
	})
}

func TestImageGraph_UnsetNodeOutputImage(t *testing.T) {
	t.Run("unsets output image for existing node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "input")

		imageID := imagegraph.MustNewImageID()
		ig.SetNodeOutputImage(nodeID, "original", imageID)

		err := ig.UnsetNodeOutputImage(nodeID, "original")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		node, _ := ig.Nodes.Get(nodeID)
		if node.Outputs["original"].HasImage() {
			t.Error("expected output image to be unset")
		}
	})

	t.Run("returns error for non-existent node", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		fakeID := imagegraph.MustNewNodeID()

		err := ig.UnsetNodeOutputImage(fakeID, "original")

		if err == nil {
			t.Fatal("expected error for non-existent node, got nil")
		}
	})

	t.Run("returns error for invalid output name", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "input")

		err := ig.UnsetNodeOutputImage(nodeID, "invalid")

		if err == nil {
			t.Fatal("expected error for invalid output name, got nil")
		}
	})

	t.Run("unsets output image without propagating to downstream", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		// Set image first, then connect (ConnectNodes will synchronously propagate)
		imageID := imagegraph.MustNewImageID()
		ig.SetNodeOutputImage(inputID, "original", imageID)
		ig.ConnectNodes(inputID, "original", resizeID, "original")

		err := ig.UnsetNodeOutputImage(inputID, "original")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify output is unset but downstream is NOT propagated (event-driven propagation will handle it)
		inputNode, _ := ig.Nodes.Get(inputID)
		if inputNode.Outputs["original"].HasImage() {
			t.Error("expected output image to be unset")
		}

		// Downstream should still have the image (propagation will be event-driven)
		resizeNode, _ := ig.Nodes.Get(resizeID)
		if !resizeNode.Inputs["original"].HasImage() {
			t.Error("expected downstream input to still have image (propagation is event-driven)")
		}
	})

	t.Run("unsets output without propagating to multiple downstream nodes", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resize1ID := imagegraph.MustNewNodeID()
		resize2ID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resize1ID, imagegraph.NodeTypeResize, "scale1")
		ig.AddNode(resize2ID, imagegraph.NodeTypeResize, "scale2")

		// Set image first, then connect (ConnectNodes will synchronously propagate)
		imageID := imagegraph.MustNewImageID()
		ig.SetNodeOutputImage(inputID, "original", imageID)
		ig.ConnectNodes(inputID, "original", resize1ID, "original")
		ig.ConnectNodes(inputID, "original", resize2ID, "original")

		err := ig.UnsetNodeOutputImage(inputID, "original")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify both downstream inputs still have images (propagation is event-driven)
		resize1Node, _ := ig.Nodes.Get(resize1ID)
		if !resize1Node.Inputs["original"].HasImage() {
			t.Error("expected scale1 input to still have image (propagation is event-driven)")
		}

		resize2Node, _ := ig.Nodes.Get(resize2ID)
		if !resize2Node.Inputs["original"].HasImage() {
			t.Error("expected scale2 input to still have image (propagation is event-driven)")
		}
	})

	t.Run("emits NodeOutputImageUnset event", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "input")

		imageID := imagegraph.MustNewImageID()
		ig.SetNodeOutputImage(nodeID, "original", imageID)
		ig.ResetEvents()

		err := ig.UnsetNodeOutputImage(nodeID, "original")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := ig.GetEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		if _, ok := events[0].(*imagegraph.NodeOutputImageUnsetEvent); !ok {
			t.Errorf("expected NodeOutputImageUnsetEvent, got %T", events[0])
		}
	})

	t.Run("emits only NodeOutputImageUnset event without downstream propagation", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		inputID := imagegraph.MustNewNodeID()
		resizeID := imagegraph.MustNewNodeID()
		ig.AddNode(inputID, imagegraph.NodeTypeInput, "input")
		ig.AddNode(resizeID, imagegraph.NodeTypeResize, "resize")

		ig.ConnectNodes(inputID, "original", resizeID, "original")
		imageID := imagegraph.MustNewImageID()
		ig.SetNodeOutputImage(inputID, "original", imageID)
		ig.ResetEvents()

		err := ig.UnsetNodeOutputImage(inputID, "original")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := ig.GetEvents()
		// Should only emit NodeOutputImageUnsetEvent, no downstream propagation
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		if _, ok := events[0].(*imagegraph.NodeOutputImageUnsetEvent); !ok {
			t.Errorf("expected NodeOutputImageUnsetEvent, got %T", events[0])
		}
	})

	t.Run("works on node without image set", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")
		nodeID := imagegraph.MustNewNodeID()
		ig.AddNode(nodeID, imagegraph.NodeTypeInput, "input")

		err := ig.UnsetNodeOutputImage(nodeID, "original")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		node, _ := ig.Nodes.Get(nodeID)
		if node.Outputs["original"].HasImage() {
			t.Error("expected output to not have image")
		}
	})

	t.Run("returns error for nil node ID", func(t *testing.T) {
		ig, _ := imagegraph.NewImageGraph(imagegraph.MustNewImageGraphID(), "test")

		err := ig.UnsetNodeOutputImage(imagegraph.NodeID{}, "original")

		if err == nil {
			t.Fatal("expected error for nil node ID, got nil")
		}
	})
}
