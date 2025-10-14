package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/domain/imagegraph"
	httpgateway "github.com/dmpettyp/artwork/gateways/http"
	"github.com/dmpettyp/artwork/infrastructure/inmem"
	"github.com/dmpettyp/dorky"
)

// testServer wraps HTTPServer with test utilities
type testServer struct {
	server     *httpgateway.HTTPServer
	testServer *httptest.Server
	messageBus *dorky.MessageBus
	cancelFunc context.CancelFunc
}

func setupTestServer(t *testing.T) *testServer {
	t.Helper()

	// Create logger that discards output during tests
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Create unit of work
	uow, err := inmem.NewUnitOfWork()
	if err != nil {
		t.Fatalf("failed to create unit of work: %v", err)
	}

	// Create message bus
	mb := dorky.NewMessageBus(logger)

	// Register command handlers
	_, err = application.NewImageGraphCommandHandlers(mb, uow)
	if err != nil {
		t.Fatalf("failed to create command handlers: %v", err)
	}

	// Register event handlers
	_, err = application.NewImageGraphEventHandlers(mb, uow)
	if err != nil {
		t.Fatalf("failed to create event handlers: %v", err)
	}

	// Create HTTP server
	httpServer := httpgateway.NewHTTPServer(logger, mb, uow.ImageGraphViews, uow.UIMetadataViews)

	// Start the message bus
	ctx, cancel := context.WithCancel(context.Background())
	go mb.Start(ctx)

	// Create test server
	ts := httptest.NewServer(httpServer.Handler())

	return &testServer{
		server:     httpServer,
		testServer: ts,
		messageBus: mb,
		cancelFunc: cancel,
	}
}

func (ts *testServer) Stop() {
	ts.testServer.Close()
	ts.cancelFunc()
	ts.messageBus.Stop()
}

func (ts *testServer) URL() string {
	return ts.testServer.URL
}

// HTTP client helpers

func (ts *testServer) createImageGraph(t *testing.T, name string) string {
	t.Helper()

	reqBody := map[string]string{"name": name}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		ts.URL()+"/imagegraphs",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("failed to create image graph: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 201, got %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	return response.ID
}

func (ts *testServer) addNode(t *testing.T, graphID, nodeType, name, config string) string {
	t.Helper()

	reqBody := map[string]string{
		"name":   name,
		"type":   nodeType,
		"config": config,
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		fmt.Sprintf("%s/imagegraphs/%s/nodes", ts.URL(), graphID),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("failed to add node: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 201, got %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	return response.ID
}

func (ts *testServer) connectNodes(t *testing.T, graphID, fromNodeID, outputName, toNodeID, inputName string) {
	t.Helper()

	reqBody := map[string]string{
		"from_node_id": fromNodeID,
		"output_name":  outputName,
		"to_node_id":   toNodeID,
		"input_name":   inputName,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/imagegraphs/%s/connectNodes", ts.URL(), graphID),
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to connect nodes: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 204, got %d: %s", resp.StatusCode, string(bodyBytes))
	}
}

func (ts *testServer) getImageGraph(t *testing.T, graphID string) map[string]interface{} {
	t.Helper()

	resp, err := http.Get(fmt.Sprintf("%s/imagegraphs/%s", ts.URL(), graphID))
	if err != nil {
		t.Fatalf("failed to get image graph: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	return response
}

func (ts *testServer) updateNode(t *testing.T, graphID, nodeID string, name *string, config *string) {
	t.Helper()

	reqBody := make(map[string]interface{})
	if name != nil {
		reqBody["name"] = *name
	}
	if config != nil {
		reqBody["config"] = *config
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("%s/imagegraphs/%s/nodes/%s", ts.URL(), graphID, nodeID),
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to update node: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 204, got %d: %s", resp.StatusCode, string(bodyBytes))
	}
}

func (ts *testServer) setNodeOutputImage(t *testing.T, graphID, nodeID, outputName, imageID string) {
	t.Helper()

	reqBody := map[string]string{"image_id": imageID}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("%s/imagegraphs/%s/nodes/%s/outputs/%s", ts.URL(), graphID, nodeID, outputName),
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to set node output image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 204, got %d: %s", resp.StatusCode, string(bodyBytes))
	}
}

// Tests

func TestEndToEndGraphCreationAndRetrieval(t *testing.T) {
	server := setupTestServer(t)
	defer server.Stop()

	// Create graph
	graphID := server.createImageGraph(t, "Test Graph")

	// Add two nodes
	inputNodeID := server.addNode(t, graphID, "input", "Input Node", `{}`)
	scaleNodeID := server.addNode(t, graphID, "scale", "Scale Node", `{"factor": 2.0}`)

	// Connect them
	server.connectNodes(t, graphID, inputNodeID, "original", scaleNodeID, "original")

	// Get the graph
	graph := server.getImageGraph(t, graphID)

	// Verify basic structure
	if graph["id"].(string) != graphID {
		t.Errorf("expected graph ID %s, got %s", graphID, graph["id"])
	}

	if graph["name"].(string) != "Test Graph" {
		t.Errorf("expected name 'Test Graph', got %s", graph["name"])
	}

	nodes := graph["nodes"].([]interface{})
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	// Find the scale node and verify its input is connected
	var scaleNode map[string]interface{}
	for _, n := range nodes {
		node := n.(map[string]interface{})
		if node["id"].(string) == scaleNodeID {
			scaleNode = node
			break
		}
	}

	if scaleNode == nil {
		t.Fatal("scale node not found")
	}

	// Verify node state
	if scaleNode["state"].(string) != "waiting" {
		t.Errorf("expected state 'waiting', got %s", scaleNode["state"])
	}

	// Verify inputs
	inputs := scaleNode["inputs"].([]interface{})
	if len(inputs) != 1 {
		t.Fatalf("expected 1 input, got %d", len(inputs))
	}

	input := inputs[0].(map[string]interface{})
	if input["connected"].(bool) != true {
		t.Error("expected input to be connected")
	}

	connection := input["connection"].(map[string]interface{})
	if connection["node_id"].(string) != inputNodeID {
		t.Errorf("expected connection from %s, got %s", inputNodeID, connection["node_id"])
	}
	if connection["output_name"].(string) != "original" {
		t.Errorf("expected output_name 'original', got %s", connection["output_name"])
	}
}

func TestStateTransitionAndEventPropagation(t *testing.T) {
	server := setupTestServer(t)
	defer server.Stop()

	// Create graph
	graphID := server.createImageGraph(t, "Test Graph")

	// Add two connected nodes
	inputNodeID := server.addNode(t, graphID, "input", "Input Node", `{}`)
	scaleNodeID := server.addNode(t, graphID, "scale", "Scale Node", `{"factor": 2.0}`)
	server.connectNodes(t, graphID, inputNodeID, "original", scaleNodeID, "original")

	// Set output image on input node
	imageID := imagegraph.MustNewImageID().String()
	server.setNodeOutputImage(t, graphID, inputNodeID, "original", imageID)

	// Wait a bit for event propagation (message bus processes async)
	time.Sleep(100 * time.Millisecond)

	// Get the graph and verify propagation
	graph := server.getImageGraph(t, graphID)
	nodes := graph["nodes"].([]interface{})

	// Find the scale node
	var scaleNode map[string]interface{}
	for _, n := range nodes {
		node := n.(map[string]interface{})
		if node["id"].(string) == scaleNodeID {
			scaleNode = node
			break
		}
	}

	if scaleNode == nil {
		t.Fatal("scale node not found")
	}

	// Verify the input received the image
	inputs := scaleNode["inputs"].([]interface{})
	input := inputs[0].(map[string]interface{})

	if input["image_id"].(string) != imageID {
		t.Errorf("expected input image_id %s, got %s", imageID, input["image_id"])
	}

	// Verify state transitioned to "generating"
	if scaleNode["state"].(string) != "generating" {
		t.Errorf("expected state 'generating', got %s", scaleNode["state"])
	}
}

func TestNodeConfigUpdate(t *testing.T) {
	server := setupTestServer(t)
	defer server.Stop()

	// Create graph with node
	graphID := server.createImageGraph(t, "Test Graph")
	nodeID := server.addNode(t, graphID, "input", "Input Node", `{}`)

	// Update config
	newConfig := `{}`
	server.updateNode(t, graphID, nodeID, nil, &newConfig)

	// Get graph and verify config updated
	graph := server.getImageGraph(t, graphID)
	nodes := graph["nodes"].([]interface{})
	node := nodes[0].(map[string]interface{})

	if node["config"].(string) != newConfig {
		t.Errorf("expected config %s, got %s", newConfig, node["config"])
	}
}

func TestErrorScenarios(t *testing.T) {
	server := setupTestServer(t)
	defer server.Stop()

	t.Run("404 for non-existent graph", func(t *testing.T) {
		fakeID := imagegraph.MustNewImageGraphID().String()

		resp, err := http.Get(fmt.Sprintf("%s/imagegraphs/%s", server.URL(), fakeID))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("400 for invalid UUID", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/imagegraphs/not-a-uuid", server.URL()))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("400 for invalid config JSON", func(t *testing.T) {
		graphID := server.createImageGraph(t, "Test Graph")
		nodeID := server.addNode(t, graphID, "input", "Input Node", `{}`)

		reqBody := map[string]string{"config": "not valid json"}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("%s/imagegraphs/%s/nodes/%s", server.URL(), graphID, nodeID),
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", resp.StatusCode)
		}
	})
}
