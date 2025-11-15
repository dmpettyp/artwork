package postgres

import (
	"encoding/json"
	"fmt"

	"github.com/dmpettyp/artwork/backend/domain/imagegraph"
	"github.com/dmpettyp/artwork/backend/domain/ui"
)

// imageGraphRow represents a row from the image_graphs table
type imageGraphRow struct {
	ID        string
	Name      string
	Version   int64
	Data      []byte // JSONB data
	CreatedAt string
	UpdatedAt string
}

// layoutRow represents a row from the layouts table
type layoutRow struct {
	GraphID   string
	Data      []byte // JSONB data
	UpdatedAt string
}

// viewportRow represents a row from the viewports table
type viewportRow struct {
	GraphID   string
	Data      []byte // JSONB data
	UpdatedAt string
}

// imageGraphData represents the JSON structure for ImageGraph data field
type imageGraphData struct {
	Nodes imagegraph.Nodes `json:"nodes"`
}

// layoutData represents the JSON structure for Layout data field
type layoutData struct {
	NodePositions []ui.NodePosition `json:"node_positions"`
}

// viewportData represents the JSON structure for Viewport data field
type viewportData struct {
	Zoom float64 `json:"zoom"`
	PanX float64 `json:"pan_x"`
	PanY float64 `json:"pan_y"`
}

// serializeImageGraph converts an ImageGraph to database row format
func serializeImageGraph(ig *imagegraph.ImageGraph) (imageGraphRow, error) {
	data := imageGraphData{
		Nodes: ig.Nodes,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return imageGraphRow{}, fmt.Errorf("failed to marshal image graph data: %w", err)
	}

	return imageGraphRow{
		ID:      ig.ID.ID,
		Name:    ig.Name,
		Version: int64(ig.Version),
		Data:    dataJSON,
	}, nil
}

// deserializeImageGraph converts a database row to an ImageGraph
func deserializeImageGraph(row imageGraphRow) (*imagegraph.ImageGraph, error) {
	id, err := imagegraph.ParseImageGraphID(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image graph ID: %w", err)
	}

	var data imageGraphData
	if err := json.Unmarshal(row.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal image graph data: %w", err)
	}

	ig := &imagegraph.ImageGraph{
		ID:      id,
		Name:    row.Name,
		Version: imagegraph.ImageGraphVersion(row.Version),
		Nodes:   data.Nodes,
	}

	return ig, nil
}

// serializeLayout converts a Layout to database row format
func serializeLayout(layout *ui.Layout) (layoutRow, error) {
	data := layoutData{
		NodePositions: layout.NodePositions,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return layoutRow{}, fmt.Errorf("failed to marshal layout data: %w", err)
	}

	return layoutRow{
		GraphID: layout.GraphID.ID,
		Data:    dataJSON,
	}, nil
}

// deserializeLayout converts a database row to a Layout
func deserializeLayout(row layoutRow) (*ui.Layout, error) {
	graphID, err := imagegraph.ParseImageGraphID(row.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse graph ID: %w", err)
	}

	var data layoutData
	if err := json.Unmarshal(row.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal layout data: %w", err)
	}

	layout := &ui.Layout{
		GraphID:       graphID,
		NodePositions: data.NodePositions,
	}

	return layout, nil
}

// serializeViewport converts a Viewport to database row format
func serializeViewport(viewport *ui.Viewport) (viewportRow, error) {
	data := viewportData{
		Zoom: viewport.Zoom,
		PanX: viewport.PanX,
		PanY: viewport.PanY,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return viewportRow{}, fmt.Errorf("failed to marshal viewport data: %w", err)
	}

	return viewportRow{
		GraphID: viewport.GraphID.ID,
		Data:    dataJSON,
	}, nil
}

// deserializeViewport converts a database row to a Viewport
func deserializeViewport(row viewportRow) (*ui.Viewport, error) {
	graphID, err := imagegraph.ParseImageGraphID(row.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse graph ID: %w", err)
	}

	var data viewportData
	if err := json.Unmarshal(row.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal viewport data: %w", err)
	}

	viewport := &ui.Viewport{
		GraphID: graphID,
		Zoom:    data.Zoom,
		PanX:    data.PanX,
		PanY:    data.PanY,
	}

	return viewport, nil
}
