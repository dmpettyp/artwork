package application

import (
	"context"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/infrastructure/imagegen"
)

// nodeOutputGenerator is a function that generates outputs for a specific node type
type nodeOutputGenerator func(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error

// nodeOutputGenerators maps node types to their output generation functions
var nodeOutputGenerators = map[imagegraph.NodeType]nodeOutputGenerator{
	imagegraph.NodeTypeBlur:           generateBlurNodeOutputs,
	imagegraph.NodeTypeCrop:           generateCropNodeOutputs,
	imagegraph.NodeTypeResize:         generateResizeNodeOutputs,
	imagegraph.NodeTypeResizeMatch:    generateResizeMatchNodeOutputs,
	imagegraph.NodeTypePixelInflate:   generatePixelInflateNodeOutputs,
	imagegraph.NodeTypePaletteExtract: generatePaletteExtractNodeOutputs,
	imagegraph.NodeTypePaletteApply:   generatePaletteApplyNodeOutputs,
	imagegraph.NodeTypeOutput:         generateOutputNodeOutputs,
}

func generateBlurNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	radius, err := event.NodeConfig.GetInt("radius")
	if err != nil {
		return err
	}

	inputImageID, err := event.GetInput("original")
	if err != nil {
		return err
	}

	return imageGen.GenerateOutputsForBlurNode(
		ctx,
		event.ImageGraphID,
		event.NodeID,
		inputImageID,
		radius,
	)
}

func generateCropNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	left, err := event.NodeConfig.GetIntOptional("left")
	if err != nil {
		return err
	}

	right, err := event.NodeConfig.GetIntOptional("right")
	if err != nil {
		return err
	}

	top, err := event.NodeConfig.GetIntOptional("top")
	if err != nil {
		return err
	}

	bottom, err := event.NodeConfig.GetIntOptional("bottom")
	if err != nil {
		return err
	}

	inputImageID, err := event.GetInput("original")
	if err != nil {
		return err
	}

	return imageGen.GenerateOutputsForCropNode(
		ctx,
		event.ImageGraphID,
		event.NodeID,
		inputImageID,
		left,
		right,
		top,
		bottom,
	)
}

func generateResizeNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	width, err := event.NodeConfig.GetIntOptional("width")
	if err != nil {
		return err
	}

	height, err := event.NodeConfig.GetIntOptional("height")
	if err != nil {
		return err
	}

	interpolation, err := event.NodeConfig.GetString("interpolation")
	if err != nil {
		return err
	}

	inputImageID, err := event.GetInput("original")
	if err != nil {
		return err
	}

	return imageGen.GenerateOutputsForResizeNode(
		ctx,
		event.ImageGraphID,
		event.NodeID,
		inputImageID,
		width,
		height,
		interpolation,
	)
}

func generateResizeMatchNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	originalImageID, err := event.GetInput("original")
	if err != nil {
		return err
	}

	sizeMatchImageID, err := event.GetInput("size_match")
	if err != nil {
		return err
	}

	interpolation, err := event.NodeConfig.GetString("interpolation")
	if err != nil {
		return err
	}

	return imageGen.GenerateOutputsForResizeMatchNode(
		ctx,
		event.ImageGraphID,
		event.NodeID,
		originalImageID,
		sizeMatchImageID,
		interpolation,
	)
}

func generatePixelInflateNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	width, err := event.NodeConfig.GetInt("width")
	if err != nil {
		return err
	}

	lineWidth, err := event.NodeConfig.GetInt("line_width")
	if err != nil {
		return err
	}

	lineColor, err := event.NodeConfig.GetString("line_color")
	if err != nil {
		return err
	}

	inputImageID, err := event.GetInput("original")
	if err != nil {
		return err
	}

	return imageGen.GenerateOutputsForPixelInflateNode(
		ctx,
		event.ImageGraphID,
		event.NodeID,
		inputImageID,
		width,
		lineWidth,
		lineColor,
	)
}

func generatePaletteExtractNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	numColors, err := event.NodeConfig.GetInt("num_colors")
	if err != nil {
		return err
	}

	clusterBy, err := event.NodeConfig.GetString("cluster_by")
	if err != nil {
		return err
	}

	sourceImageID, err := event.GetInput("source")
	if err != nil {
		return err
	}

	return imageGen.GenerateOutputsForPaletteExtractNode(
		ctx,
		event.ImageGraphID,
		event.NodeID,
		sourceImageID,
		numColors,
		clusterBy,
	)
}

func generatePaletteApplyNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	sourceImageID, err := event.GetInput("source")
	if err != nil {
		return err
	}

	paletteImageID, err := event.GetInput("palette")
	if err != nil {
		return err
	}

	return imageGen.GenerateOutputsForPaletteApplyNode(
		ctx,
		event.ImageGraphID,
		event.NodeID,
		sourceImageID,
		paletteImageID,
	)
}

func generateOutputNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	inputImageID, err := event.GetInput("input")
	if err != nil {
		return err
	}

	return imageGen.GenerateOutputsForOutputNode(
		ctx,
		event.ImageGraphID,
		event.NodeID,
		inputImageID,
	)
}
