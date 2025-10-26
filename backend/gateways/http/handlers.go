package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
	"github.com/dmpettyp/mapper"
)

type createImageGraphRequest struct {
	Name string `json:"name"`
}

type createImageGraphResponse struct {
	ID string `json:"id"`
}

type addNodeRequest struct {
	Name   string                `json:"name"`
	Type   string                `json:"type"`
	Config imagegraph.NodeConfig `json:"config"`
}

type addNodeResponse struct {
	ID string `json:"id"`
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

var nodeTypeMapper = mapper.MustNew[string, imagegraph.NodeType](
	"input", imagegraph.NodeTypeInput,
	"blur", imagegraph.NodeTypeBlur,
	"output", imagegraph.NodeTypeOutput,
	"resize", imagegraph.NodeTypeResize,
	"resize_match", imagegraph.NodeTypeResizeMatch,
)

var nodeStateMapper = mapper.MustNew[string, imagegraph.NodeState](
	"waiting", imagegraph.Waiting,
	"generating", imagegraph.Generating,
	"generated", imagegraph.Generated,
)

type errorResponse struct {
	Error string `json:"error"`
}

type nodeTypeSchemasResponse struct {
	NodeTypes map[string]imagegraph.NodeTypeSchema `json:"node_types"`
}

func (s *HTTPServer) handleGetNodeTypeSchemas(w http.ResponseWriter, r *http.Request) {
	// Get all node type schemas from domain
	schemas := imagegraph.GetAllNodeTypeSchemas()

	// Return schemas as JSON
	respondJSON(w, http.StatusOK, nodeTypeSchemasResponse{
		NodeTypes: schemas,
	})
}

func (s *HTTPServer) handleListImageGraphs(w http.ResponseWriter, r *http.Request) {
	// Fetch all ImageGraphs from views
	imageGraphs, err := s.imageGraphViews.List(r.Context())
	if err != nil {
		s.logger.Error("failed to list image graphs", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to list image graphs"})
		return
	}

	// Map to response DTOs
	summaries := make([]imageGraphSummary, 0, len(imageGraphs))
	for _, ig := range imageGraphs {
		summaries = append(summaries, imageGraphSummary{
			ID:   ig.ID.String(),
			Name: ig.Name,
		})
	}

	// Return successful response
	respondJSON(w, http.StatusOK, listImageGraphsResponse{ImageGraphs: summaries})
}

func (s *HTTPServer) handleCreateImageGraph(w http.ResponseWriter, r *http.Request) {
	var req createImageGraphRequest

	// Parse JSON request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to parse request body", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	// Validate name
	if req.Name == "" {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "name is required"})
		return
	}

	// Generate new ImageGraphID
	imageGraphID := imagegraph.MustNewImageGraphID()

	// Create command
	command := application.NewCreateImageGraphCommand(imageGraphID, req.Name)

	// Send command to message bus
	if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
		s.logger.Error("failed to handle CreateImageGraphCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to create image graph"})
		return
	}

	// Return successful response
	respondJSON(w, http.StatusCreated, createImageGraphResponse{ID: imageGraphID.String()})
}

func (s *HTTPServer) handleGetImageGraph(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	idStr := r.PathValue("id")

	// Parse ImageGraphID
	imageGraphID, err := imagegraph.ParseImageGraphID(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	// Fetch ImageGraph from view
	ig, err := s.imageGraphViews.Get(r.Context(), imageGraphID)
	if err != nil {
		// Check if it's a not found error
		if errors.Is(err, application.ErrImageGraphNotFound) {
			respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
			return
		}
		s.logger.Error("failed to get image graph", "error", err, "id", imageGraphID)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to retrieve image graph"})
		return
	}

	// Map domain model to response DTO
	response := mapImageGraphToResponse(ig)

	// Return successful response
	respondJSON(w, http.StatusOK, response)
}

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

func (s *HTTPServer) handleAddNode(w http.ResponseWriter, r *http.Request) {
	// Extract ImageGraph ID from path
	imageGraphIDStr := r.PathValue("id")

	// Parse ImageGraphID
	imageGraphID, err := imagegraph.ParseImageGraphID(imageGraphIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	// Parse request body
	var req addNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to parse request body", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	// Validate inputs
	if req.Name == "" {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "name is required"})
		return
	}
	if req.Type == "" {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "type is required"})
		return
	}
	if req.Config == nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "config is required"})
		return
	}

	// Parse NodeType
	nodeType, err := nodeTypeMapper.To(req.Type)

	if err != nil {
		s.logger.Error("failed to parse request body", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid node type"})
		return
	}

	// Generate new NodeID
	nodeID := imagegraph.MustNewNodeID()

	// Create command
	command := application.NewAddImageGraphNodeCommand(
		imageGraphID,
		nodeID,
		nodeType,
		req.Name,
		req.Config,
	)

	// Send command to message bus
	if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
		// Check if it's a not found error
		if errors.Is(err, application.ErrImageGraphNotFound) {
			respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
			return
		}
		s.logger.Error("failed to handle AddImageGraphNodeCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to add node"})
		return
	}

	// Return successful response
	respondJSON(w, http.StatusCreated, addNodeResponse{ID: nodeID.String()})
}

func (s *HTTPServer) handleDeleteNode(w http.ResponseWriter, r *http.Request) {
	// Extract ImageGraph ID from path
	imageGraphIDStr := r.PathValue("id")

	// Parse ImageGraphID
	imageGraphID, err := imagegraph.ParseImageGraphID(imageGraphIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	// Extract Node ID from path
	nodeIDStr := r.PathValue("node_id")

	// Parse NodeID
	nodeID, err := imagegraph.ParseNodeID(nodeIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid node ID"})
		return
	}

	// Create command
	command := application.NewRemoveImageGraphNodeCommand(imageGraphID, nodeID)

	// Send command to message bus
	if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
		// Check if it's a not found error
		if errors.Is(err, application.ErrImageGraphNotFound) {
			respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
			return
		}
		s.logger.Error("failed to handle RemoveImageGraphNodeCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to delete node"})
		return
	}

	// Return successful response with no content
	w.WriteHeader(http.StatusNoContent)
}

func (s *HTTPServer) handleConnectNodes(w http.ResponseWriter, r *http.Request) {
	// Extract ImageGraph ID from path
	imageGraphIDStr := r.PathValue("id")

	// Parse ImageGraphID
	imageGraphID, err := imagegraph.ParseImageGraphID(imageGraphIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	// Parse request body
	var req connectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to parse request body", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	// Validate inputs
	if req.FromNodeID == "" {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "from_node_id is required"})
		return
	}
	if req.OutputName == "" {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "output_name is required"})
		return
	}
	if req.ToNodeID == "" {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "to_node_id is required"})
		return
	}
	if req.InputName == "" {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "input_name is required"})
		return
	}

	// Parse FromNodeID
	fromNodeID, err := imagegraph.ParseNodeID(req.FromNodeID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid from_node_id"})
		return
	}

	// Parse ToNodeID
	toNodeID, err := imagegraph.ParseNodeID(req.ToNodeID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid to_node_id"})
		return
	}

	// Create command
	command := application.NewConnectImageGraphNodesCommand(
		imageGraphID,
		fromNodeID,
		imagegraph.OutputName(req.OutputName),
		toNodeID,
		imagegraph.InputName(req.InputName),
	)

	// Send command to message bus
	if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
		// Check if it's a not found error
		if errors.Is(err, application.ErrImageGraphNotFound) {
			respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
			return
		}
		s.logger.Error("failed to handle ConnectImageGraphNodesCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to connect nodes"})
		return
	}

	// Return successful response with no content
	w.WriteHeader(http.StatusNoContent)
}

func (s *HTTPServer) handleDisconnectNodes(w http.ResponseWriter, r *http.Request) {
	// Extract ImageGraph ID from path
	imageGraphIDStr := r.PathValue("id")

	// Parse ImageGraphID
	imageGraphID, err := imagegraph.ParseImageGraphID(imageGraphIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	// Parse request body
	var req connectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to parse request body", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	// Validate inputs
	if req.FromNodeID == "" {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "from_node_id is required"})
		return
	}
	if req.OutputName == "" {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "output_name is required"})
		return
	}
	if req.ToNodeID == "" {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "to_node_id is required"})
		return
	}
	if req.InputName == "" {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "input_name is required"})
		return
	}

	// Parse FromNodeID
	fromNodeID, err := imagegraph.ParseNodeID(req.FromNodeID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid from_node_id"})
		return
	}

	// Parse ToNodeID
	toNodeID, err := imagegraph.ParseNodeID(req.ToNodeID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid to_node_id"})
		return
	}

	// Create command
	command := application.NewDisconnectImageGraphNodesCommand(
		imageGraphID,
		fromNodeID,
		imagegraph.OutputName(req.OutputName),
		toNodeID,
		imagegraph.InputName(req.InputName),
	)

	// Send command to message bus
	if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
		// Check if it's a not found error
		if errors.Is(err, application.ErrImageGraphNotFound) {
			respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
			return
		}
		s.logger.Error("failed to handle DisconnectImageGraphNodesCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to disconnect nodes"})
		return
	}

	// Return successful response with no content
	w.WriteHeader(http.StatusNoContent)
}

func (s *HTTPServer) handleUpdateNode(w http.ResponseWriter, r *http.Request) {
	// Extract ImageGraph ID from path
	imageGraphIDStr := r.PathValue("id")

	// Parse ImageGraphID
	imageGraphID, err := imagegraph.ParseImageGraphID(imageGraphIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	// Extract Node ID from path
	nodeIDStr := r.PathValue("node_id")

	// Parse NodeID
	nodeID, err := imagegraph.ParseNodeID(nodeIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid node ID"})
		return
	}

	// Parse request body
	var req updateNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to parse request body", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	// Validate that at least one field is provided
	if req.Name == nil && req.Config == nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "at least one of name or config must be provided"})
		return
	}

	// Update name if provided
	if req.Name != nil {
		command := application.NewSetImageGraphNodeNameCommand(
			imageGraphID,
			nodeID,
			*req.Name,
		)

		if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
			// Check if it's a not found error
			if errors.Is(err, application.ErrImageGraphNotFound) {
				respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
				return
			}
			s.logger.Error("failed to handle SetImageGraphNodeNameCommand", "error", err)
			respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to update node name"})
			return
		}
	}

	// Update config if provided
	if req.Config != nil {
		command := application.NewSetImageGraphNodeConfigCommand(
			imageGraphID,
			nodeID,
			req.Config,
		)

		if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
			// Check if it's a not found error
			if errors.Is(err, application.ErrImageGraphNotFound) {
				respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
				return
			}
			s.logger.Error("failed to handle SetImageGraphNodeConfigCommand", "error", err)
			respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to update node config"})
			return
		}
	}

	// Return successful response with no content
	w.WriteHeader(http.StatusNoContent)
}

func (s *HTTPServer) handleUploadNodeOutputImage(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 10 * 1024 * 1024 // 10 MB

	// Extract ImageGraph ID from path
	imageGraphIDStr := r.PathValue("id")

	// Parse ImageGraphID
	imageGraphID, err := imagegraph.ParseImageGraphID(imageGraphIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	// Extract Node ID from path
	nodeIDStr := r.PathValue("node_id")

	// Parse NodeID
	nodeID, err := imagegraph.ParseNodeID(nodeIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid node ID"})
		return
	}

	// Extract output name from path
	outputName := r.PathValue("output_name")
	if outputName == "" {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "output_name is required"})
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		s.logger.Error("failed to parse multipart form", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid multipart form data"})
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("image")
	if err != nil {
		s.logger.Error("failed to get form file", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "image file is required"})
		return
	}
	defer file.Close()

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "file must be an image"})
		return
	}

	// Validate file size
	if header.Size > maxUploadSize {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "image file too large (max 10MB)"})
		return
	}

	// Read file data
	imageData, err := io.ReadAll(file)
	if err != nil {
		s.logger.Error("failed to read image data", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to read image file"})
		return
	}

	// Generate new ImageID
	imageID := imagegraph.MustNewImageID()

	// Save image to storage
	if err := s.imageStorage.Save(imageID, imageData); err != nil {
		s.logger.Error("failed to save image to storage", "error", err, "image_id", imageID)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to save image"})
		return
	}

	// Create command to set the image on the node output
	command := application.NewSetImageGraphNodeOutputImageCommand(
		imageGraphID,
		nodeID,
		imagegraph.OutputName(outputName),
		imageID,
	)

	// Send command to message bus
	if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
		// Check if it's a not found error
		if errors.Is(err, application.ErrImageGraphNotFound) {
			respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
			return
		}
		s.logger.Error("failed to handle SetImageGraphNodeOutputImageCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to set node output image"})
		return
	}

	// Return the generated image ID
	respondJSON(w, http.StatusCreated, uploadImageResponse{ImageID: imageID.String()})
}

// respondJSON writes a JSON response with the given status code
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Layout Handlers

type layoutResponse struct {
	GraphID       string                 `json:"graph_id"`
	NodePositions []nodePositionResponse `json:"node_positions"`
}

type nodePositionResponse struct {
	NodeID string  `json:"node_id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
}

type updateLayoutRequest struct {
	NodePositions []nodePositionResponse `json:"node_positions"`
}

// Viewport Handlers

type viewportResponse struct {
	GraphID string  `json:"graph_id"`
	Zoom    float64 `json:"zoom"`
	PanX    float64 `json:"pan_x"`
	PanY    float64 `json:"pan_y"`
}

type updateViewportRequest struct {
	Zoom float64 `json:"zoom"`
	PanX float64 `json:"pan_x"`
	PanY float64 `json:"pan_y"`
}

func (s *HTTPServer) handleGetLayout(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	idStr := r.PathValue("id")

	// Parse ImageGraphID
	imageGraphID, err := imagegraph.ParseImageGraphID(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	// Fetch layout from view
	layout, err := s.layoutViews.Get(r.Context(), imageGraphID)
	if err != nil {
		// If not found, return empty layout with 200 OK
		if errors.Is(err, application.ErrLayoutNotFound) {
			respondJSON(w, http.StatusOK, layoutResponse{
				GraphID:       imageGraphID.String(),
				NodePositions: []nodePositionResponse{},
			})
			return
		}
		s.logger.Error("failed to get layout", "error", err, "id", imageGraphID)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to retrieve layout"})
		return
	}

	// Map domain model to response DTO
	nodePositions := make([]nodePositionResponse, 0, len(layout.NodePositions))
	for _, pos := range layout.NodePositions {
		nodePositions = append(nodePositions, nodePositionResponse{
			NodeID: pos.NodeID.String(),
			X:      pos.X,
			Y:      pos.Y,
		})
	}

	response := layoutResponse{
		GraphID:       layout.GraphID.String(),
		NodePositions: nodePositions,
	}

	// Return successful response
	respondJSON(w, http.StatusOK, response)
}

func (s *HTTPServer) handleUpdateLayout(w http.ResponseWriter, r *http.Request) {
	// Extract ImageGraph ID from path
	imageGraphIDStr := r.PathValue("id")

	// Parse ImageGraphID
	imageGraphID, err := imagegraph.ParseImageGraphID(imageGraphIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	// Parse request body
	var req updateLayoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to parse request body", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	// Convert node positions to domain types
	nodePositions := make([]ui.NodePosition, 0, len(req.NodePositions))
	for _, pos := range req.NodePositions {
		nodeID, err := imagegraph.ParseNodeID(pos.NodeID)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid node ID: " + pos.NodeID})
			return
		}
		nodePositions = append(nodePositions, ui.NodePosition{
			NodeID: nodeID,
			X:      pos.X,
			Y:      pos.Y,
		})
	}

	// Create command
	command := application.NewUpdateLayoutCommand(
		imageGraphID,
		nodePositions,
	)

	// Send command to message bus
	if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
		s.logger.Error("failed to handle UpdateLayoutCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to update layout"})
		return
	}

	// Return successful response with no content
	w.WriteHeader(http.StatusNoContent)
}

func (s *HTTPServer) handleGetViewport(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	idStr := r.PathValue("id")

	// Parse ImageGraphID
	imageGraphID, err := imagegraph.ParseImageGraphID(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	// Fetch viewport from view
	viewport, err := s.viewportViews.Get(r.Context(), imageGraphID)
	if err != nil {
		// If not found, return default viewport with 200 OK
		if errors.Is(err, application.ErrViewportNotFound) {
			respondJSON(w, http.StatusOK, viewportResponse{
				GraphID: imageGraphID.String(),
				Zoom:    1.0,
				PanX:    0,
				PanY:    0,
			})
			return
		}
		s.logger.Error("failed to get viewport", "error", err, "id", imageGraphID)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to retrieve viewport"})
		return
	}

	// Map domain model to response DTO
	response := viewportResponse{
		GraphID: viewport.GraphID.String(),
		Zoom:    viewport.Zoom,
		PanX:    viewport.PanX,
		PanY:    viewport.PanY,
	}

	// Return successful response
	respondJSON(w, http.StatusOK, response)
}

func (s *HTTPServer) handleUpdateViewport(w http.ResponseWriter, r *http.Request) {
	// Extract ImageGraph ID from path
	imageGraphIDStr := r.PathValue("id")

	// Parse ImageGraphID
	imageGraphID, err := imagegraph.ParseImageGraphID(imageGraphIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	// Parse request body
	var req updateViewportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to parse request body", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	// Create command
	command := application.NewUpdateViewportCommand(
		imageGraphID,
		req.Zoom,
		req.PanX,
		req.PanY,
	)

	// Send command to message bus
	if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
		s.logger.Error("failed to handle UpdateViewportCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to update viewport"})
		return
	}

	// Return successful response with no content
	w.WriteHeader(http.StatusNoContent)
}

// Image Retrieval Handlers

func (s *HTTPServer) handleGetImage(w http.ResponseWriter, r *http.Request) {
	// Extract image ID from path
	imageIDStr := r.PathValue("image_id")

	// Parse ImageID
	imageID, err := imagegraph.ParseImageID(imageIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image ID"})
		return
	}

	// Get image from storage
	imageData, err := s.imageStorage.Get(imageID)
	if err != nil {
		s.logger.Error("failed to get image from storage", "error", err, "image_id", imageID)
		respondJSON(w, http.StatusNotFound, errorResponse{Error: "image not found"})
		return
	}

	// Set content type and write image data
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}
