package main

import (
	"context"
	"log/slog"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
	"github.com/dmpettyp/dorky/messagebus"
)

func ptr[T any](v T) *T {
	return &v
}

// bootstrap creates a default ImageGraph extracted from the running server
func bootstrap(ctx context.Context, logger *slog.Logger, messageBus *messagebus.MessageBus) error {
	logger.Info("bootstrapping application with default ImageGraph")

	// Generate IDs for the graph and nodes
	graphID := imagegraph.MustNewImageGraphID()
	inputNodeID := imagegraph.MustNewNodeID()
	cropNodeID := imagegraph.MustNewNodeID()
	resizeShrinkNodeID := imagegraph.MustNewNodeID()
	blurNodeID := imagegraph.MustNewNodeID()
	resizeGrowNodeID := imagegraph.MustNewNodeID()
	resizeMatchNodeID := imagegraph.MustNewNodeID()
	pixelInflateNodeID := imagegraph.MustNewNodeID()
	output1NodeID := imagegraph.MustNewNodeID()
	output2NodeID := imagegraph.MustNewNodeID()
	output3NodeID := imagegraph.MustNewNodeID()
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
		imagegraph.NewNodeConfigInput(),
	)
	if err := messageBus.HandleCommand(ctx, addInputCmd); err != nil {
		return err
	}
	logger.Info("added Input node", "id", inputNodeID.String())

	// Add Crop node
	cropConfig := imagegraph.NewNodeConfigCrop()
	cropConfig.Left = ptr(564)
	cropConfig.Right = ptr(1565)
	cropConfig.Top = ptr(771)
	cropConfig.Bottom = ptr(1994)
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
	resizeShrinkConfig := imagegraph.NewNodeConfigResize()
	resizeShrinkConfig.Width = ptr(15)
	resizeShrinkConfig.Interpolation = "Bicubic"
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
	blurConfig := imagegraph.NewNodeConfigBlur()
	blurConfig.Radius = 1
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
	resizeGrowConfig := imagegraph.NewNodeConfigResize()
	resizeGrowConfig.Width = ptr(500)
	resizeGrowConfig.Interpolation = "NearestNeighbor"
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

	// Add Resize Match node
	resizeMatchConfig := imagegraph.NewNodeConfigResizeMatch()
	resizeMatchConfig.Interpolation = "NearestNeighbor"
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

	// Add Pixel Inflate node
	pixelInflateConfig := imagegraph.NewNodeConfigPixelInflate()
	pixelInflateConfig.Width = 500
	pixelInflateConfig.LineWidth = 3
	pixelInflateConfig.LineColor = "#333333"
	addPixelInflateCmd := application.NewAddImageGraphNodeCommand(
		graphID,
		pixelInflateNodeID,
		imagegraph.NodeTypePixelInflate,
		"",
		pixelInflateConfig,
	)
	if err := messageBus.HandleCommand(ctx, addPixelInflateCmd); err != nil {
		return err
	}
	logger.Info("added Pixel Inflate node", "id", pixelInflateNodeID.String())

	// Add Output node (Width 500)
	addOutput1Cmd := application.NewAddImageGraphNodeCommand(
		graphID,
		output1NodeID,
		imagegraph.NodeTypeOutput,
		"Width 500",
		imagegraph.NewNodeConfigOutput(),
	)
	if err := messageBus.HandleCommand(ctx, addOutput1Cmd); err != nil {
		return err
	}
	logger.Info("added Output node (Width 500)", "id", output1NodeID.String())

	// Add Output node (Output with original size)
	addOutput2Cmd := application.NewAddImageGraphNodeCommand(
		graphID,
		output2NodeID,
		imagegraph.NodeTypeOutput,
		"Output with original size",
		imagegraph.NewNodeConfigOutput(),
	)
	if err := messageBus.HandleCommand(ctx, addOutput2Cmd); err != nil {
		return err
	}
	logger.Info("added Output node (Output with original size)", "id", output2NodeID.String())

	// Add Output node (no lines)
	addOutput3Cmd := application.NewAddImageGraphNodeCommand(
		graphID,
		output3NodeID,
		imagegraph.NodeTypeOutput,
		"no lines",
		imagegraph.NewNodeConfigOutput(),
	)
	if err := messageBus.HandleCommand(ctx, addOutput3Cmd); err != nil {
		return err
	}
	logger.Info("added Output node (no lines)", "id", output3NodeID.String())

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

	// Connect Resize (shrink) → Pixel Inflate
	connect6Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		resizeShrinkNodeID,
		"resized",
		pixelInflateNodeID,
		"original",
	)
	if err := messageBus.HandleCommand(ctx, connect6Cmd); err != nil {
		return err
	}
	logger.Info("connected Resize (shrink) to Pixel Inflate")

	// Connect Blur → Resize (grow to 500w)
	connect7Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		blurNodeID,
		"blurred",
		resizeGrowNodeID,
		"original",
	)
	if err := messageBus.HandleCommand(ctx, connect7Cmd); err != nil {
		return err
	}
	logger.Info("connected Blur to Resize (grow to 500w)")

	// Connect Resize (grow to 500w) → Output (Width 500)
	connect8Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		resizeGrowNodeID,
		"resized",
		output1NodeID,
		"input",
	)
	if err := messageBus.HandleCommand(ctx, connect8Cmd); err != nil {
		return err
	}
	logger.Info("connected Resize (grow to 500w) to Output (Width 500)")

	// Connect Pixel Inflate → Output (Output with original size)
	connect9Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		pixelInflateNodeID,
		"inflated",
		output2NodeID,
		"input",
	)
	if err := messageBus.HandleCommand(ctx, connect9Cmd); err != nil {
		return err
	}
	logger.Info("connected Pixel Inflate to Output (Output with original size)")

	// Connect Resize Match → Output (no lines)
	connect10Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		resizeMatchNodeID,
		"resized",
		output3NodeID,
		"input",
	)
	if err := messageBus.HandleCommand(ctx, connect10Cmd); err != nil {
		return err
	}
	logger.Info("connected Resize Match to Output (no lines)")

	// Set node layout positions
	layoutCmd := application.NewUpdateLayoutCommand(
		graphID,
		[]ui.NodePosition{
			{NodeID: inputNodeID, X: -530.6755718206077, Y: 697.8155894863006},
			{NodeID: cropNodeID, X: -203.67722892973154, Y: 467.9825097594408},
			{NodeID: resizeShrinkNodeID, X: 88.46872385139525, Y: 140.27954065464667},
			{NodeID: blurNodeID, X: 441.1165295054946, Y: 68.33292188308917},
			{NodeID: resizeGrowNodeID, X: 759.1643755098712, Y: 173.30806002694175},
			{NodeID: output1NodeID, X: 1097.7823165595007, Y: 195.33684713308418},
			{NodeID: resizeMatchNodeID, X: 626.7919132756106, Y: 502.2265824180362},
			{NodeID: output2NodeID, X: 933.5495647535879, Y: 922.3431853255},
			{NodeID: pixelInflateNodeID, X: 545.3986714153481, Y: 922.3273270022839},
			{NodeID: output3NodeID, X: 991.5299221548803, Y: 540.9343650135985},
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
		1,
	)
	if err := messageBus.HandleCommand(ctx, setOutputCmd); err != nil {
		return err
	}
	logger.Info("set Input node output", "imageID", imageID.String())

	logger.Info("bootstrap complete", "graphID", graphID.String())
	return nil
}
