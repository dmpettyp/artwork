package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/coder/websocket"
	"github.com/dmpettyp/artwork/domain/imagegraph"
)

// WebSocketHub manages WebSocket connections for image graphs
type WebSocketHub struct {
	logger *slog.Logger

	// Map of graph ID to set of connections
	graphConnections map[imagegraph.ImageGraphID]map[*websocket.Conn]bool
	mu               sync.RWMutex

	// Channel for broadcasting messages
	broadcast chan *BroadcastMessage
	done      chan struct{}
}

// BroadcastMessage represents a message to broadcast to clients
type BroadcastMessage struct {
	GraphID imagegraph.ImageGraphID
	Data    interface{}
}

// WebSocketMessage is the structure sent to clients
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// NodeUpdateMessage contains node state changes
type NodeUpdateMessage struct {
	NodeID  string      `json:"node_id"`
	State   string      `json:"state"`
	Outputs interface{} `json:"outputs,omitempty"`
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub(logger *slog.Logger) *WebSocketHub {
	hub := &WebSocketHub{
		logger:           logger,
		graphConnections: make(map[imagegraph.ImageGraphID]map[*websocket.Conn]bool),
		broadcast:        make(chan *BroadcastMessage, 256),
		done:             make(chan struct{}),
	}

	// Start the broadcast loop
	go hub.run()

	return hub
}

// run is the main loop that handles broadcasting messages
func (h *WebSocketHub) run() {
	for {
		select {
		case msg := <-h.broadcast:
			h.broadcastToGraph(msg.GraphID, msg.Data)
		case <-h.done:
			return
		}
	}
}

// Register adds a connection for a specific graph
func (h *WebSocketHub) Register(graphID imagegraph.ImageGraphID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.graphConnections[graphID] == nil {
		h.graphConnections[graphID] = make(map[*websocket.Conn]bool)
	}
	h.graphConnections[graphID][conn] = true

	h.logger.Info("client connected", "graph_id", graphID.String(), "total_connections", len(h.graphConnections[graphID]))
}

// Unregister removes a connection
func (h *WebSocketHub) Unregister(graphID imagegraph.ImageGraphID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if connections, ok := h.graphConnections[graphID]; ok {
		delete(connections, conn)
		if len(connections) == 0 {
			delete(h.graphConnections, graphID)
		}
	}

	h.logger.Info("client disconnected", "graph_id", graphID.String())
}

// Broadcast sends a message to all clients connected to a specific graph
func (h *WebSocketHub) Broadcast(graphID imagegraph.ImageGraphID, data interface{}) {
	select {
	case h.broadcast <- &BroadcastMessage{GraphID: graphID, Data: data}:
	default:
		h.logger.Warn("broadcast channel full, dropping message", "graph_id", graphID.String())
	}
}

// broadcastToGraph sends data to all connections for a graph
func (h *WebSocketHub) broadcastToGraph(graphID imagegraph.ImageGraphID, data interface{}) {
	h.mu.RLock()
	connections := h.graphConnections[graphID]
	h.mu.RUnlock()

	if len(connections) == 0 {
		return
	}

	// Marshal the message once
	messageBytes, err := json.Marshal(data)
	if err != nil {
		h.logger.Error("failed to marshal websocket message", "error", err)
		return
	}

	// Send to all connections
	for conn := range connections {
		go func(c *websocket.Conn) {
			ctx := context.Background()
			if err := c.Write(ctx, websocket.MessageText, messageBytes); err != nil {
				h.logger.Error("failed to write to websocket", "error", err)
				// Connection is broken, unregister it
				h.Unregister(graphID, c)
			}
		}(conn)
	}
}

// BroadcastNodeUpdate sends a node update to all clients viewing the graph
func (h *WebSocketHub) BroadcastNodeUpdate(graphID imagegraph.ImageGraphID, nodeUpdate NodeUpdateMessage) {
	msg := WebSocketMessage{
		Type: "node_update",
		Data: nodeUpdate,
	}
	h.Broadcast(graphID, msg)
}

// Close shuts down the hub
func (h *WebSocketHub) Close() {
	close(h.done)

	// Close all connections
	h.mu.Lock()
	defer h.mu.Unlock()

	for graphID, connections := range h.graphConnections {
		for conn := range connections {
			conn.Close(websocket.StatusNormalClosure, "server shutting down")
			delete(connections, conn)
		}
		delete(h.graphConnections, graphID)
	}
}
