package http

import (
	"context"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/dmpettyp/artwork/domain/imagegraph"
)

// handleWebSocket upgrades HTTP connections to WebSocket for real-time updates
func (s *HTTPServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	graphIDStr := r.PathValue("id")
	if graphIDStr == "" {
		http.Error(w, "graph ID required", http.StatusBadRequest)
		return
	}

	graphID, err := imagegraph.ParseImageGraphID(graphIDStr)
	if err != nil {
		http.Error(w, "invalid graph ID", http.StatusBadRequest)
		return
	}

	// Accept the WebSocket connection
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionDisabled, // Disable compression for lower latency
	})
	if err != nil {
		s.logger.Error("failed to accept websocket", "error", err)
		return
	}

	// Register the connection with the hub
	s.wsHub.Register(graphID, conn)

	// Ensure cleanup on exit
	defer func() {
		s.wsHub.Unregister(graphID, conn)
		conn.Close(websocket.StatusNormalClosure, "")
	}()

	// Set up a context with timeout for the connection
	ctx := r.Context()

	// Keep the connection alive with ping/pong
	go s.keepAlive(ctx, conn)

	// Wait for the connection to close
	// We don't expect clients to send messages, so we just wait for disconnect
	s.waitForClose(ctx, conn)
}

// keepAlive sends periodic pings to keep the connection alive
func (s *HTTPServer) keepAlive(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := conn.Ping(ctx); err != nil {
				s.logger.Debug("ping failed, connection likely closed", "error", err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// waitForClose waits for the WebSocket connection to close
func (s *HTTPServer) waitForClose(ctx context.Context, conn *websocket.Conn) {
	for {
		_, _, err := conn.Read(ctx)
		if err != nil {
			// Connection closed or error
			return
		}
		// We don't expect clients to send messages, but if they do, ignore them
	}
}
