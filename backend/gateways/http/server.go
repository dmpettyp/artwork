package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/dorky"
)

type HTTPServer struct {
	logger          *slog.Logger
	messageBus      *dorky.MessageBus
	imageGraphViews application.ImageGraphViews
	server          *http.Server
	port            string
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
	opts ...ServerOption,
) *HTTPServer {
	s := &HTTPServer{
		logger:          logger,
		messageBus:      mb,
		imageGraphViews: imageGraphViews,
		port:            "8080", // default port
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /imagegraphs", s.handleCreateImageGraph)
	mux.HandleFunc("GET /imagegraphs/{id}", s.handleGetImageGraph)
	mux.HandleFunc("POST /imagegraphs/{id}/nodes", s.handleAddNode)
	mux.HandleFunc("DELETE /imagegraphs/{id}/nodes/{node_id}", s.handleDeleteNode)
	mux.HandleFunc("PUT /imagegraphs/{id}/connectNodes", s.handleConnectNodes)
	mux.HandleFunc("PUT /imagegraphs/{id}/disconnectNodes", s.handleDisconnectNodes)

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
