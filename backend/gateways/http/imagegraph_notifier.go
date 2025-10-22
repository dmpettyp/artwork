package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/coder/websocket"
	"github.com/dmpettyp/artwork/domain/imagegraph"
)

// ImageGraphNotifier manages WebSocket connections for image graphs
// and broadcasts notifications about graph changes to connected clients
type ImageGraphNotifier struct {
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

// NewImageGraphNotifier creates a new ImageGraphNotifier
func NewImageGraphNotifier(logger *slog.Logger) *ImageGraphNotifier {
	notifier := &ImageGraphNotifier{
		logger:           logger,
		graphConnections: make(map[imagegraph.ImageGraphID]map[*websocket.Conn]bool),
		broadcast:        make(chan *BroadcastMessage, 256),
		done:             make(chan struct{}),
	}

	// Start the broadcast loop
	go notifier.run()

	return notifier
}

// run is the main loop that handles broadcasting messages
func (n *ImageGraphNotifier) run() {
	for {
		select {
		case msg := <-n.broadcast:
			n.broadcastToGraph(msg.GraphID, msg.Data)
		case <-n.done:
			return
		}
	}
}

// Register adds a connection for a specific graph
func (n *ImageGraphNotifier) Register(graphID imagegraph.ImageGraphID, conn *websocket.Conn) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.graphConnections[graphID] == nil {
		n.graphConnections[graphID] = make(map[*websocket.Conn]bool)
	}
	n.graphConnections[graphID][conn] = true

	n.logger.Info("client connected", "graph_id", graphID.String(), "total_connections", len(n.graphConnections[graphID]))
}

// Unregister removes a connection
func (n *ImageGraphNotifier) Unregister(graphID imagegraph.ImageGraphID, conn *websocket.Conn) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if connections, ok := n.graphConnections[graphID]; ok {
		delete(connections, conn)
		if len(connections) == 0 {
			delete(n.graphConnections, graphID)
		}
	}

	n.logger.Info("client disconnected", "graph_id", graphID.String())
}

// Broadcast sends a message to all clients connected to a specific graph
func (n *ImageGraphNotifier) Broadcast(graphID imagegraph.ImageGraphID, data interface{}) {
	select {
	case n.broadcast <- &BroadcastMessage{GraphID: graphID, Data: data}:
	default:
		n.logger.Warn("broadcast channel full, dropping message", "graph_id", graphID.String())
	}
}

// broadcastToGraph sends data to all connections for a graph
func (n *ImageGraphNotifier) broadcastToGraph(graphID imagegraph.ImageGraphID, data interface{}) {
	n.mu.RLock()
	connections := n.graphConnections[graphID]
	n.mu.RUnlock()

	if len(connections) == 0 {
		return
	}

	// Marshal the message once
	messageBytes, err := json.Marshal(data)
	if err != nil {
		n.logger.Error("failed to marshal websocket message", "error", err)
		return
	}

	// Send to all connections
	for conn := range connections {
		go func(c *websocket.Conn) {
			ctx := context.Background()
			if err := c.Write(ctx, websocket.MessageText, messageBytes); err != nil {
				n.logger.Error("failed to write to websocket", "error", err)
				// Connection is broken, unregister it
				n.Unregister(graphID, c)
			}
		}(conn)
	}
}

// BroadcastNodeUpdate sends a node update to all clients viewing the graph
func (n *ImageGraphNotifier) BroadcastNodeUpdate(graphID imagegraph.ImageGraphID, nodeUpdate interface{}) {
	msg := WebSocketMessage{
		Type: "node_update",
		Data: nodeUpdate,
	}
	n.Broadcast(graphID, msg)
}

// BroadcastLayoutUpdate sends a layout update notification to all clients viewing the graph
func (n *ImageGraphNotifier) BroadcastLayoutUpdate(graphID imagegraph.ImageGraphID) {
	msg := WebSocketMessage{
		Type: "layout_update",
		Data: map[string]interface{}{},
	}
	n.Broadcast(graphID, msg)
}

// Close shuts down the notifier
func (n *ImageGraphNotifier) Close() {
	close(n.done)

	// Close all connections
	n.mu.Lock()
	defer n.mu.Unlock()

	for graphID, connections := range n.graphConnections {
		for conn := range connections {
			conn.Close(websocket.StatusNormalClosure, "server shutting down")
			delete(connections, conn)
		}
		delete(n.graphConnections, graphID)
	}
}
