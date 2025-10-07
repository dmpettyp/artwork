package http

import (
	"context"
	"encoding/json"
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
	if err := s.messageBus.HandleCommand(context.TODO(), command); err != nil {
		s.logger.Error("failed to handle CreateImageGraphCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to create image graph"})
		return
	}

	// Return successful response
	respondJSON(w, http.StatusCreated, createImageGraphResponse{ID: imageGraphID.String()})
}

// respondJSON writes a JSON response with the given status code
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
