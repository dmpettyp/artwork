package http

import (
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
	"github.com/dmpettyp/mapper"
)

// Request types

type createImageGraphRequest struct {
	Name string `json:"name"`
}

type addNodeRequest struct {
	Name   string                `json:"name"`
	Type   string                `json:"type"`
	Config imagegraph.NodeConfig `json:"config"`
}

type connectionRequest struct {
	FromNodeID string `json:"from_node_id"`
	OutputName string `json:"output_name"`
	ToNodeID   string `json:"to_node_id"`
	InputName  string `json:"input_name"`
}

type updateNodeRequest struct {
	Name   *string               `json:"name,omitempty"`
	Config imagegraph.NodeConfig `json:"config,omitempty"`
}

type updateLayoutRequest struct {
	NodePositions []nodePosition `json:"node_positions"`
}

// toDomain converts the request to domain types
func (r *updateLayoutRequest) toDomain() ([]ui.NodePosition, error) {
	nodePositions := make([]ui.NodePosition, 0, len(r.NodePositions))
	for _, pos := range r.NodePositions {
		nodeID, err := imagegraph.ParseNodeID(pos.NodeID)
		if err != nil {
			return nil, fmt.Errorf("invalid node ID: %s", pos.NodeID)
		}
		nodePositions = append(nodePositions, ui.NodePosition{
			NodeID: nodeID,
			X:      pos.X,
			Y:      pos.Y,
		})
	}
	return nodePositions, nil
}

type updateViewportRequest struct {
	Zoom float64 `json:"zoom"`
	PanX float64 `json:"pan_x"`
	PanY float64 `json:"pan_y"`
}

// Response types

type createImageGraphResponse struct {
	ID string `json:"id"`
}

type addNodeResponse struct {
	ID string `json:"id"`
}

type uploadImageResponse struct {
	ImageID string `json:"image_id"`
}

type listImageGraphsResponse struct {
	ImageGraphs []imageGraphSummary `json:"imagegraphs"`
}

type imageGraphSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type imageGraphResponse struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Version int            `json:"version"`
	Nodes   []nodeResponse `json:"nodes"`
}

type nodeResponse struct {
	ID      string                `json:"id"`
	Name    string                `json:"name"`
	Type    string                `json:"type"`
	Version int                   `json:"version"`
	Config  imagegraph.NodeConfig `json:"config"`
	State   string                `json:"state"`
	Preview string                `json:"preview,omitempty"`
	Inputs  []inputResponse       `json:"inputs"`
	Outputs []outputResponse      `json:"outputs"`
}

type inputResponse struct {
	Name       string                   `json:"name"`
	ImageID    string                   `json:"image_id,omitempty"`
	Connected  bool                     `json:"connected"`
	Connection *inputConnectionResponse `json:"connection,omitempty"`
}

type inputConnectionResponse struct {
	NodeID     string `json:"node_id"`
	OutputName string `json:"output_name"`
}

type outputResponse struct {
	Name        string                     `json:"name"`
	ImageID     string                     `json:"image_id,omitempty"`
	Connections []outputConnectionResponse `json:"connections"`
}

type outputConnectionResponse struct {
	NodeID    string `json:"node_id"`
	InputName string `json:"input_name"`
}

type layoutResponse struct {
	GraphID       string         `json:"graph_id"`
	NodePositions []nodePosition `json:"node_positions"`
}

type nodePosition struct {
	NodeID string  `json:"node_id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
}

type viewportResponse struct {
	GraphID string  `json:"graph_id"`
	Zoom    float64 `json:"zoom"`
	PanX    float64 `json:"pan_x"`
	PanY    float64 `json:"pan_y"`
}

type nodeTypeSchemasResponse struct {
	NodeTypes []nodeTypeSchemaAPIEntry `json:"node_types"`
}

type nodeTypeSchemaAPIEntry struct {
	Name        string         `json:"name"`
	DisplayName string         `json:"display_name"`
	Schema      nodeTypeSchema `json:"schema"`
}

type nodeTypeSchema struct {
	Inputs       []string              `json:"inputs"`
	Outputs      []string              `json:"outputs"`
	NameRequired bool                  `json:"name_required"`
	Fields       []nodeTypeSchemaField `json:"fields"`
}

type nodeTypeSchemaField struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Required bool        `json:"required"`
	Options  []string    `json:"options,omitempty"`
	Default  interface{} `json:"default,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// Mappers

var nodeTypeMapper = mapper.MustNew[string, imagegraph.NodeType](
	"input", imagegraph.NodeTypeInput,
	"output", imagegraph.NodeTypeOutput,
	"crop", imagegraph.NodeTypeCrop,
	"blur", imagegraph.NodeTypeBlur,
	"resize", imagegraph.NodeTypeResize,
	"resize_match", imagegraph.NodeTypeResizeMatch,
	"pixel_inflate", imagegraph.NodeTypePixelInflate,
)

var nodeStateMapper = mapper.MustNew[string, imagegraph.NodeState](
	"waiting", imagegraph.Waiting,
	"generating", imagegraph.Generating,
	"generated", imagegraph.Generated,
)

var fieldTypeMapper = mapper.MustNew[imagegraph.NodeConfigFieldType, string](
	imagegraph.NodeConfigTypeString, "string",
	imagegraph.NodeConfigTypeInt, "int",
	imagegraph.NodeConfigTypeFloat, "float",
	imagegraph.NodeConfigTypeBool, "bool",
	imagegraph.NodeConfigTypeOption, "option",
)

// nodeTypeInfo holds both the API name and display name for a node type
type nodeTypeInfo struct {
	name        string
	displayName string
}

// nodeTypeMetadata maps NodeType constants to their API metadata
var nodeTypeMetadata = map[imagegraph.NodeType]nodeTypeInfo{
	imagegraph.NodeTypeInput:        {"input", "Input"},
	imagegraph.NodeTypeOutput:       {"output", "Output"},
	imagegraph.NodeTypeCrop:         {"crop", "Crop"},
	imagegraph.NodeTypeBlur:         {"blur", "Blur"},
	imagegraph.NodeTypeResize:       {"resize", "Resize"},
	imagegraph.NodeTypeResizeMatch:  {"resize_match", "Resize Match"},
	imagegraph.NodeTypePixelInflate: {"pixel_inflate", "Pixel Inflate"},
}

// Conversion functions

// mapImageGraphToResponse converts a domain ImageGraph to an API response
func mapImageGraphToResponse(ig *imagegraph.ImageGraph) imageGraphResponse {
	nodes := make([]nodeResponse, 0, len(ig.Nodes))

	for _, node := range ig.Nodes {
		// Map inputs in the order defined by the node type configuration
		inputNames := node.Type.InputNames()
		inputs := make([]inputResponse, 0, len(inputNames))
		for _, inputName := range inputNames {
			input, ok := node.Inputs[inputName]
			if !ok {
				continue
			}

			inputResp := inputResponse{
				Name:      string(input.Name),
				Connected: input.Connected,
			}

			if !input.ImageID.IsNil() {
				inputResp.ImageID = input.ImageID.String()
			}

			if input.Connected {
				inputResp.Connection = &inputConnectionResponse{
					NodeID:     input.InputConnection.NodeID.String(),
					OutputName: string(input.InputConnection.OutputName),
				}
			}

			inputs = append(inputs, inputResp)
		}

		// Map outputs in the order defined by the node type configuration
		outputNames := node.Type.OutputNames()
		outputs := make([]outputResponse, 0, len(outputNames))
		for _, outputName := range outputNames {
			output, ok := node.Outputs[outputName]
			if !ok {
				continue
			}

			outputResp := outputResponse{
				Name:        string(output.Name),
				Connections: make([]outputConnectionResponse, 0, len(output.Connections)),
			}

			if !output.ImageID.IsNil() {
				outputResp.ImageID = output.ImageID.String()
			}

			for conn := range output.Connections {
				outputResp.Connections = append(outputResp.Connections, outputConnectionResponse{
					NodeID:    conn.NodeID.String(),
					InputName: string(conn.InputName),
				})
			}

			outputs = append(outputs, outputResp)
		}

		nodeResp := nodeResponse{
			ID:      node.ID.String(),
			Name:    node.Name,
			Type:    nodeTypeMapper.FromWithDefault(node.Type, "unknown"),
			Version: int(node.Version),
			Config:  node.Config,
			State:   nodeStateMapper.FromWithDefault(node.State.Get(), "unknown"),
			Inputs:  inputs,
			Outputs: outputs,
		}

		if !node.Preview.IsNil() {
			nodeResp.Preview = node.Preview.String()
		}

		nodes = append(nodes, nodeResp)
	}

	return imageGraphResponse{
		ID:      ig.ID.String(),
		Name:    ig.Name,
		Version: int(ig.Version),
		Nodes:   nodes,
	}
}

// buildNodeTypeSchemas converts domain node type configs to API schema entries
func buildNodeTypeSchemas() []nodeTypeSchemaAPIEntry {
	configs := imagegraph.NodeTypeConfigs
	apiSchemas := make([]nodeTypeSchemaAPIEntry, 0, len(configs))

	for _, cfg := range configs {
		info, ok := nodeTypeMetadata[cfg.NodeType]
		if !ok {
			// Skip unknown node types
			continue
		}

		// Convert inputs
		inputs := make([]string, len(cfg.Inputs))
		for i, input := range cfg.Inputs {
			inputs[i] = string(input)
		}

		// Convert outputs
		outputs := make([]string, len(cfg.Outputs))
		for i, output := range cfg.Outputs {
			outputs[i] = string(output)
		}

		// Convert fields (preserve order from domain)
		fields := make([]nodeTypeSchemaField, len(cfg.Fields))
		for i, field := range cfg.Fields {
			fields[i] = nodeTypeSchemaField{
				Name:     field.Name,
				Type:     fieldTypeMapper.ToWithDefault(field.FieldType, "unknown"),
				Required: field.Required,
				Options:  field.Options,
				Default:  field.Default,
			}
		}

		apiSchemas = append(apiSchemas, nodeTypeSchemaAPIEntry{
			Name:        info.name,
			DisplayName: info.displayName,
			Schema: nodeTypeSchema{
				Inputs:       inputs,
				Outputs:      outputs,
				NameRequired: cfg.NameRequired,
				Fields:       fields,
			},
		})
	}

	return apiSchemas
}
