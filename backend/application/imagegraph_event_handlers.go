package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/infrastructure/imagegen"
	"github.com/dmpettyp/dorky/messagebus"
	"github.com/dmpettyp/dorky/messages"
)

// ImageGraphNotifier is an interface for broadcasting graph notifications
type ImageGraphNotifier interface {
	BroadcastNodeUpdate(graphID imagegraph.ImageGraphID, nodeUpdate any)
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
	mb *messagebus.MessageBus,
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
		messagebus.RegisterEventHandler(mb, handlers.HandleNodeAddedEvent),
		messagebus.RegisterEventHandler(mb, handlers.HandleNodeInputConnectedEvent),
		messagebus.RegisterEventHandler(mb, handlers.HandleNodeInputDisconnectedEvent),
		messagebus.RegisterEventHandler(mb, handlers.HandleNodeNeedsOutputsEvent),
		messagebus.RegisterEventHandler(mb, handlers.HandleNodeOutputImageSetEvent),
		messagebus.RegisterEventHandler(mb, handlers.HandleNodeOutputImageUnsetEvent),
		messagebus.RegisterEventHandler(mb, handlers.HandleNodePreviewSetEvent),
		messagebus.RegisterEventHandler(mb, handlers.HandleNodeRemovedEvent),
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
	[]messages.Event,
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

func (h *ImageGraphEventHandlers) HandleNodeNeedsOutputsEvent(
	ctx context.Context,
	event *imagegraph.NodeNeedsOutputsEvent,
) (
	[]messages.Event,
	error,
) {
	h.notifier.BroadcastNodeUpdate(event.ImageGraphID, map[string]any{
		"node_id": event.NodeID.String(),
		"state":   "processing",
	})

	generator, ok := nodeOutputGenerators[event.NodeType]
	if !ok {
		return nil, fmt.Errorf(
			"no output generator registered for node type %q",
			event.NodeType,
		)
	}

	go func() {
		err := generator(ctx, event, h.imageGen)

		if err != nil {
			fmt.Println(err)
		}
	}()

	return nil, nil
}

func (h *ImageGraphEventHandlers) HandleNodeOutputImageSetEvent(
	ctx context.Context,
	event *imagegraph.NodeOutputImageSetEvent,
) (
	[]messages.Event,
	error,
) {
	h.notifier.BroadcastNodeUpdate(event.ImageGraphID, map[string]any{
		"node_id": event.NodeID.String(),
		"state":   "completed",
		"outputs": map[string]any{
			string(event.OutputName): event.ImageID.String(),
		},
	})

	if event.NodeType == imagegraph.NodeTypeInput {
		go func() {
			_ = h.imageGen.GeneratePreviewForInputNode(
				ctx,
				event.ImageGraphID,
				event.NodeID,
				event.NodeVersion,
				event.ImageID,
			)
		}()
	}

	// Propagate output image to downstream nodes
	return h.uow.Run(ctx, func(repos *Repos) error {
		ig, err := repos.ImageGraphRepository.Get(event.ImageGraphID)
		if err != nil {
			return fmt.Errorf(
				"could not process NodeOutputImageSetEvent for ImageGraph %q: %w",
				event.ImageGraphID, err,
			)
		}

		err = ig.PropagateOutputImageToConnections(
			event.NodeID,
			event.OutputName,
			event.ImageID,
		)

		if err != nil {
			return fmt.Errorf(
				"could not process NodeOutputImageSetEvent for ImageGraph %q: %w",
				event.ImageGraphID, err,
			)
		}

		return nil
	})
}

func (h *ImageGraphEventHandlers) HandleNodePreviewSetEvent(
	ctx context.Context,
	event *imagegraph.NodePreviewSetEvent,
) (
	[]messages.Event,
	error,
) {
	h.notifier.BroadcastNodeUpdate(event.ImageGraphID, map[string]any{
		"node_id": event.NodeID.String(),
	})

	return nil, nil
}

func (h *ImageGraphEventHandlers) HandleNodeAddedEvent(
	ctx context.Context,
	event *imagegraph.NodeAddedEvent,
) (
	[]messages.Event,
	error,
) {
	// Broadcast that node was added
	h.notifier.BroadcastNodeUpdate(event.ImageGraphID, map[string]any{
		"node_id": event.NodeID.String(),
		"state":   "added",
	})

	return nil, nil
}

func (h *ImageGraphEventHandlers) HandleNodeRemovedEvent(
	ctx context.Context,
	event *imagegraph.NodeRemovedEvent,
) (
	[]messages.Event,
	error,
) {
	// Broadcast that node was removed
	h.notifier.BroadcastNodeUpdate(event.ImageGraphID, map[string]any{
		"node_id": event.NodeID.String(),
		"state":   "removed",
	})

	return nil, nil
}

func (h *ImageGraphEventHandlers) HandleNodeInputConnectedEvent(
	ctx context.Context,
	event *imagegraph.NodeInputConnectedEvent,
) (
	[]messages.Event,
	error,
) {
	// Broadcast that connection was made
	h.notifier.BroadcastNodeUpdate(event.ImageGraphID, map[string]any{
		"node_id": event.NodeID.String(),
		"state":   "connected",
	})

	return nil, nil
}

func (h *ImageGraphEventHandlers) HandleNodeInputDisconnectedEvent(
	ctx context.Context,
	event *imagegraph.NodeInputDisconnectedEvent,
) (
	[]messages.Event,
	error,
) {
	// Broadcast that connection was removed
	h.notifier.BroadcastNodeUpdate(event.ImageGraphID, map[string]any{
		"node_id": event.NodeID.String(),
		"state":   "disconnected",
	})

	return nil, nil
}
