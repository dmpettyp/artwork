package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/domain/imagegraph"
)

type createImageGraphRequest struct {
	Name string `json:"name"`
}

type createImageGraphResponse struct {
	ID string `json:"id"`
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

// respondJSON writes a JSON response with the given status code
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
