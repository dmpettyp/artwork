package main

import (
	"context"
	"log/slog"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
	"github.com/dmpettyp/dorky"
)

// bootstrap creates a default ImageGraph extracted from the running server
func bootstrap(ctx context.Context, logger *slog.Logger, messageBus *dorky.MessageBus) error {
	logger.Info("bootstrapping application with default ImageGraph")

	// Generate IDs for the graph and nodes
	graphID := imagegraph.MustNewImageGraphID()
	inputNodeID := imagegraph.MustNewNodeID()
	cropNodeID := imagegraph.MustNewNodeID()
	resizeShrinkNodeID := imagegraph.MustNewNodeID()
	blurNodeID := imagegraph.MustNewNodeID()
	resizeGrowNodeID := imagegraph.MustNewNodeID()
	output1NodeID := imagegraph.MustNewNodeID()
	resizeMatchNodeID := imagegraph.MustNewNodeID()
	output2NodeID := imagegraph.MustNewNodeID()
	imageID, _ := imagegraph.ParseImageID("117284ec-f712-42e9-827e-342bd61368db")

	// Create the ImageGraph
	createGraphCmd := application.NewCreateImageGraphCommand(graphID, "Default Pipeline")
	if err := messageBus.HandleCommand(ctx, createGraphCmd); err != nil {
		return err
	}
	logger.Info("created default ImageGraph", "id", graphID.String())

	// Add Input node
	addInputCmd := application.NewAddImageGraphNodeCommand(
		graphID,
		inputNodeID,
		imagegraph.NodeTypeInput,
		"",
		imagegraph.NodeConfig{},
	)
	if err := messageBus.HandleCommand(ctx, addInputCmd); err != nil {
		return err
	}
	logger.Info("added Input node", "id", inputNodeID.String())

	// Add Crop node
	cropConfig := imagegraph.NodeConfig{
		"left":   500,
		"right":  1599,
		"top":    400,
		"bottom": 1900,
	}
	addCropCmd := application.NewAddImageGraphNodeCommand(
		graphID,
		cropNodeID,
		imagegraph.NodeTypeCrop,
		"",
		cropConfig,
	)
	if err := messageBus.HandleCommand(ctx, addCropCmd); err != nil {
		return err
	}
	logger.Info("added Crop node", "id", cropNodeID.String())

	// Add Resize node (shrink)
	resizeShrinkConfig := imagegraph.NodeConfig{
		"width":         15,
		"interpolation": "Bicubic",
	}
	addResizeShrinkCmd := application.NewAddImageGraphNodeCommand(
		graphID,
		resizeShrinkNodeID,
		imagegraph.NodeTypeResize,
		"shrink",
		resizeShrinkConfig,
	)
	if err := messageBus.HandleCommand(ctx, addResizeShrinkCmd); err != nil {
		return err
	}
	logger.Info("added Resize node (shrink)", "id", resizeShrinkNodeID.String())

	// Add Blur node
	blurConfig := imagegraph.NodeConfig{
		"radius": 1,
	}
	addBlurCmd := application.NewAddImageGraphNodeCommand(
		graphID,
		blurNodeID,
		imagegraph.NodeTypeBlur,
		"",
		blurConfig,
	)
	if err := messageBus.HandleCommand(ctx, addBlurCmd); err != nil {
		return err
	}
	logger.Info("added Blur node", "id", blurNodeID.String())

	// Add Resize node (grow to 500w)
	resizeGrowConfig := imagegraph.NodeConfig{
		"width":         500,
		"interpolation": "NearestNeighbor",
	}
	addResizeGrowCmd := application.NewAddImageGraphNodeCommand(
		graphID,
		resizeGrowNodeID,
		imagegraph.NodeTypeResize,
		"grow to 500w",
		resizeGrowConfig,
	)
	if err := messageBus.HandleCommand(ctx, addResizeGrowCmd); err != nil {
		return err
	}
	logger.Info("added Resize node (grow to 500w)", "id", resizeGrowNodeID.String())

	// Add Output node (Width 500)
	addOutput1Cmd := application.NewAddImageGraphNodeCommand(
		graphID,
		output1NodeID,
		imagegraph.NodeTypeOutput,
		"Width 500",
		imagegraph.NodeConfig{},
	)
	if err := messageBus.HandleCommand(ctx, addOutput1Cmd); err != nil {
		return err
	}
	logger.Info("added Output node (Width 500)", "id", output1NodeID.String())

	// Add Resize Match node
	resizeMatchConfig := imagegraph.NodeConfig{
		"interpolation": "NearestNeighbor",
	}
	addResizeMatchCmd := application.NewAddImageGraphNodeCommand(
		graphID,
		resizeMatchNodeID,
		imagegraph.NodeTypeResizeMatch,
		"",
		resizeMatchConfig,
	)
	if err := messageBus.HandleCommand(ctx, addResizeMatchCmd); err != nil {
		return err
	}
	logger.Info("added Resize Match node", "id", resizeMatchNodeID.String())

	// Add Output node (Output with original size)
	addOutput2Cmd := application.NewAddImageGraphNodeCommand(
		graphID,
		output2NodeID,
		imagegraph.NodeTypeOutput,
		"Output with original size",
		imagegraph.NodeConfig{},
	)
	if err := messageBus.HandleCommand(ctx, addOutput2Cmd); err != nil {
		return err
	}
	logger.Info("added Output node (Output with original size)", "id", output2NodeID.String())

	// Connect Input → Crop
	connect1Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		inputNodeID,
		"original",
		cropNodeID,
		"original",
	)
	if err := messageBus.HandleCommand(ctx, connect1Cmd); err != nil {
		return err
	}
	logger.Info("connected Input to Crop")

	// Connect Crop → Resize (shrink)
	connect2Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		cropNodeID,
		"cropped",
		resizeShrinkNodeID,
		"original",
	)
	if err := messageBus.HandleCommand(ctx, connect2Cmd); err != nil {
		return err
	}
	logger.Info("connected Crop to Resize (shrink)")

	// Connect Crop → Resize Match (size_match)
	connect3Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		cropNodeID,
		"cropped",
		resizeMatchNodeID,
		"size_match",
	)
	if err := messageBus.HandleCommand(ctx, connect3Cmd); err != nil {
		return err
	}
	logger.Info("connected Crop to Resize Match (size_match)")

	// Connect Resize (shrink) → Blur
	connect4Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		resizeShrinkNodeID,
		"resized",
		blurNodeID,
		"original",
	)
	if err := messageBus.HandleCommand(ctx, connect4Cmd); err != nil {
		return err
	}
	logger.Info("connected Resize (shrink) to Blur")

	// Connect Resize (shrink) → Resize Match (original)
	connect5Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		resizeShrinkNodeID,
		"resized",
		resizeMatchNodeID,
		"original",
	)
	if err := messageBus.HandleCommand(ctx, connect5Cmd); err != nil {
		return err
	}
	logger.Info("connected Resize (shrink) to Resize Match (original)")

	// Connect Blur → Resize (grow to 500w)
	connect6Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		blurNodeID,
		"blurred",
		resizeGrowNodeID,
		"original",
	)
	if err := messageBus.HandleCommand(ctx, connect6Cmd); err != nil {
		return err
	}
	logger.Info("connected Blur to Resize (grow to 500w)")

	// Connect Resize (grow to 500w) → Output (Width 500)
	connect7Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		resizeGrowNodeID,
		"resized",
		output1NodeID,
		"input",
	)
	if err := messageBus.HandleCommand(ctx, connect7Cmd); err != nil {
		return err
	}
	logger.Info("connected Resize (grow to 500w) to Output (Width 500)")

	// Connect Resize Match → Output (Output with original size)
	connect8Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		resizeMatchNodeID,
		"resized",
		output2NodeID,
		"input",
	)
	if err := messageBus.HandleCommand(ctx, connect8Cmd); err != nil {
		return err
	}
	logger.Info("connected Resize Match to Output (Output with original size)")

	// Set node layout positions
	layoutCmd := application.NewUpdateLayoutCommand(
		graphID,
		[]ui.NodePosition{
			{NodeID: inputNodeID, X: -530.6755718206077, Y: 697.8155894863006},
			{NodeID: cropNodeID, X: -203.67722892973154, Y: 469.38986386272},
			{NodeID: resizeShrinkNodeID, X: 88.46872385139525, Y: 140.27954065464667},
			{NodeID: blurNodeID, X: 441.1165295054946, Y: 68.33292188308917},
			{NodeID: resizeGrowNodeID, X: 759.1643755098712, Y: 173.30806002694175},
			{NodeID: output1NodeID, X: 1097.7823165595007, Y: 195.33684713308418},
			{NodeID: resizeMatchNodeID, X: 626.7919132756106, Y: 502.2265824180362},
			{NodeID: output2NodeID, X: 1106.6541194569395, Y: 481.84135099908485},
		},
	)
	if err := messageBus.HandleCommand(ctx, layoutCmd); err != nil {
		return err
	}
	logger.Info("set node layout positions")

	// Set viewport state
	viewportCmd := application.NewUpdateViewportCommand(
		graphID,
		0.7105532272722948,
		423.4652138758026,
		166.63734119709807,
	)
	if err := messageBus.HandleCommand(ctx, viewportCmd); err != nil {
		return err
	}
	logger.Info("set viewport state")

	// Set the Input node's "original" output to the specified ImageID
	setOutputCmd := application.NewSetImageGraphNodeOutputImageCommand(
		graphID,
		inputNodeID,
		"original",
		imageID,
	)
	if err := messageBus.HandleCommand(ctx, setOutputCmd); err != nil {
		return err
	}
	logger.Info("set Input node output", "imageID", imageID.String())

	logger.Info("bootstrap complete", "graphID", graphID.String())
	return nil
}
