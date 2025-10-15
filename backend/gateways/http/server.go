package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/infrastructure/filestorage"
	"github.com/dmpettyp/dorky"
)

type HTTPServer struct {
	logger           *slog.Logger
	messageBus       *dorky.MessageBus
	imageGraphViews  application.ImageGraphViews
	uiMetadataViews  application.UIMetadataViews
	imageStorage     filestorage.ImageStorage
	server           *http.Server
	port             string
}

// ServerOption is a functional option for configuring the HTTPServer
type ServerOption func(*HTTPServer)

// WithPort sets a custom port for the HTTP server
func WithPort(port string) ServerOption {
	return func(s *HTTPServer) {
		s.port = port
	}
}

// NewHTTPServer creates a new HTTP server that handles requests by sending
// commands to the provided message bus
func NewHTTPServer(
	logger *slog.Logger,
	mb *dorky.MessageBus,
	imageGraphViews application.ImageGraphViews,
	uiMetadataViews application.UIMetadataViews,
	imageStorage filestorage.ImageStorage,
	opts ...ServerOption,
) *HTTPServer {
	s := &HTTPServer{
		logger:           logger,
		messageBus:       mb,
		imageGraphViews:  imageGraphViews,
		uiMetadataViews:  uiMetadataViews,
		imageStorage:     imageStorage,
		port:             "8080", // default port
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// Set up routes
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/imagegraphs", s.handleListImageGraphs)
	mux.HandleFunc("POST /api/imagegraphs", s.handleCreateImageGraph)
	mux.HandleFunc("GET /api/imagegraphs/{id}", s.handleGetImageGraph)
	mux.HandleFunc("POST /api/imagegraphs/{id}/nodes", s.handleAddNode)
	mux.HandleFunc("DELETE /api/imagegraphs/{id}/nodes/{node_id}", s.handleDeleteNode)
	mux.HandleFunc("PUT /api/imagegraphs/{id}/connectNodes", s.handleConnectNodes)
	mux.HandleFunc("PUT /api/imagegraphs/{id}/disconnectNodes", s.handleDisconnectNodes)
	mux.HandleFunc("PATCH /api/imagegraphs/{id}/nodes/{node_id}", s.handleUpdateNode)
	mux.HandleFunc("PUT /api/imagegraphs/{id}/nodes/{node_id}/outputs/{output_name}", s.handleUploadNodeOutputImage)

	// Image retrieval
	mux.HandleFunc("GET /api/images/{image_id}", s.handleGetImage)

	// UI Metadata routes
	mux.HandleFunc("GET /api/imagegraphs/{id}/ui-metadata", s.handleGetUIMetadata)
	mux.HandleFunc("PUT /api/imagegraphs/{id}/ui-metadata", s.handleUpdateUIMetadata)

	// Serve static frontend files
	fs := http.FileServer(http.Dir("../frontend"))
	mux.Handle("/", fs)

	s.server = &http.Server{
		Addr:    ":" + s.port,
		Handler: mux,
	}

	return s
}

// Start starts the HTTP server in a background goroutine
func (s *HTTPServer) Start() {
	go func() {
		s.logger.Info("starting HTTP server", "port", s.port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", "error", err)
		}
	}()
}

// Stop gracefully shuts down the HTTP server
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("stopping HTTP server")
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}
	return nil
}

// Handler returns the HTTP handler for testing
func (s *HTTPServer) Handler() http.Handler {
	return s.server.Handler
}
