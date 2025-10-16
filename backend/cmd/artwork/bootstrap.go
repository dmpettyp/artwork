package main

import (
	"context"
	"log/slog"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky"
)

// bootstrap creates a default ImageGraph with an Input node connected to a Scale node
func bootstrap(ctx context.Context, logger *slog.Logger, messageBus *dorky.MessageBus) error {
	logger.Info("bootstrapping application with default ImageGraph")

	// Generate IDs for the graph and nodes
	graphID := imagegraph.MustNewImageGraphID()
	inputNodeID := imagegraph.MustNewNodeID()
	scaleNodeID := imagegraph.MustNewNodeID()
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

	// Add Scale node with factor configuration
	scaleConfig := imagegraph.NodeConfig{
		"factor": 0.5,
	}
	addScaleCmd := application.NewAddImageGraphNodeCommand(
		graphID,
		scaleNodeID,
		imagegraph.NodeTypeScale,
		"Scale",
		scaleConfig,
	)
	if err := messageBus.HandleCommand(ctx, addScaleCmd); err != nil {
		return err
	}
	logger.Info("added Scale node", "id", scaleNodeID.String())

	// Connect Input node's "original" output to Scale node's "original" input
	connectCmd := application.NewConnectImageGraphNodesCommand(
		graphID,
		inputNodeID,
		"original",
		scaleNodeID,
		"original",
	)
	if err := messageBus.HandleCommand(ctx, connectCmd); err != nil {
		return err
	}
	logger.Info("connected Input to Scale node")

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
