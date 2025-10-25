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
	resize1NodeID := imagegraph.MustNewNodeID()
	blurNodeID := imagegraph.MustNewNodeID()
	resize2NodeID := imagegraph.MustNewNodeID()
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
		"Input",
		imagegraph.NodeConfig{},
	)
	if err := messageBus.HandleCommand(ctx, addInputCmd); err != nil {
		return err
	}
	logger.Info("added Input node", "id", inputNodeID.String())

	// Add Resize Small node (width: 30)
	resize1Config := imagegraph.NodeConfig{
		"width": float64(30),
	}
	addResize1Cmd := application.NewAddImageGraphNodeCommand(
		graphID,
		resize1NodeID,
		imagegraph.NodeTypeResize,
		"Resize Small",
		resize1Config,
	)
	if err := messageBus.HandleCommand(ctx, addResize1Cmd); err != nil {
		return err
	}
	logger.Info("added Resize Small node", "id", resize1NodeID.String())

	// Add Blur node (radius: 2)
	blurConfig := imagegraph.NodeConfig{
		"radius": float64(2),
	}
	addBlurCmd := application.NewAddImageGraphNodeCommand(
		graphID,
		blurNodeID,
		imagegraph.NodeTypeBlur,
		"Blur",
		blurConfig,
	)
	if err := messageBus.HandleCommand(ctx, addBlurCmd); err != nil {
		return err
	}
	logger.Info("added Blur node", "id", blurNodeID.String())

	// Add Resize Large node (width: 500)
	resize2Config := imagegraph.NodeConfig{
		"width": float64(500),
	}
	addResize2Cmd := application.NewAddImageGraphNodeCommand(
		graphID,
		resize2NodeID,
		imagegraph.NodeTypeResize,
		"Resize Large",
		resize2Config,
	)
	if err := messageBus.HandleCommand(ctx, addResize2Cmd); err != nil {
		return err
	}
	logger.Info("added Resize Large node", "id", resize2NodeID.String())

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
	addResizeMatchCmd := application.NewAddImageGraphNodeCommand(
		graphID,
		resizeMatchNodeID,
		imagegraph.NodeTypeResizeMatch,
		"balh",
		imagegraph.NodeConfig{
			"interpolation": "NearestNeighbor",
		},
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

	// Connect Input → Resize Small
	connect1Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		inputNodeID,
		"original",
		resize1NodeID,
		"original",
	)
	if err := messageBus.HandleCommand(ctx, connect1Cmd); err != nil {
		return err
	}
	logger.Info("connected Input to Resize Small")

	// Connect Resize Small → Blur
	connect2Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		resize1NodeID,
		"resized",
		blurNodeID,
		"original",
	)
	if err := messageBus.HandleCommand(ctx, connect2Cmd); err != nil {
		return err
	}
	logger.Info("connected Resize Small to Blur")

	// Connect Blur → Resize Large
	connect3Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		blurNodeID,
		"blurred",
		resize2NodeID,
		"original",
	)
	if err := messageBus.HandleCommand(ctx, connect3Cmd); err != nil {
		return err
	}
	logger.Info("connected Blur to Resize Large")

	// Connect Resize Large → Output (Width 500)
	connect4Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		resize2NodeID,
		"resized",
		output1NodeID,
		"input",
	)
	if err := messageBus.HandleCommand(ctx, connect4Cmd); err != nil {
		return err
	}
	logger.Info("connected Resize Large to Output (Width 500)")

	// Connect Blur → Resize Match (original input)
	connect5Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		blurNodeID,
		"blurred",
		resizeMatchNodeID,
		"original",
	)
	if err := messageBus.HandleCommand(ctx, connect5Cmd); err != nil {
		return err
	}
	logger.Info("connected Blur to Resize Match (original)")

	// Connect Input → Resize Match (size_match input)
	connect6Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		inputNodeID,
		"original",
		resizeMatchNodeID,
		"size_match",
	)
	if err := messageBus.HandleCommand(ctx, connect6Cmd); err != nil {
		return err
	}
	logger.Info("connected Input to Resize Match (size_match)")

	// Connect Resize Match → Output (Output with original size)
	connect7Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		resizeMatchNodeID,
		"resized",
		output2NodeID,
		"input",
	)
	if err := messageBus.HandleCommand(ctx, connect7Cmd); err != nil {
		return err
	}
	logger.Info("connected Resize Match to Output (Output with original size)")

	// Set node layout positions
	layoutCmd := application.NewUpdateLayoutCommand(
		graphID,
		[]ui.NodePosition{
			{NodeID: inputNodeID, X: -227.22783540000677, Y: 382.9095273884398},
			{NodeID: resize1NodeID, X: 88.46872385139525, Y: 140.5656953322104},
			{NodeID: blurNodeID, X: 429.26196803959334, Y: 261.3929228991944},
			{NodeID: resize2NodeID, X: 743.922796482284, Y: 124.19630538249392},
			{NodeID: output1NodeID, X: 1128.2654746146752, Y: 122.51596955683397},
			{NodeID: resizeMatchNodeID, X: 738.8547707519044, Y: 461.6760488159614},
			{NodeID: output2NodeID, X: 1127.5719534667228, Y: 458.5823177366225},
		},
	)
	if err := messageBus.HandleCommand(ctx, layoutCmd); err != nil {
		return err
	}
	logger.Info("set node layout positions")

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
