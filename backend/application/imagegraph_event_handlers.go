package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/infrastructure/imagegen"
	"github.com/dmpettyp/dorky"
)

// ImageGraphNotifier is an interface for broadcasting graph notifications
type ImageGraphNotifier interface {
	BroadcastNodeUpdate(graphID imagegraph.ImageGraphID, nodeUpdate interface{})
	BroadcastLayoutUpdate(graphID imagegraph.ImageGraphID)
}

type imageRemover interface {
	Remove(imageID imagegraph.ImageID) error
}

type ImageGraphEventHandlers struct {
	uow          UnitOfWork
	imageGen     *imagegen.ImageGen
	imageRemover imageRemover
	notifier     ImageGraphNotifier
}

// NewImageGraphEventHandlers initializes the handlers struct that processes
// all ImageGraph Events and registers all handlers with the provided
// message bus
func NewImageGraphEventHandlers(
	mb *dorky.MessageBus,
	uow UnitOfWork,
	imageGen *imagegen.ImageGen,
	imageRemover imageRemover,
	notifier ImageGraphNotifier,
) (
	*ImageGraphEventHandlers,
	error,
) {
	handlers := &ImageGraphEventHandlers{
		uow:          uow,
		imageGen:     imageGen,
		imageRemover: imageRemover,
		notifier:     notifier,
	}

	err := errors.Join(
		dorky.RegisterEventHandler(mb, handlers.HandleNodeOutputImageUnsetEvent),
		dorky.RegisterEventHandler(mb, handlers.HandleNodePreviewSetEvent),
		dorky.RegisterEventHandler(mb, handlers.HandleNodeNeedsOutputsEvent),
		dorky.RegisterEventHandler(mb, handlers.HandleNodeOutputImageSetEvent),
		dorky.RegisterEventHandler(mb, handlers.HandleNodeAddedEvent),
		dorky.RegisterEventHandler(mb, handlers.HandleNodeRemovedEvent),
		dorky.RegisterEventHandler(mb, handlers.HandleNodeInputConnectedEvent),
		dorky.RegisterEventHandler(mb, handlers.HandleNodeInputDisconnectedEvent),
	)

	if err != nil {
		return nil, fmt.Errorf("could not create image graph event handlers: %w", err)
	}

	return handlers, nil
}

func (h *ImageGraphEventHandlers) HandleNodeOutputImageUnsetEvent(
	ctx context.Context,
	event *imagegraph.NodeOutputImageUnsetEvent,
) (
	[]dorky.Event,
	error,
) {
	if err := h.imageRemover.Remove(event.ImageID); err != nil {
		return nil, fmt.Errorf(
			"could not process NodeOutputImageUnsetEvent for ImageGraph %q: %w",
			event.ImageGraphID, err,
		)
	}

	return h.uow.Run(ctx, func(repos *Repos) error {
		ig, err := repos.ImageGraphRepository.Get(event.ImageGraphID)

		if err != nil {
			return fmt.Errorf("could not process NodeOutputImageUnsetEvent for ImageGraph %q: %w", event.ImageGraphID, err)
		}

		err = ig.UnsetNodeOutputConnections(
			event.NodeID,
			event.OutputName,
		)

		if err != nil {
			return fmt.Errorf("could not process NodeOutputImageUnsetEvent for ImageGraph %q: %w", event.ImageGraphID, err)
		}

		return nil
	})
}

// type Registry struct {
// 	mappings map[imagegraph.NodeType]func(
// }

func (h *ImageGraphEventHandlers) HandleNodeNeedsOutputsEvent(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
) (
	[]dorky.Event,
	error,
) {
	// Broadcast that node is processing
	h.notifier.BroadcastNodeUpdate(event.ImageGraphID, map[string]interface{}{
		"node_id": event.NodeID.String(),
		"state":   "processing",
	})

	if event.NodeType == imagegraph.NodeTypeBlur {
		radius, err := event.NodeConfig.GetInt("radius")

		if err != nil {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: %w", err)
		}

		// Find the "original" input
		var inputImageID imagegraph.ImageID
		for _, input := range event.Inputs {
			if input.Name == "original" {
				inputImageID = input.ImageID
				break
			}
		}

		if inputImageID.IsNil() {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: missing 'original' input")
		}

		go func() {
			_ = h.imageGen.GenerateOutputsForBlurNode(
				ctx,
				event.ImageGraphID,
				event.NodeID,
				inputImageID,
				radius,
				"blurred",
			)
		}()
	}

	if event.NodeType == imagegraph.NodeTypeCrop {
		// Extract optional crop coordinates
		left, err := event.NodeConfig.GetIntOptional("left")
		if err != nil {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: %w", err)
		}

		right, err := event.NodeConfig.GetIntOptional("right")
		if err != nil {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: %w", err)
		}

		top, err := event.NodeConfig.GetIntOptional("top")
		if err != nil {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: %w", err)
		}

		bottom, err := event.NodeConfig.GetIntOptional("bottom")
		if err != nil {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: %w", err)
		}

		// Find the "original" input
		var inputImageID imagegraph.ImageID
		for _, input := range event.Inputs {
			if input.Name == "original" {
				inputImageID = input.ImageID
				break
			}
		}

		if inputImageID.IsNil() {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: missing 'original' input")
		}

		go func() {
			_ = h.imageGen.GenerateOutputsForCropNode(
				ctx,
				event.ImageGraphID,
				event.NodeID,
				inputImageID,
				left,
				right,
				top,
				bottom,
				"cropped",
			)
		}()
	}

	if event.NodeType == imagegraph.NodeTypeResize {
		// Extract width and height (either or both may be present)
		width, err := event.NodeConfig.GetIntOptional("width")
		if err != nil {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: %w", err)
		}

		height, err := event.NodeConfig.GetIntOptional("height")
		if err != nil {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: %w", err)
		}

		interpolation, err := event.NodeConfig.GetString("interpolation")
		if err != nil {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: %w", err)
		}

		// Find the "original" input
		var inputImageID imagegraph.ImageID
		for _, input := range event.Inputs {
			if input.Name == "original" {
				inputImageID = input.ImageID
				break
			}
		}

		if inputImageID.IsNil() {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: missing 'original' input")
		}

		go func() {
			_ = h.imageGen.GenerateOutputsForResizeNode(
				ctx,
				event.ImageGraphID,
				event.NodeID,
				inputImageID,
				width,
				height,
				interpolation,
				"resized",
			)
		}()
	}

	if event.NodeType == imagegraph.NodeTypeResizeMatch {
		// Find the "original" input
		var originalImageID imagegraph.ImageID
		for _, input := range event.Inputs {
			if input.Name == "original" {
				originalImageID = input.ImageID
				break
			}
		}

		// Find the "size_match" input
		var sizeMatchImageID imagegraph.ImageID
		for _, input := range event.Inputs {
			if input.Name == "size_match" {
				sizeMatchImageID = input.ImageID
				break
			}
		}

		interpolation, err := event.NodeConfig.GetString("interpolation")
		if err != nil {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: %w", err)
		}

		if originalImageID.IsNil() {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: missing 'original' input")
		}

		if sizeMatchImageID.IsNil() {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: missing 'size_match' input")
		}

		go func() {
			_ = h.imageGen.GenerateOutputsForResizeMatchNode(
				ctx,
				event.ImageGraphID,
				event.NodeID,
				originalImageID,
				sizeMatchImageID,
				interpolation,
				"resized",
			)
		}()
	}

	if event.NodeType == imagegraph.NodeTypePixelInflate {
		width, err := event.NodeConfig.GetInt("width")
		if err != nil {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: %w", err)
		}

		lineWidth, err := event.NodeConfig.GetInt("line_width")
		if err != nil {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: %w", err)
		}

		lineColor, err := event.NodeConfig.GetString("line_color")
		if err != nil {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: %w", err)
		}

		// Find the "original" input
		var inputImageID imagegraph.ImageID
		for _, input := range event.Inputs {
			if input.Name == "original" {
				inputImageID = input.ImageID
				break
			}
		}

		if inputImageID.IsNil() {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: missing 'original' input")
		}

		go func() {
			_ = h.imageGen.GenerateOutputsForPixelInflateNode(
				ctx,
				event.ImageGraphID,
				event.NodeID,
				inputImageID,
				width,
				lineWidth,
				lineColor,
				"inflated",
			)
		}()
	}

	if event.NodeType == imagegraph.NodeTypePaletteExtract {
		// Extract config values
		numColors, err := event.NodeConfig.GetInt("num_colors")
		if err != nil {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: %w", err)
		}

		clusterBy, err := event.NodeConfig.GetString("cluster_by")
		if err != nil {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: %w", err)
		}

		// Find the "source" input
		var sourceImageID imagegraph.ImageID
		for _, input := range event.Inputs {
			if input.Name == "source" {
				sourceImageID = input.ImageID
				break
			}
		}

		if sourceImageID.IsNil() {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: missing 'source' input")
		}

		// Process asynchronously
		go func() {
			_ = h.imageGen.GenerateOutputsForPaletteExtractNode(
				ctx,
				event.ImageGraphID,
				event.NodeID,
				sourceImageID,
				numColors,
				clusterBy,
				"palette",
			)
		}()
	}

	if event.NodeType == imagegraph.NodeTypePaletteApply {
		// Find the "source" and "palette" inputs
		var sourceImageID, paletteImageID imagegraph.ImageID
		for _, input := range event.Inputs {
			if input.Name == "source" {
				sourceImageID = input.ImageID
			} else if input.Name == "palette" {
				paletteImageID = input.ImageID
			}
		}

		if sourceImageID.IsNil() {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: missing 'source' input")
		}
		if paletteImageID.IsNil() {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: missing 'palette' input")
		}

		// Process asynchronously
		go func() {
			_ = h.imageGen.GenerateOutputsForPaletteApplyNode(
				ctx,
				event.ImageGraphID,
				event.NodeID,
				sourceImageID,
				paletteImageID,
				"mapped",
			)
		}()
	}

	if event.NodeType == imagegraph.NodeTypeOutput {
		// Find the "input" input
		var inputImageID imagegraph.ImageID
		for _, input := range event.Inputs {
			if input.Name == "input" {
				inputImageID = input.ImageID
				break
			}
		}

		if inputImageID.IsNil() {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent: missing 'input' input")
		}

		go func() {
			_ = h.imageGen.GenerateOutputsForOutputNode(
				ctx,
				event.ImageGraphID,
				event.NodeID,
				inputImageID,
				"final",
			)
		}()
	}

	return nil, nil
}

func (h *ImageGraphEventHandlers) HandleNodeOutputImageSetEvent(
	ctx context.Context,
	event *imagegraph.NodeOutputImageSetEvent,
) (
	[]dorky.Event,
	error,
) {
	// Broadcast that node output is complete
	h.notifier.BroadcastNodeUpdate(event.ImageGraphID, map[string]interface{}{
		"node_id": event.NodeID.String(),
		"state":   "completed",
		"outputs": map[string]interface{}{
			string(event.OutputName): event.ImageID.String(),
		},
	})

	if event.NodeType == imagegraph.NodeTypeInput {
		go func() {
			_ = h.imageGen.GeneratePreviewForInputNode(
				ctx,
				event.ImageGraphID,
				event.NodeID,
				event.ImageID,
			)
		}()
	}

	return nil, nil
}

func (h *ImageGraphEventHandlers) HandleNodePreviewSetEvent(
	ctx context.Context,
	event *imagegraph.NodePreviewSetEvent,
) (
	[]dorky.Event,
	error,
) {
	h.notifier.BroadcastNodeUpdate(event.ImageGraphID, map[string]interface{}{
		"node_id": event.NodeID.String(),
	})

	return nil, nil
}

func (h *ImageGraphEventHandlers) HandleNodeAddedEvent(
	ctx context.Context,
	event *imagegraph.NodeAddedEvent,
) (
	[]dorky.Event,
	error,
) {
	// Broadcast that node was added
	h.notifier.BroadcastNodeUpdate(event.ImageGraphID, map[string]interface{}{
		"node_id": event.NodeID.String(),
		"state":   "added",
	})

	return nil, nil
}

func (h *ImageGraphEventHandlers) HandleNodeRemovedEvent(
	ctx context.Context,
	event *imagegraph.NodeRemovedEvent,
) (
	[]dorky.Event,
	error,
) {
	// Broadcast that node was removed
	h.notifier.BroadcastNodeUpdate(event.ImageGraphID, map[string]interface{}{
		"node_id": event.NodeID.String(),
		"state":   "removed",
	})

	return nil, nil
}

func (h *ImageGraphEventHandlers) HandleNodeInputConnectedEvent(
	ctx context.Context,
	event *imagegraph.NodeInputConnectedEvent,
) (
	[]dorky.Event,
	error,
) {
	// Broadcast that connection was made
	h.notifier.BroadcastNodeUpdate(event.ImageGraphID, map[string]interface{}{
		"node_id": event.NodeID.String(),
		"state":   "connected",
	})

	return nil, nil
}

func (h *ImageGraphEventHandlers) HandleNodeInputDisconnectedEvent(
	ctx context.Context,
	event *imagegraph.NodeInputDisconnectedEvent,
) (
	[]dorky.Event,
	error,
) {
	// Broadcast that connection was removed
	h.notifier.BroadcastNodeUpdate(event.ImageGraphID, map[string]interface{}{
		"node_id": event.NodeID.String(),
		"state":   "disconnected",
	})

	return nil, nil
}
