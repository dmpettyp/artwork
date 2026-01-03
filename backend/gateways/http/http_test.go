package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"testing"
	"time"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/domain/imagegraph"
	httpgateway "github.com/dmpettyp/artwork/gateways/http"
	"github.com/dmpettyp/artwork/infrastructure/imagegen"
	"github.com/dmpettyp/artwork/infrastructure/inmem"
	"github.com/dmpettyp/artwork/metrics"
	"github.com/dmpettyp/dorky/messagebus"
)

// mockImageStorage is a simple in-memory image storage for testing
type mockImageStorage struct {
	data map[string][]byte
}

func (m *mockImageStorage) Save(imageID imagegraph.ImageID, imageData []byte) error {
	m.data[imageID.String()] = imageData
	return nil
}

func (m *mockImageStorage) Get(imageID imagegraph.ImageID) ([]byte, error) {
	data, ok := m.data[imageID.String()]
	if !ok {
		return nil, fmt.Errorf("image not found: %s", imageID.String())
	}
	return data, nil
}

func (m *mockImageStorage) Exists(imageID imagegraph.ImageID) (bool, error) {
	_, ok := m.data[imageID.String()]
	return ok, nil
}

func (m *mockImageStorage) Remove(imageID imagegraph.ImageID) error {
	delete(m.data, imageID.String())
	return nil
}

// testServer wraps HTTPServer with test utilities
type testServer struct {
	server     *httpgateway.HTTPServer
	httpServer *http.Server
	listener   net.Listener
	baseURL    string
	messageBus *messagebus.MessageBus
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
	mb := messagebus.New(logger)

	// Create mock image storage
	imageStorage := &mockImageStorage{data: make(map[string][]byte)}

	// Create node updater for ImageGen
	nodeUpdater := application.NewNodeUpdater(mb)

	// Create ImageGen with dependencies
	imageGen := imagegen.NewImageGen(imageStorage, nodeUpdater, logger, nil)

	// Create notifier
	notifier := httpgateway.NewImageGraphNotifier(logger)

	// Register command handlers
	_, err = application.NewImageGraphCommandHandlers(mb, uow)
	if err != nil {
		t.Fatalf("failed to create command handlers: %v", err)
	}

	// Register event handlers
	_, err = application.NewImageGraphEventHandlers(mb, uow, imageGen, imageStorage, notifier)
	if err != nil {
		t.Fatalf("failed to create event handlers: %v", err)
	}

	// Create HTTP server
	appMetrics := metrics.NewAppMetrics()
	httpServer := httpgateway.NewHTTPServer(
		logger,
		mb,
		uow.ImageGraphViews,
		uow.LayoutViews,
		uow.ViewportViews,
		imageStorage,
		notifier,
		appMetrics,
	)

	// Start the message bus
	ctx, cancel := context.WithCancel(context.Background())
	go mb.Start(ctx)

	// Create test server bound to IPv4 (tcp6 may be disallowed in some environments)
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("skipping HTTP tests: cannot listen on tcp4: %v", err)
		return nil
	}
	srv := &http.Server{Handler: httpServer.Handler()}
	go func() {
		_ = srv.Serve(ln)
	}()

	return &testServer{
		server:     httpServer,
		httpServer: srv,
		listener:   ln,
		baseURL:    "http://" + ln.Addr().String(),
		messageBus: mb,
		cancelFunc: cancel,
	}
}

func (ts *testServer) Stop() {
	if ts.httpServer != nil {
		_ = ts.httpServer.Shutdown(context.Background())
	}
	if ts.listener != nil {
		_ = ts.listener.Close()
	}
	ts.cancelFunc()
	ts.messageBus.Stop()
}

func (ts *testServer) URL() string {
	return ts.baseURL
}

// HTTP client helpers

func (ts *testServer) createImageGraph(t *testing.T, name string) string {
	t.Helper()

	reqBody := map[string]string{"name": name}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		ts.URL()+"/api/imagegraphs",
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

	reqBody := map[string]interface{}{
		"name": name,
		"type": nodeType,
	}

	if config != "" {
		var configObj map[string]interface{}
		if err := json.Unmarshal([]byte(config), &configObj); err != nil {
			t.Fatalf("failed to unmarshal config: %v", err)
		}
		reqBody["config"] = configObj
	}

	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		fmt.Sprintf("%s/api/imagegraphs/%s/nodes", ts.URL(), graphID),
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
		fmt.Sprintf("%s/api/imagegraphs/%s/connectNodes", ts.URL(), graphID),
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

	resp, err := http.Get(fmt.Sprintf("%s/api/imagegraphs/%s", ts.URL(), graphID))
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
		var configObj map[string]interface{}
		if err := json.Unmarshal([]byte(*config), &configObj); err != nil {
			t.Fatalf("failed to unmarshal config: %v", err)
		}
		reqBody["config"] = configObj
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("%s/api/imagegraphs/%s/nodes/%s", ts.URL(), graphID, nodeID),
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

func (ts *testServer) setNodeOutputImage(t *testing.T, graphID, nodeID, outputName, imageID string) string {
	t.Helper()

	graph := ts.getImageGraph(t, graphID)
	nodes, ok := graph["nodes"].([]interface{})
	if !ok {
		t.Fatalf("unexpected nodes shape in graph response")
	}
	nodeVersion := 0
	for _, n := range nodes {
		if nodeMap, ok := n.(map[string]interface{}); ok {
			if nodeMap["id"] == nodeID {
				if v, ok := nodeMap["version"].(float64); ok {
					nodeVersion = int(v)
				}
				break
			}
		}
	}
	if nodeVersion == 0 {
		t.Fatalf("could not determine node version for node %s", nodeID)
	}

	// Create a simple 1x1 PNG image for testing
	imageData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1 dimensions
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
		0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41, // IDAT chunk
		0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, // IEND chunk
		0x42, 0x60, 0x82,
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Create form file with proper Content-Type header
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="image"; filename="test.png"`)
	h.Set("Content-Type", "image/png")

	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	if _, err := part.Write(imageData); err != nil {
		t.Fatalf("failed to write image data: %v", err)
	}

	if err := writer.WriteField("node_version", fmt.Sprintf("%d", nodeVersion)); err != nil {
		t.Fatalf("failed to add node_version field: %v", err)
	}

	writer.Close()

	req, _ := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/api/imagegraphs/%s/nodes/%s/outputs/%s", ts.URL(), graphID, nodeID, outputName),
		&body,
	)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to set node output image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 201, got %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		ImageID string `json:"image_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	return response.ImageID
}

// Tests

func TestEndToEndGraphCreationAndRetrieval(t *testing.T) {
	server := setupTestServer(t)
	defer server.Stop()

	// Create graph
	graphID := server.createImageGraph(t, "Test Graph")

	// Add two nodes
	inputNodeID := server.addNode(t, graphID, "input", "Input Node", `{}`)
	resizeNodeID := server.addNode(t, graphID, "resize", "Resize Node", `{"width": 800, "interpolation": "Bilinear"}`)

	// Connect them
	server.connectNodes(t, graphID, inputNodeID, "original", resizeNodeID, "original")

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

	// Find the resize node and verify its input is connected
	var resizeNode map[string]interface{}
	for _, n := range nodes {
		node := n.(map[string]interface{})
		if node["id"].(string) == resizeNodeID {
			resizeNode = node
			break
		}
	}

	if resizeNode == nil {
		t.Fatal("resize node not found")
	}

	// Verify node state
	if resizeNode["state"].(string) != "waiting" {
		t.Errorf("expected state 'waiting', got %s", resizeNode["state"])
	}

	// Verify inputs
	inputs := resizeNode["inputs"].([]interface{})
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
	resizeNodeID := server.addNode(t, graphID, "resize", "Resize Node", `{"width": 800, "interpolation": "Bilinear"}`)
	server.connectNodes(t, graphID, inputNodeID, "original", resizeNodeID, "original")

	// Set output image on input node
	imageID := server.setNodeOutputImage(t, graphID, inputNodeID, "original", "")

	// Wait a bit for event propagation (message bus processes async)
	time.Sleep(100 * time.Millisecond)

	// Get the graph and verify propagation
	graph := server.getImageGraph(t, graphID)
	nodes := graph["nodes"].([]interface{})

	// Find the resize node
	var resizeNode map[string]interface{}
	for _, n := range nodes {
		node := n.(map[string]interface{})
		if node["id"].(string) == resizeNodeID {
			resizeNode = node
			break
		}
	}

	if resizeNode == nil {
		t.Fatal("resize node not found")
	}

	// Verify the input received the image
	inputs := resizeNode["inputs"].([]interface{})
	input := inputs[0].(map[string]interface{})

	if input["image_id"].(string) != imageID {
		t.Errorf("expected input image_id %s, got %s", imageID, input["image_id"])
	}

	// Verify state is either "generating" or "generated" (depends on timing)
	state := resizeNode["state"].(string)
	if state != "generating" && state != "generated" {
		t.Errorf("expected state 'generating' or 'generated', got %s", state)
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

	config, ok := node["config"].(map[string]interface{})
	if !ok {
		t.Fatalf("config is not a map: %T", node["config"])
	}
	if len(config) != 0 {
		t.Errorf("expected empty config, got %v", config)
	}
}

func TestErrorScenarios(t *testing.T) {
	server := setupTestServer(t)
	defer server.Stop()

	t.Run("404 for non-existent graph", func(t *testing.T) {
		fakeID := imagegraph.MustNewImageGraphID().String()

		resp, err := http.Get(fmt.Sprintf("%s/api/imagegraphs/%s", server.URL(), fakeID))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("400 for invalid UUID", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/api/imagegraphs/not-a-uuid", server.URL()))
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
			fmt.Sprintf("%s/api/imagegraphs/%s/nodes/%s", server.URL(), graphID, nodeID),
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}
	})
}
