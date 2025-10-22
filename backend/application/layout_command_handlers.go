package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmpettyp/artwork/domain/ui"
	"github.com/dmpettyp/dorky"
)

type LayoutCommandHandlers struct {
	uow UnitOfWork
}

// NewLayoutCommandHandlers initializes the handlers struct that processes
// all Layout Commands and registers all handlers with the provided
// message bus
func NewLayoutCommandHandlers(
	mb *dorky.MessageBus,
	uow UnitOfWork,
) (
	*LayoutCommandHandlers,
	error,
) {
	handlers := &LayoutCommandHandlers{uow: uow}

	err := dorky.RegisterCommandHandler(mb, handlers.HandleUpdateLayoutCommand)

	if err != nil {
		return nil, fmt.Errorf("could not create layout command handlers: %w", err)
	}

	return handlers, nil
}

func (h *LayoutCommandHandlers) HandleUpdateLayoutCommand(
	ctx context.Context,
	command *UpdateLayoutCommand,
) (
	[]dorky.Event,
	error,
) {
	return h.uow.Run(ctx, func(repos *Repos) error {
		// Try to get existing layout, or create and add new if it doesn't exist
		layout, err := repos.LayoutRepository.Get(command.GraphID)

		if err != nil {
			if !errors.Is(err, ErrLayoutNotFound) {
				return fmt.Errorf("could not get Layout for ImageGraph %q: %w", command.GraphID, err)
			}

			layout, err = ui.NewLayout(command.GraphID)
			if err != nil {
				return fmt.Errorf("could not create Layout for ImageGraph %q: %w", command.GraphID, err)
			}

			err = repos.LayoutRepository.Add(layout)
			if err != nil {
				return fmt.Errorf("could not add Layout for ImageGraph %q: %w", command.GraphID, err)
			}
		}

		// Replace all node positions
		layout.NodePositions = command.NodePositions

		return nil
	})
}
