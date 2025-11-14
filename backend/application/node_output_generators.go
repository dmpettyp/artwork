package application

import (
	"context"
	"fmt"

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

	var inputImageID imagegraph.ImageID
	for _, input := range event.Inputs {
		if input.Name == "original" {
			inputImageID = input.ImageID
			break
		}
	}

	if inputImageID.IsNil() {
		return fmt.Errorf("missing 'original' input")
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

	var inputImageID imagegraph.ImageID
	for _, input := range event.Inputs {
		if input.Name == "original" {
			inputImageID = input.ImageID
			break
		}
	}

	if inputImageID.IsNil() {
		return fmt.Errorf("missing 'original' input")
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

	var inputImageID imagegraph.ImageID
	for _, input := range event.Inputs {
		if input.Name == "original" {
			inputImageID = input.ImageID
			break
		}
	}

	if inputImageID.IsNil() {
		return fmt.Errorf("missing 'original' input")
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
	var originalImageID imagegraph.ImageID
	for _, input := range event.Inputs {
		if input.Name == "original" {
			originalImageID = input.ImageID
			break
		}
	}

	var sizeMatchImageID imagegraph.ImageID
	for _, input := range event.Inputs {
		if input.Name == "size_match" {
			sizeMatchImageID = input.ImageID
			break
		}
	}

	interpolation, err := event.NodeConfig.GetString("interpolation")
	if err != nil {
		return err
	}

	if originalImageID.IsNil() {
		return fmt.Errorf("missing 'original' input")
	}

	if sizeMatchImageID.IsNil() {
		return fmt.Errorf("missing 'size_match' input")
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

	var inputImageID imagegraph.ImageID
	for _, input := range event.Inputs {
		if input.Name == "original" {
			inputImageID = input.ImageID
			break
		}
	}

	if inputImageID.IsNil() {
		return fmt.Errorf("missing 'original' input")
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

	var sourceImageID imagegraph.ImageID
	for _, input := range event.Inputs {
		if input.Name == "source" {
			sourceImageID = input.ImageID
			break
		}
	}

	if sourceImageID.IsNil() {
		return fmt.Errorf("missing 'source' input")
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
	var sourceImageID, paletteImageID imagegraph.ImageID
	for _, input := range event.Inputs {
		if input.Name == "source" {
			sourceImageID = input.ImageID
		} else if input.Name == "palette" {
			paletteImageID = input.ImageID
		}
	}

	if sourceImageID.IsNil() {
		return fmt.Errorf("missing 'source' input")
	}

	if paletteImageID.IsNil() {
		return fmt.Errorf("missing 'palette' input")
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
	var inputImageID imagegraph.ImageID
	for _, input := range event.Inputs {
		if input.Name == "input" {
			inputImageID = input.ImageID
			break
		}
	}

	if inputImageID.IsNil() {
		return fmt.Errorf("missing 'input' input")
	}

	return imageGen.GenerateOutputsForOutputNode(
		ctx,
		event.ImageGraphID,
		event.NodeID,
		inputImageID,
	)
}
