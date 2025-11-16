package postgres

import (
	"testing"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
	"github.com/dmpettyp/state"
)

func TestImageGraphRoundTrip(t *testing.T) {
	imageGraphID := imagegraph.MustNewImageGraphID()
	node1ID := imagegraph.MustNewNodeID()
	node2ID := imagegraph.MustNewNodeID()
	imageID1 := imagegraph.MustNewImageID()
	imageID2 := imagegraph.MustNewImageID()
	previewID := imagegraph.MustNewImageID()

	node1State, err := state.NewState(imagegraph.Generating)
	if err != nil {
		t.Fatalf("failed to create node1 state: %v", err)
	}

	node2State, err := state.NewState(imagegraph.Waiting)
	if err != nil {
		t.Fatalf("failed to create node2 state: %v", err)
	}

	original := &imagegraph.ImageGraph{
		ID:      imageGraphID,
		Name:    "Test Graph",
		Version: 5,
		Nodes: imagegraph.Nodes{
			node1ID: {
				ID:      node1ID,
				Version: 2,
				Type:    imagegraph.NodeTypeBlur,
				Name:    "Blur Node",
				State:   node1State,
				Config: imagegraph.NodeConfig{
					"radius": 5.0,
				},
				Preview: previewID,
				Inputs: imagegraph.Inputs{
					"input": {
						Name:      "input",
						ImageID:   imageID1,
						Connected: false,
					},
				},
				Outputs: imagegraph.Outputs{
					"output": {
						Name:    "output",
						ImageID: imageID2,
						Connections: map[imagegraph.OutputConnection]struct{}{
							{NodeID: node2ID, InputName: "input"}: {},
						},
					},
				},
			},
			node2ID: {
				ID:      node2ID,
				Version: 1,
				Type:    imagegraph.NodeTypeOutput,
				Name:    "Output Node",
				State:   node2State,
				Config:  imagegraph.NodeConfig{},
				Inputs: imagegraph.Inputs{
					"input": {
						Name:      "input",
						ImageID:   imageID2,
						Connected: true,
						InputConnection: imagegraph.InputConnection{
							NodeID:     node1ID,
							OutputName: "output",
						},
					},
				},
				Outputs: imagegraph.Outputs{},
			},
		},
	}

	row, err := serializeImageGraph(original)
	if err != nil {
		t.Fatalf("serializeImageGraph failed: %v", err)
	}

	deserialized, err := deserializeImageGraph(row)
	if err != nil {
		t.Fatalf("deserializeImageGraph failed: %v", err)
	}

	if deserialized.ID != original.ID {
		t.Errorf("ID mismatch: got %v, want %v", deserialized.ID, original.ID)
	}

	if deserialized.Name != original.Name {
		t.Errorf("Name mismatch: got %v, want %v", deserialized.Name, original.Name)
	}

	if deserialized.Version != original.Version {
		t.Errorf("Version mismatch: got %v, want %v", deserialized.Version, original.Version)
	}

	if len(deserialized.Nodes) != len(original.Nodes) {
		t.Fatalf("Nodes count mismatch: got %d, want %d", len(deserialized.Nodes), len(original.Nodes))
	}

	node1 := deserialized.Nodes[node1ID]
	if node1 == nil {
		t.Fatal("node1 not found")
	}

	if node1.Type != imagegraph.NodeTypeBlur {
		t.Errorf("node1 type mismatch: got %v, want %v", node1.Type, imagegraph.NodeTypeBlur)
	}

	if node1.State.Get() != imagegraph.Generating {
		t.Errorf("node1 state mismatch: got %v, want %v", node1.State.Get(), imagegraph.Generating)
	}

	if node1.Preview != previewID {
		t.Errorf("node1 preview mismatch: got %v, want %v", node1.Preview, previewID)
	}

	if radius, ok := node1.Config["radius"].(float64); !ok || radius != 5.0 {
		t.Errorf("node1 config radius mismatch: got %v, want 5.0", node1.Config["radius"])
	}

	input := node1.Inputs["input"]
	if input == nil {
		t.Fatal("node1 input not found")
	}

	if input.ImageID != imageID1 {
		t.Errorf("input ImageID mismatch: got %v, want %v", input.ImageID, imageID1)
	}

	output := node1.Outputs["output"]
	if output == nil {
		t.Fatal("node1 output not found")
	}

	if output.ImageID != imageID2 {
		t.Errorf("output ImageID mismatch: got %v, want %v", output.ImageID, imageID2)
	}

	expectedConn := imagegraph.OutputConnection{NodeID: node2ID, InputName: "input"}
	if _, ok := output.Connections[expectedConn]; !ok {
		t.Errorf("output connection not found: %v", expectedConn)
	}

	node2 := deserialized.Nodes[node2ID]
	if node2 == nil {
		t.Fatal("node2 not found")
	}

	if node2.Type != imagegraph.NodeTypeOutput {
		t.Errorf("node2 type mismatch: got %v, want %v", node2.Type, imagegraph.NodeTypeOutput)
	}

	node2Input := node2.Inputs["input"]
	if !node2Input.Connected {
		t.Error("node2 input should be connected")
	}

	if node2Input.InputConnection.NodeID != node1ID {
		t.Errorf("node2 input connection NodeID mismatch: got %v, want %v", node2Input.InputConnection.NodeID, node1ID)
	}

	if node2Input.InputConnection.OutputName != "output" {
		t.Errorf("node2 input connection OutputName mismatch: got %v, want output", node2Input.InputConnection.OutputName)
	}
}

func TestImageGraphEmptyNodes(t *testing.T) {
	imageGraphID := imagegraph.MustNewImageGraphID()

	original := &imagegraph.ImageGraph{
		ID:      imageGraphID,
		Name:    "Empty Graph",
		Version: 1,
		Nodes:   imagegraph.Nodes{},
	}

	row, err := serializeImageGraph(original)
	if err != nil {
		t.Fatalf("serializeImageGraph failed: %v", err)
	}

	deserialized, err := deserializeImageGraph(row)
	if err != nil {
		t.Fatalf("deserializeImageGraph failed: %v", err)
	}

	if len(deserialized.Nodes) != 0 {
		t.Errorf("Expected empty nodes, got %d nodes", len(deserialized.Nodes))
	}
}

func TestLayoutRoundTrip(t *testing.T) {
	graphID := imagegraph.MustNewImageGraphID()
	node1ID := imagegraph.MustNewNodeID()
	node2ID := imagegraph.MustNewNodeID()

	original := &ui.Layout{
		GraphID: graphID,
		NodePositions: []ui.NodePosition{
			{NodeID: node1ID, X: 100.5, Y: 200.75},
			{NodeID: node2ID, X: 300.25, Y: 400.5},
		},
	}

	row, err := serializeLayout(original)
	if err != nil {
		t.Fatalf("serializeLayout failed: %v", err)
	}

	deserialized, err := deserializeLayout(row)
	if err != nil {
		t.Fatalf("deserializeLayout failed: %v", err)
	}

	if deserialized.GraphID != original.GraphID {
		t.Errorf("GraphID mismatch: got %v, want %v", deserialized.GraphID, original.GraphID)
	}

	if len(deserialized.NodePositions) != len(original.NodePositions) {
		t.Fatalf("NodePositions count mismatch: got %d, want %d", len(deserialized.NodePositions), len(original.NodePositions))
	}

	for i, pos := range deserialized.NodePositions {
		origPos := original.NodePositions[i]
		if pos.NodeID != origPos.NodeID {
			t.Errorf("Position %d NodeID mismatch: got %v, want %v", i, pos.NodeID, origPos.NodeID)
		}
		if pos.X != origPos.X {
			t.Errorf("Position %d X mismatch: got %v, want %v", i, pos.X, origPos.X)
		}
		if pos.Y != origPos.Y {
			t.Errorf("Position %d Y mismatch: got %v, want %v", i, pos.Y, origPos.Y)
		}
	}
}

func TestViewportRoundTrip(t *testing.T) {
	graphID := imagegraph.MustNewImageGraphID()

	original := &ui.Viewport{
		GraphID: graphID,
		Zoom:    1.5,
		PanX:    100.25,
		PanY:    200.75,
	}

	row, err := serializeViewport(original)
	if err != nil {
		t.Fatalf("serializeViewport failed: %v", err)
	}

	deserialized, err := deserializeViewport(row)
	if err != nil {
		t.Fatalf("deserializeViewport failed: %v", err)
	}

	if deserialized.GraphID != original.GraphID {
		t.Errorf("GraphID mismatch: got %v, want %v", deserialized.GraphID, original.GraphID)
	}

	if deserialized.Zoom != original.Zoom {
		t.Errorf("Zoom mismatch: got %v, want %v", deserialized.Zoom, original.Zoom)
	}

	if deserialized.PanX != original.PanX {
		t.Errorf("PanX mismatch: got %v, want %v", deserialized.PanX, original.PanX)
	}

	if deserialized.PanY != original.PanY {
		t.Errorf("PanY mismatch: got %v, want %v", deserialized.PanY, original.PanY)
	}
}

func TestNodeTypeMapping(t *testing.T) {
	tests := []struct {
		nodeType imagegraph.NodeType
		want     string
	}{
		{imagegraph.NodeTypeInput, "input"},
		{imagegraph.NodeTypeOutput, "output"},
		{imagegraph.NodeTypeBlur, "blur"},
		{imagegraph.NodeTypeCrop, "crop"},
		{imagegraph.NodeTypeResize, "resize"},
		{imagegraph.NodeTypeResizeMatch, "resize_match"},
		{imagegraph.NodeTypePixelInflate, "pixel_inflate"},
		{imagegraph.NodeTypePaletteExtract, "palette_extract"},
		{imagegraph.NodeTypePaletteApply, "palette_apply"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := nodeTypeMapper.FromWithDefault(tt.nodeType, "unknown")
			if got != tt.want {
				t.Errorf("nodeTypeMapper.From(%v) = %v, want %v", tt.nodeType, got, tt.want)
			}

			roundtrip, err := nodeTypeMapper.To(got)
			if err != nil {
				t.Fatalf("nodeTypeMapper.To(%v) failed: %v", got, err)
			}
			if roundtrip != tt.nodeType {
				t.Errorf("roundtrip failed: got %v, want %v", roundtrip, tt.nodeType)
			}
		})
	}
}

func TestNodeStateMapping(t *testing.T) {
	tests := []struct {
		state imagegraph.NodeState
		want  string
	}{
		{imagegraph.Waiting, "waiting"},
		{imagegraph.Generating, "generating"},
		{imagegraph.Generated, "generated"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := nodeStateMapper.FromWithDefault(tt.state, "unknown")
			if got != tt.want {
				t.Errorf("nodeStateMapper.From(%v) = %v, want %v", tt.state, got, tt.want)
			}

			roundtrip, err := nodeStateMapper.To(got)
			if err != nil {
				t.Fatalf("nodeStateMapper.To(%v) failed: %v", got, err)
			}
			if roundtrip != tt.state {
				t.Errorf("roundtrip failed: got %v, want %v", roundtrip, tt.state)
			}
		})
	}
}
