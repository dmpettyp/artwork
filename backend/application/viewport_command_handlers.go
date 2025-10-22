package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmpettyp/artwork/domain/ui"
	"github.com/dmpettyp/dorky"
)

type ViewportCommandHandlers struct {
	uow UnitOfWork
}

// NewViewportCommandHandlers initializes the handlers struct that processes
// all Viewport Commands and registers all handlers with the provided
// message bus
func NewViewportCommandHandlers(
	mb *dorky.MessageBus,
	uow UnitOfWork,
) (
	*ViewportCommandHandlers,
	error,
) {
	handlers := &ViewportCommandHandlers{uow: uow}

	err := dorky.RegisterCommandHandler(mb, handlers.HandleUpdateViewportCommand)

	if err != nil {
		return nil, fmt.Errorf("could not create viewport command handlers: %w", err)
	}

	return handlers, nil
}

func (h *ViewportCommandHandlers) HandleUpdateViewportCommand(
	ctx context.Context,
	command *UpdateViewportCommand,
) (
	[]dorky.Event,
	error,
) {
	return h.uow.Run(ctx, func(repos *Repos) error {
		// Try to get existing viewport, or create and add new if it doesn't exist
		viewport, err := repos.ViewportRepository.Get(command.GraphID)

		if err != nil {
			if !errors.Is(err, ErrViewportNotFound) {
				return fmt.Errorf("could not get Viewport for ImageGraph %q: %w", command.GraphID, err)
			}

			viewport, err = ui.NewViewport(command.GraphID)
			if err != nil {
				return fmt.Errorf("could not create Viewport for ImageGraph %q: %w", command.GraphID, err)
			}

			err = repos.ViewportRepository.Add(viewport)
			if err != nil {
				return fmt.Errorf("could not add Viewport for ImageGraph %q: %w", command.GraphID, err)
			}
		}

		// Update viewport using domain method (emits event internally)
		err = viewport.Set(command.Zoom, command.PanX, command.PanY)
		if err != nil {
			return fmt.Errorf("could not update Viewport for ImageGraph %q: %w", command.GraphID, err)
		}

		return nil
	})
}
