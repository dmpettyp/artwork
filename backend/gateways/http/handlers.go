package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/mapper"
)

type createImageGraphRequest struct {
	Name string `json:"name"`
}

type createImageGraphResponse struct {
	ID string `json:"id"`
}

type addNodeRequest struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Config string `json:"config"`
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

type imageGraphResponse struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Version int            `json:"version"`
	Nodes   []nodeResponse `json:"nodes"`
}

type nodeResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version int    `json:"version"`
	Config  string `json:"config"`
}

var nodeTypeMapper = mapper.MustNew[string, imagegraph.NodeType](
	"input", imagegraph.NodeTypeInput,
	"scale", imagegraph.NodeTypeScale,
)

type errorResponse struct {
	Error string `json:"error"`
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
		nodes = append(nodes, nodeResponse{
			ID:      node.ID.String(),
			Name:    node.Name,
			Type:    string(node.Type),
			Version: int(node.Version),
			Config:  node.Config,
		})
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
	if req.Config == "" {
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

// respondJSON writes a JSON response with the given status code
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
