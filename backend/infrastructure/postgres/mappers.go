package postgres

import (
	"encoding/json"
	"fmt"

	"github.com/dmpettyp/artwork/backend/domain/imagegraph"
	"github.com/dmpettyp/artwork/backend/domain/ui"
	"github.com/dmpettyp/mapper"
)

var nodeTypeMapper = mapper.MustNew[string, imagegraph.NodeType](
	"input", imagegraph.NodeTypeInput,
	"output", imagegraph.NodeTypeOutput,
	"crop", imagegraph.NodeTypeCrop,
	"blur", imagegraph.NodeTypeBlur,
	"resize", imagegraph.NodeTypeResize,
	"resize_match", imagegraph.NodeTypeResizeMatch,
	"pixel_inflate", imagegraph.NodeTypePixelInflate,
	"palette_extract", imagegraph.NodeTypePaletteExtract,
	"palette_apply", imagegraph.NodeTypePaletteApply,
)

var nodeStateMapper = mapper.MustNew[string, imagegraph.NodeState](
	"waiting", imagegraph.Waiting,
	"generating", imagegraph.Generating,
	"generated", imagegraph.Generated,
)

type imageGraphRow struct {
	ID        string
	Name      string
	Version   int64
	Data      []byte
	CreatedAt string
	UpdatedAt string
}

type layoutRow struct {
	GraphID   string
	Data      []byte
	UpdatedAt string
}

type viewportRow struct {
	GraphID   string
	Data      []byte
	UpdatedAt string
}

type imageGraphDTO struct {
	Nodes map[string]nodeDTO `json:"nodes"`
}

type nodeDTO struct {
	ID             string                 `json:"id"`
	Version        int64                  `json:"version"`
	Type           string                 `json:"type"`
	Name           string                 `json:"name"`
	State          string                 `json:"state"`
	Config         map[string]interface{} `json:"config"`
	PreviewImageID string                 `json:"preview_image_id,omitempty"`
	Inputs         map[string]inputDTO    `json:"inputs"`
	Outputs        map[string]outputDTO   `json:"outputs"`
}

type inputDTO struct {
	Name       string                `json:"name"`
	ImageID    string                `json:"image_id,omitempty"`
	Connected  bool                  `json:"connected"`
	Connection *inputConnectionDTO   `json:"connection,omitempty"`
}

type inputConnectionDTO struct {
	NodeID     string `json:"node_id"`
	OutputName string `json:"output_name"`
}

type outputDTO struct {
	Name        string                `json:"name"`
	ImageID     string                `json:"image_id,omitempty"`
	Connections []outputConnectionDTO `json:"connections"`
}

type outputConnectionDTO struct {
	NodeID    string `json:"node_id"`
	InputName string `json:"input_name"`
}

type layoutDTO struct {
	NodePositions []nodePositionDTO `json:"node_positions"`
}

type nodePositionDTO struct {
	NodeID string  `json:"node_id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
}

type viewportDTO struct {
	Zoom float64 `json:"zoom"`
	PanX float64 `json:"pan_x"`
	PanY float64 `json:"pan_y"`
}

func serializeImageGraph(ig *imagegraph.ImageGraph) (imageGraphRow, error) {
	nodesDTO := make(map[string]nodeDTO, len(ig.Nodes))

	for nodeID, node := range ig.Nodes {
		inputsDTO := make(map[string]inputDTO, len(node.Inputs))
		for inputName, input := range node.Inputs {
			inputDTO := inputDTO{
				Name:      string(input.Name),
				Connected: input.Connected,
			}

			if !input.ImageID.IsNil() {
				inputDTO.ImageID = input.ImageID.String()
			}

			if input.Connected {
				inputDTO.Connection = &inputConnectionDTO{
					NodeID:     input.InputConnection.NodeID.String(),
					OutputName: string(input.InputConnection.OutputName),
				}
			}

			inputsDTO[string(inputName)] = inputDTO
		}

		outputsDTO := make(map[string]outputDTO, len(node.Outputs))
		for outputName, output := range node.Outputs {
			outputDTO := outputDTO{
				Name:        string(output.Name),
				Connections: make([]outputConnectionDTO, 0, len(output.Connections)),
			}

			if !output.ImageID.IsNil() {
				outputDTO.ImageID = output.ImageID.String()
			}

			for conn := range output.Connections {
				outputDTO.Connections = append(outputDTO.Connections, outputConnectionDTO{
					NodeID:    conn.NodeID.String(),
					InputName: string(conn.InputName),
				})
			}

			outputsDTO[string(outputName)] = outputDTO
		}

		nodeDTO := nodeDTO{
			ID:      node.ID.String(),
			Version: int64(node.Version),
			Type:    nodeTypeMapper.FromWithDefault(node.Type, "unknown"),
			Name:    node.Name,
			State:   nodeStateMapper.FromWithDefault(node.State.Get(), "unknown"),
			Config:  node.Config,
			Inputs:  inputsDTO,
			Outputs: outputsDTO,
		}

		if !node.Preview.IsNil() {
			nodeDTO.PreviewImageID = node.Preview.String()
		}

		nodesDTO[nodeID.String()] = nodeDTO
	}

	dto := imageGraphDTO{
		Nodes: nodesDTO,
	}

	dataJSON, err := json.Marshal(dto)
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

func deserializeImageGraph(row imageGraphRow) (*imagegraph.ImageGraph, error) {
	id, err := imagegraph.ParseImageGraphID(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image graph ID: %w", err)
	}

	var dto imageGraphDTO
	if err := json.Unmarshal(row.Data, &dto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal image graph data: %w", err)
	}

	nodes := make(imagegraph.Nodes, len(dto.Nodes))

	for nodeIDStr, nodeDTO := range dto.Nodes {
		nodeID, err := imagegraph.ParseNodeID(nodeIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node ID %s: %w", nodeIDStr, err)
		}

		nodeType, err := nodeTypeMapper.To(nodeDTO.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node type %s: %w", nodeDTO.Type, err)
		}

		nodeState, err := nodeStateMapper.To(nodeDTO.State)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node state %s: %w", nodeDTO.State, err)
		}

		inputs := make(imagegraph.Inputs, len(nodeDTO.Inputs))
		for inputNameStr, inputDTO := range nodeDTO.Inputs {
			inputName := imagegraph.InputName(inputNameStr)

			input := &imagegraph.Input{
				Name:      inputName,
				Connected: inputDTO.Connected,
			}

			if inputDTO.ImageID != "" {
				imageID, err := imagegraph.ParseImageID(inputDTO.ImageID)
				if err != nil {
					return nil, fmt.Errorf("failed to parse input image ID %s: %w", inputDTO.ImageID, err)
				}
				input.ImageID = imageID
			}

			if inputDTO.Connection != nil {
				connNodeID, err := imagegraph.ParseNodeID(inputDTO.Connection.NodeID)
				if err != nil {
					return nil, fmt.Errorf("failed to parse connection node ID %s: %w", inputDTO.Connection.NodeID, err)
				}
				input.InputConnection = imagegraph.InputConnection{
					NodeID:     connNodeID,
					OutputName: imagegraph.OutputName(inputDTO.Connection.OutputName),
				}
			}

			inputs[inputName] = input
		}

		outputs := make(imagegraph.Outputs, len(nodeDTO.Outputs))
		for outputNameStr, outputDTO := range nodeDTO.Outputs {
			outputName := imagegraph.OutputName(outputNameStr)

			output := &imagegraph.Output{
				Name:        outputName,
				Connections: make(map[imagegraph.OutputConnection]struct{}),
			}

			if outputDTO.ImageID != "" {
				imageID, err := imagegraph.ParseImageID(outputDTO.ImageID)
				if err != nil {
					return nil, fmt.Errorf("failed to parse output image ID %s: %w", outputDTO.ImageID, err)
				}
				output.ImageID = imageID
			}

			for _, connDTO := range outputDTO.Connections {
				connNodeID, err := imagegraph.ParseNodeID(connDTO.NodeID)
				if err != nil {
					return nil, fmt.Errorf("failed to parse output connection node ID %s: %w", connDTO.NodeID, err)
				}
				conn := imagegraph.OutputConnection{
					NodeID:    connNodeID,
					InputName: imagegraph.InputName(connDTO.InputName),
				}
				output.Connections[conn] = struct{}{}
			}

			outputs[outputName] = output
		}

		node := &imagegraph.Node{
			ID:      nodeID,
			Version: imagegraph.NodeVersion(nodeDTO.Version),
			Type:    nodeType,
			Name:    nodeDTO.Name,
			State:   imagegraph.NewNodeState(nodeState),
			Config:  nodeDTO.Config,
			Inputs:  inputs,
			Outputs: outputs,
		}

		if nodeDTO.PreviewImageID != "" {
			previewID, err := imagegraph.ParseImageID(nodeDTO.PreviewImageID)
			if err != nil {
				return nil, fmt.Errorf("failed to parse preview image ID %s: %w", nodeDTO.PreviewImageID, err)
			}
			node.Preview = previewID
		}

		nodes[nodeID] = node
	}

	ig := &imagegraph.ImageGraph{
		ID:      id,
		Name:    row.Name,
		Version: imagegraph.ImageGraphVersion(row.Version),
		Nodes:   nodes,
	}

	return ig, nil
}

func serializeLayout(layout *ui.Layout) (layoutRow, error) {
	positions := make([]nodePositionDTO, len(layout.NodePositions))
	for i, pos := range layout.NodePositions {
		positions[i] = nodePositionDTO{
			NodeID: pos.NodeID.String(),
			X:      pos.X,
			Y:      pos.Y,
		}
	}

	dto := layoutDTO{
		NodePositions: positions,
	}

	dataJSON, err := json.Marshal(dto)
	if err != nil {
		return layoutRow{}, fmt.Errorf("failed to marshal layout data: %w", err)
	}

	return layoutRow{
		GraphID: layout.GraphID.ID,
		Data:    dataJSON,
	}, nil
}

func deserializeLayout(row layoutRow) (*ui.Layout, error) {
	graphID, err := imagegraph.ParseImageGraphID(row.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse graph ID: %w", err)
	}

	var dto layoutDTO
	if err := json.Unmarshal(row.Data, &dto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal layout data: %w", err)
	}

	positions := make([]ui.NodePosition, len(dto.NodePositions))
	for i, posDTO := range dto.NodePositions {
		nodeID, err := imagegraph.ParseNodeID(posDTO.NodeID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node ID %s: %w", posDTO.NodeID, err)
		}
		positions[i] = ui.NodePosition{
			NodeID: nodeID,
			X:      posDTO.X,
			Y:      posDTO.Y,
		}
	}

	layout := &ui.Layout{
		GraphID:       graphID,
		NodePositions: positions,
	}

	return layout, nil
}

func serializeViewport(viewport *ui.Viewport) (viewportRow, error) {
	dto := viewportDTO{
		Zoom: viewport.Zoom,
		PanX: viewport.PanX,
		PanY: viewport.PanY,
	}

	dataJSON, err := json.Marshal(dto)
	if err != nil {
		return viewportRow{}, fmt.Errorf("failed to marshal viewport data: %w", err)
	}

	return viewportRow{
		GraphID: viewport.GraphID.ID,
		Data:    dataJSON,
	}, nil
}

func deserializeViewport(row viewportRow) (*ui.Viewport, error) {
	graphID, err := imagegraph.ParseImageGraphID(row.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse graph ID: %w", err)
	}

	var dto viewportDTO
	if err := json.Unmarshal(row.Data, &dto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal viewport data: %w", err)
	}

	viewport := &ui.Viewport{
		GraphID: graphID,
		Zoom:    dto.Zoom,
		PanX:    dto.PanX,
		PanY:    dto.PanY,
	}

	return viewport, nil
}
