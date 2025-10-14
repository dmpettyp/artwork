package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmpettyp/artwork/domain/ui"
	"github.com/dmpettyp/dorky"
)

type UIMetadataCommandHandlers struct {
	uow UnitOfWork
}

// NewUIMetadataCommandHandlers initializes the handlers struct that processes
// all UI Metadata Commands and registers all handlers with the provided
// message bus
func NewUIMetadataCommandHandlers(
	mb *dorky.MessageBus,
	uow UnitOfWork,
) (
	*UIMetadataCommandHandlers,
	error,
) {
	handlers := &UIMetadataCommandHandlers{uow: uow}

	err := dorky.RegisterCommandHandler(mb, handlers.HandleUpdateUIMetadataCommand)

	if err != nil {
		return nil, fmt.Errorf("could not create UI metadata command handlers: %w", err)
	}

	return handlers, nil
}

func (h *UIMetadataCommandHandlers) HandleUpdateUIMetadataCommand(
	ctx context.Context,
	command *UpdateUIMetadataCommand,
) (
	[]dorky.Event,
	error,
) {
	return h.uow.Run(ctx, func(repos *Repos) error {
		// Try to get existing metadata, or create and add new if it doesn't exist
		metadata, err := repos.UIMetadataRepository.Get(command.GraphID)

		if err != nil {
			if errors.Is(err, ErrUIMetadataNotFound) {
				// Create new metadata
				metadata, err = ui.NewUIMetadata(command.GraphID)
				if err != nil {
					return fmt.Errorf("could not create UIMetadata for ImageGraph %q: %w", command.GraphID, err)
				}

				// Add it to the repository
				err = repos.UIMetadataRepository.Add(metadata)
				if err != nil {
					return fmt.Errorf("could not add UIMetadata for ImageGraph %q: %w", command.GraphID, err)
				}
			} else {
				return fmt.Errorf("could not get UIMetadata for ImageGraph %q: %w", command.GraphID, err)
			}
		}

		// Update viewport (mutates the domain object directly)
		err = metadata.SetViewport(command.Zoom, command.PanX, command.PanY)
		if err != nil {
			return fmt.Errorf("could not update viewport for ImageGraph %q: %w", command.GraphID, err)
		}

		// Replace all node positions
		metadata.NodePositions = command.NodePositions

		return nil
	})
}
