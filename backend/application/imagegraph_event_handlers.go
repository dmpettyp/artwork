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

type ImageGraphEventHandlers struct {
	uow      UnitOfWork
	imageGen *imagegen.ImageGen
	notifier ImageGraphNotifier
}

// NewImageGraphEventHandlers initializes the handlers struct that processes
// all ImageGraph Events and registers all handlers with the provided
// message bus
func NewImageGraphEventHandlers(
	mb *dorky.MessageBus,
	uow UnitOfWork,
	imageGen *imagegen.ImageGen,
	notifier ImageGraphNotifier,
) (
	*ImageGraphEventHandlers,
	error,
) {
	handlers := &ImageGraphEventHandlers{
		uow:      uow,
		imageGen: imageGen,
		notifier: notifier,
	}

	err := errors.Join(
		dorky.RegisterEventHandler(mb, handlers.HandleNodeOutputImageUnsetEvent),
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
		radius := event.NodeConfig["radius"].(int)

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

	if event.NodeType == imagegraph.NodeTypeResize {
		// Extract width and height (either or both may be present)
		var width, height *int
		if w, ok := event.NodeConfig["width"]; ok {
			wInt := w.(int)
			width = &wInt
		}
		if h, ok := event.NodeConfig["height"]; ok {
			hInt := h.(int)
			height = &hInt
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
				"resized",
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

		// Passive passthrough - just set the output to the same image as the input
		_, err := h.uow.Run(ctx, func(repos *Repos) error {
			ig, err := repos.ImageGraphRepository.Get(event.ImageGraphID)
			if err != nil {
				return fmt.Errorf("could not get image graph: %w", err)
			}

			err = ig.SetNodeOutputImage(event.NodeID, "final", inputImageID)
			if err != nil {
				return fmt.Errorf("could not set output image: %w", err)
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("could not process NodeNeedsOutputsEvent for output node: %w", err)
		}
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
