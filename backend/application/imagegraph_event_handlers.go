package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/infrastructure/imagegen"
	"github.com/dmpettyp/dorky"
)

type ImageGraphEventHandlers struct {
	uow      UnitOfWork
	imageGen *imagegen.ImageGen
}

// NewImageGraphEventHandlers initializes the handlers struct that processes
// all ImageGraph Events and registers all handlers with the provided
// message bus
func NewImageGraphEventHandlers(
	mb *dorky.MessageBus,
	uow UnitOfWork,
	imageGen *imagegen.ImageGen,
) (
	*ImageGraphEventHandlers,
	error,
) {
	handlers := &ImageGraphEventHandlers{
		uow:      uow,
		imageGen: imageGen,
	}

	err := errors.Join(
		dorky.RegisterEventHandler(mb, handlers.HandleNodeOutputImageUnsetEvent),
		dorky.RegisterEventHandler(mb, handlers.HandleNodeNeedsOutputsEvent),
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
	return h.uow.Run(ctx, func(repos *Repos) error {
		ig, err := repos.ImageGraphRepository.Get(event.ImageGraphID)

		if err != nil {
			return fmt.Errorf("could not process NodeNeedsOutputsEvent for ImageGraph %q: %w", event.ImageGraphID, err)
		}

		if event.NodeType == imagegraph.NodeTypeScale {
			err = ig.SetNodeOutputImage(event.NodeID, "scaled", event.Inputs[0].ImageID)

			if err != nil {
				return fmt.Errorf("could not process NodeNeedsOutputsEvent for ImageGraph %q: %w", event.ImageGraphID, err)
			}
		}

		return nil
	})
}
