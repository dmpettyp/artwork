package http

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/dmpettyp/dorky/messagebus"
	"github.com/google/uuid"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/infrastructure/filestorage"
)

type HTTPServer struct {
	logger          *slog.Logger
	messageBus      *messagebus.MessageBus
	imageGraphViews application.ImageGraphViews
	layoutViews     application.LayoutViews
	viewportViews   application.ViewportViews
	imageStorage    filestorage.ImageStorage
	notifier        *ImageGraphNotifier
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
	messageBus *messagebus.MessageBus,
	imageGraphViews application.ImageGraphViews,
	layoutViews application.LayoutViews,
	viewportViews application.ViewportViews,
	imageStorage filestorage.ImageStorage,
	notifier *ImageGraphNotifier,
	opts ...ServerOption,
) *HTTPServer {
	s := &HTTPServer{
		logger:          logger,
		messageBus:      messageBus,
		imageGraphViews: imageGraphViews,
		layoutViews:     layoutViews,
		viewportViews:   viewportViews,
		imageStorage:    imageStorage,
		notifier:        notifier,
		port:            "8080", // default port
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// Set up routes
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/node-types", s.handleGetNodeTypeSchemas)
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

	// Layout routes
	mux.HandleFunc("GET /api/imagegraphs/{id}/layout", s.handleGetLayout)
	mux.HandleFunc("PUT /api/imagegraphs/{id}/layout", s.handleUpdateLayout)

	// Viewport routes
	mux.HandleFunc("GET /api/imagegraphs/{id}/viewport", s.handleGetViewport)
	mux.HandleFunc("PUT /api/imagegraphs/{id}/viewport", s.handleUpdateViewport)

	// WebSocket route
	mux.HandleFunc("GET /api/imagegraphs/{id}/ws", s.handleWebSocket)

	// Serve static frontend files
	fs := http.FileServer(http.Dir("../frontend"))
	mux.Handle("/", fs)

	s.server = &http.Server{
		Addr:    ":" + s.port,
		Handler: loggingMiddleware(logger, mux),
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

	// Close notifier first
	s.notifier.Close()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}
	return nil
}

// Handler returns the HTTP handler for testing
func (s *HTTPServer) Handler() http.Handler {
	return s.server.Handler
}

type ctxKey string

const requestIDKey ctxKey = "request_id"

// loggingMiddleware wraps handlers with basic structured request logging and
// request ID propagation.
func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		r = r.WithContext(ctx)

		logger.Info("http_request_start",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"request_id", reqID,
		)

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(lrw, r)

		logger.Info("http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", lrw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"request_id", reqID,
			"bytes", lrw.bytesWritten,
		)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.status = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := lrw.ResponseWriter.Write(b)
	lrw.bytesWritten += n
	return n, err
}

// Hijack delegates to the underlying ResponseWriter if it supports
// http.Hijacker (needed for websockets).
func (lrw *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := lrw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("hijacker not supported")
	}
	return h.Hijack()
}

// Flush delegates to the underlying ResponseWriter if it supports http.Flusher.
func (lrw *loggingResponseWriter) Flush() {
	if f, ok := lrw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Push delegates to the underlying ResponseWriter if it supports http.Pusher.
func (lrw *loggingResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := lrw.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}
