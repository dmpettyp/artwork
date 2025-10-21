package main

import (
	"context"
	"log/slog"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky"
)

// bootstrap creates a default ImageGraph with Input → Resize → Blur → Resize → Output pipeline
func bootstrap(ctx context.Context, logger *slog.Logger, messageBus *dorky.MessageBus) error {
	logger.Info("bootstrapping application with default ImageGraph")

	// Generate IDs for the graph and nodes
	graphID := imagegraph.MustNewImageGraphID()
	inputNodeID := imagegraph.MustNewNodeID()
	resize1NodeID := imagegraph.MustNewNodeID()
	blurNodeID := imagegraph.MustNewNodeID()
	resize2NodeID := imagegraph.MustNewNodeID()
	outputNodeID := imagegraph.MustNewNodeID()
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

	// Add Resize node (width: 30)
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
	logger.Info("added Resize node (width 30)", "id", resize1NodeID.String())

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
	logger.Info("added Blur node (radius 2)", "id", blurNodeID.String())

	// Add Resize node (width: 500)
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
	logger.Info("added Resize node (width 500)", "id", resize2NodeID.String())

	// Add Output node
	addOutputCmd := application.NewAddImageGraphNodeCommand(
		graphID,
		outputNodeID,
		imagegraph.NodeTypeOutput,
		"Final",
		imagegraph.NodeConfig{},
	)
	if err := messageBus.HandleCommand(ctx, addOutputCmd); err != nil {
		return err
	}
	logger.Info("added Output node", "id", outputNodeID.String())

	// Connect Input → Resize1
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
	logger.Info("connected Input to Resize1")

	// Connect Resize1 → Blur
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
	logger.Info("connected Resize1 to Blur")

	// Connect Blur → Resize2
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
	logger.Info("connected Blur to Resize2")

	// Connect Resize2 → Output
	connect4Cmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		resize2NodeID,
		"resized",
		outputNodeID,
		"input",
	)
	if err := messageBus.HandleCommand(ctx, connect4Cmd); err != nil {
		return err
	}
	logger.Info("connected Resize2 to Output")

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
