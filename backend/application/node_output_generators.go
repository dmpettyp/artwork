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
	imagegraph.NodeTypePaletteCreate:  generatePaletteCreateNodeOutputs,
	imagegraph.NodeTypePaletteEdit:    generatePaletteEditNodeOutputs,
	imagegraph.NodeTypeOutput:         generateOutputNodeOutputs,
}

func generateBlurNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	config, ok := event.NodeConfig.(*imagegraph.NodeConfigBlur)
	if !ok {
		return fmt.Errorf("invalid config provided to generate Blur Node outputs")
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
		config.Radius,
	)
}

func generateCropNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	config, ok := event.NodeConfig.(*imagegraph.NodeConfigCrop)
	if !ok {
		return fmt.Errorf("invalid config provided to generate Crop Node outputs")
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
		config.Left,
		config.Right,
		config.Top,
		config.Bottom,
	)
}

func generateResizeNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	config, ok := event.NodeConfig.(*imagegraph.NodeConfigResize)
	if !ok {
		return fmt.Errorf("invalid config provided to generate Resize Node outputs")
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
		config.Width,
		config.Height,
		config.Interpolation,
	)
}

func generateResizeMatchNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	config, ok := event.NodeConfig.(*imagegraph.NodeConfigResizeMatch)
	if !ok {
		return fmt.Errorf("invalid config provided to generate ResizeMatch Node outputs")
	}

	originalImageID, err := event.GetInput("original")
	if err != nil {
		return err
	}

	sizeMatchImageID, err := event.GetInput("size_match")
	if err != nil {
		return err
	}

	return imageGen.GenerateOutputsForResizeMatchNode(
		ctx,
		event.ImageGraphID,
		event.NodeID,
		originalImageID,
		sizeMatchImageID,
		config.Interpolation,
	)
}

func generatePixelInflateNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	config, ok := event.NodeConfig.(*imagegraph.NodeConfigPixelInflate)
	if !ok {
		return fmt.Errorf("invalid config provided to generate PixelInflate Node outputs")
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
		config.Width,
		config.LineWidth,
		config.LineColor,
	)
}

func generatePaletteExtractNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	config, ok := event.NodeConfig.(*imagegraph.NodeConfigPaletteExtract)
	if !ok {
		return fmt.Errorf("invalid config provided to generate PaletteExtract Node outputs")
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
		config.NumColors,
		config.Method,
	)
}

func generatePaletteApplyNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	config, ok := event.NodeConfig.(*imagegraph.NodeConfigPaletteApply)
	if !ok {
		return fmt.Errorf("invalid config provided to generate PaletteApply Node outputs")
	}

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
		config,
	)
}

func generatePaletteCreateNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	config, ok := event.NodeConfig.(*imagegraph.NodeConfigPaletteCreate)
	if !ok {
		return fmt.Errorf("invalid config provided to generate PaletteCreate Node outputs")
	}

	colors, err := config.ColorsList()
	if err != nil {
		return err
	}

	return imageGen.GenerateOutputsForPaletteCreateNode(
		ctx,
		event.ImageGraphID,
		event.NodeID,
		colors,
	)
}

func generatePaletteEditNodeOutputs(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
	imageGen *imagegen.ImageGen,
) error {
	config, ok := event.NodeConfig.(*imagegraph.NodeConfigPaletteEdit)
	if !ok {
		return fmt.Errorf("invalid config provided to generate PaletteEdit Node outputs")
	}

	sourceImageID, err := event.GetInput("source")
	if err != nil {
		return err
	}

	rawList, err := config.ColorsRawList()
	if err != nil {
		return err
	}

	return imageGen.GenerateOutputsForPaletteEditNode(
		ctx,
		event.ImageGraphID,
		event.NodeID,
		sourceImageID,
		rawList,
		config.Colors,
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
