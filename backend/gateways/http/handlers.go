package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/domain/imagegraph"
)

func (s *HTTPServer) handleGetNodeTypeSchemas(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, nodeTypeSchemasResponse{
		NodeTypes: buildNodeTypeSchemas(),
	})
}

func (s *HTTPServer) handleListImageGraphs(w http.ResponseWriter, r *http.Request) {
	imageGraphs, err := s.imageGraphViews.List(r.Context())
	if err != nil {
		s.logger.Error("failed to list image graphs", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to list image graphs"})
		return
	}

	summaries := make([]imageGraphSummary, 0, len(imageGraphs))
	for _, ig := range imageGraphs {
		summaries = append(summaries, imageGraphSummary{
			ID:   ig.ID.String(),
			Name: ig.Name,
		})
	}

	respondJSON(w, http.StatusOK, listImageGraphsResponse{ImageGraphs: summaries})
}

func (s *HTTPServer) handleCreateImageGraph(w http.ResponseWriter, r *http.Request) {
	var req createImageGraphRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to parse request body", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	imageGraphID := imagegraph.MustNewImageGraphID()
	command := application.NewCreateImageGraphCommand(imageGraphID, req.Name)

	if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
		s.logger.Error("failed to handle CreateImageGraphCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to create image graph"})
		return
	}

	respondJSON(w, http.StatusCreated, createImageGraphResponse{ID: imageGraphID.String()})
}

func (s *HTTPServer) handleGetImageGraph(w http.ResponseWriter, r *http.Request) {
	imageGraphID, err := imagegraph.ParseImageGraphID(r.PathValue("id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	ig, err := s.imageGraphViews.Get(r.Context(), imageGraphID)
	if err != nil {
		if errors.Is(err, application.ErrImageGraphNotFound) {
			respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
			return
		}
		s.logger.Error("failed to get image graph", "error", err, "id", imageGraphID)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to retrieve image graph"})
		return
	}

	respondJSON(w, http.StatusOK, mapImageGraphToResponse(ig))
}

func (s *HTTPServer) handleAddNode(w http.ResponseWriter, r *http.Request) {
	imageGraphIDStr := r.PathValue("id")

	imageGraphID, err := imagegraph.ParseImageGraphID(imageGraphIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	var req addNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to parse request body", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
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

	nodeType, err := imagegraph.NodeTypeMapper.To(req.Type)

	if err != nil {
		s.logger.Error("failed to parse request body", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid node type"})
		return
	}

	config := imagegraph.NewNodeConfig(nodeType)
	if err := json.Unmarshal(req.Config, config); err != nil {
		s.logger.Error("failed to parse config", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid config"})
		return
	}

	nodeID := imagegraph.MustNewNodeID()

	command := application.NewAddImageGraphNodeCommand(
		imageGraphID,
		nodeID,
		nodeType,
		req.Name,
		config,
	)

	if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
		if errors.Is(err, application.ErrImageGraphNotFound) {
			respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
			return
		}
		s.logger.Error("failed to handle AddImageGraphNodeCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to add node"})
		return
	}

	respondJSON(w, http.StatusCreated, addNodeResponse{ID: nodeID.String()})
}

func (s *HTTPServer) handleDeleteNode(w http.ResponseWriter, r *http.Request) {
	imageGraphIDStr := r.PathValue("id")

	imageGraphID, err := imagegraph.ParseImageGraphID(imageGraphIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	nodeIDStr := r.PathValue("node_id")

	nodeID, err := imagegraph.ParseNodeID(nodeIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid node ID"})
		return
	}

	command := application.NewRemoveImageGraphNodeCommand(imageGraphID, nodeID)

	if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
		if errors.Is(err, application.ErrImageGraphNotFound) {
			respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
			return
		}
		s.logger.Error("failed to handle RemoveImageGraphNodeCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to delete node"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *HTTPServer) handleConnectNodes(w http.ResponseWriter, r *http.Request) {
	imageGraphIDStr := r.PathValue("id")

	imageGraphID, err := imagegraph.ParseImageGraphID(imageGraphIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	var req connectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to parse request body", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

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

	fromNodeID, err := imagegraph.ParseNodeID(req.FromNodeID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid from_node_id"})
		return
	}

	toNodeID, err := imagegraph.ParseNodeID(req.ToNodeID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid to_node_id"})
		return
	}

	command := application.NewConnectImageGraphNodesCommand(
		imageGraphID,
		fromNodeID,
		imagegraph.OutputName(req.OutputName),
		toNodeID,
		imagegraph.InputName(req.InputName),
	)

	if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
		if errors.Is(err, application.ErrImageGraphNotFound) {
			respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
			return
		}
		s.logger.Error("failed to handle ConnectImageGraphNodesCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to connect nodes"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *HTTPServer) handleDisconnectNodes(w http.ResponseWriter, r *http.Request) {
	imageGraphIDStr := r.PathValue("id")

	imageGraphID, err := imagegraph.ParseImageGraphID(imageGraphIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	var req connectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to parse request body", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

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

	fromNodeID, err := imagegraph.ParseNodeID(req.FromNodeID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid from_node_id"})
		return
	}

	toNodeID, err := imagegraph.ParseNodeID(req.ToNodeID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid to_node_id"})
		return
	}

	command := application.NewDisconnectImageGraphNodesCommand(
		imageGraphID,
		fromNodeID,
		imagegraph.OutputName(req.OutputName),
		toNodeID,
		imagegraph.InputName(req.InputName),
	)

	if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
		if errors.Is(err, application.ErrImageGraphNotFound) {
			respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
			return
		}
		s.logger.Error("failed to handle DisconnectImageGraphNodesCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to disconnect nodes"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *HTTPServer) handleUpdateNode(w http.ResponseWriter, r *http.Request) {
	imageGraphIDStr := r.PathValue("id")

	imageGraphID, err := imagegraph.ParseImageGraphID(imageGraphIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	nodeIDStr := r.PathValue("node_id")

	nodeID, err := imagegraph.ParseNodeID(nodeIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid node ID"})
		return
	}

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
		// Look up the image graph to get the node's type
		ig, err := s.imageGraphViews.Get(r.Context(), imageGraphID)
		if err != nil {
			if errors.Is(err, application.ErrImageGraphNotFound) {
				respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
				return
			}
			s.logger.Error("failed to get image graph", "error", err)
			respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to get image graph"})
			return
		}

		node, exists := ig.Nodes[nodeID]
		if !exists {
			respondJSON(w, http.StatusNotFound, errorResponse{Error: "node not found"})
			return
		}

		config := imagegraph.NewNodeConfig(node.Type)
		if err := json.Unmarshal(req.Config, config); err != nil {
			s.logger.Error("failed to parse config", "error", err)
			respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid config"})
			return
		}

		command := application.NewSetImageGraphNodeConfigCommand(
			imageGraphID,
			nodeID,
			config,
		)

		if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
			if errors.Is(err, application.ErrImageGraphNotFound) {
				respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
				return
			}
			s.logger.Error("failed to handle SetImageGraphNodeConfigCommand", "error", err)
			respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to update node config"})
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *HTTPServer) handleUploadNodeOutputImage(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 10 * 1024 * 1024 // 10 MB

	imageGraphIDStr := r.PathValue("id")

	imageGraphID, err := imagegraph.ParseImageGraphID(imageGraphIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	nodeIDStr := r.PathValue("node_id")

	nodeID, err := imagegraph.ParseNodeID(nodeIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid node ID"})
		return
	}

	outputName := r.PathValue("output_name")
	if outputName == "" {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "output_name is required"})
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		s.logger.Error("failed to parse multipart form", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid multipart form data"})
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		s.logger.Error("failed to get form file", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "image file is required"})
		return
	}
	defer file.Close()

	s.logger.Info("filename", "f", header.Filename)

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

	imageData, err := io.ReadAll(file)
	if err != nil {
		s.logger.Error("failed to read image data", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to read image file"})
		return
	}

	imageID := imagegraph.MustNewImageID()

	if err := s.imageStorage.Save(imageID, imageData); err != nil {
		s.logger.Error("failed to save image to storage", "error", err, "image_id", imageID)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to save image"})
		return
	}

	command := application.NewSetImageGraphNodeOutputImageCommand(
		imageGraphID,
		nodeID,
		imagegraph.OutputName(outputName),
		imageID,
		0,
	)

	if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
		if errors.Is(err, application.ErrImageGraphNotFound) {
			respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
			return
		}
		s.logger.Error("failed to handle SetImageGraphNodeOutputImageCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to set node output image"})
		return
	}

	setNameCommand := application.NewSetImageGraphNodeNameCommand(
		imageGraphID,
		nodeID,
		header.Filename,
	)

	if err := s.messageBus.HandleCommand(r.Context(), setNameCommand); err != nil {
		if errors.Is(err, application.ErrImageGraphNotFound) {
			respondJSON(w, http.StatusNotFound, errorResponse{Error: "image graph not found"})
			return
		}
		s.logger.Error("failed to handle SetImageGraphNodeOutputImageCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to set node output image"})
		return
	}

	respondJSON(w, http.StatusCreated, uploadImageResponse{ImageID: imageID.String()})
}

// respondJSON writes a JSON response with the given status code
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Layout Handlers

func (s *HTTPServer) handleGetLayout(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	imageGraphID, err := imagegraph.ParseImageGraphID(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	layout, err := s.layoutViews.Get(r.Context(), imageGraphID)
	if err != nil {
		// If not found, return empty layout with 200 OK
		if errors.Is(err, application.ErrLayoutNotFound) {
			respondJSON(w, http.StatusOK, layoutResponse{
				GraphID:       imageGraphID.String(),
				NodePositions: []nodePosition{},
			})
			return
		}
		s.logger.Error("failed to get layout", "error", err, "id", imageGraphID)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to retrieve layout"})
		return
	}

	nodePositions := make([]nodePosition, 0, len(layout.NodePositions))
	for _, pos := range layout.NodePositions {
		nodePositions = append(nodePositions, nodePosition{
			NodeID: pos.NodeID.String(),
			X:      pos.X,
			Y:      pos.Y,
		})
	}

	response := layoutResponse{
		GraphID:       layout.GraphID.String(),
		NodePositions: nodePositions,
	}

	respondJSON(w, http.StatusOK, response)
}

func (s *HTTPServer) handleUpdateLayout(w http.ResponseWriter, r *http.Request) {
	imageGraphIDStr := r.PathValue("id")

	imageGraphID, err := imagegraph.ParseImageGraphID(imageGraphIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	var req updateLayoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to parse request body", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	nodePositions, err := req.toDomain()
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	command := application.NewUpdateLayoutCommand(
		imageGraphID,
		nodePositions,
	)

	if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
		s.logger.Error("failed to handle UpdateLayoutCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to update layout"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *HTTPServer) handleGetViewport(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	imageGraphID, err := imagegraph.ParseImageGraphID(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

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

	response := viewportResponse{
		GraphID: viewport.GraphID.String(),
		Zoom:    viewport.Zoom,
		PanX:    viewport.PanX,
		PanY:    viewport.PanY,
	}

	respondJSON(w, http.StatusOK, response)
}

func (s *HTTPServer) handleUpdateViewport(w http.ResponseWriter, r *http.Request) {
	imageGraphIDStr := r.PathValue("id")

	imageGraphID, err := imagegraph.ParseImageGraphID(imageGraphIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image graph ID"})
		return
	}

	var req updateViewportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to parse request body", "error", err)
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	command := application.NewUpdateViewportCommand(
		imageGraphID,
		req.Zoom,
		req.PanX,
		req.PanY,
	)

	if err := s.messageBus.HandleCommand(r.Context(), command); err != nil {
		s.logger.Error("failed to handle UpdateViewportCommand", "error", err)
		respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to update viewport"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Image Retrieval Handlers

func (s *HTTPServer) handleGetImage(w http.ResponseWriter, r *http.Request) {
	imageIDStr := r.PathValue("image_id")

	imageID, err := imagegraph.ParseImageID(imageIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid image ID"})
		return
	}

	imageData, err := s.imageStorage.Get(imageID)
	if err != nil {
		s.logger.Error("failed to get image from storage", "error", err, "image_id", imageID)
		respondJSON(w, http.StatusNotFound, errorResponse{Error: "image not found"})
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}
