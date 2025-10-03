package application

import (
	"context"
	"fmt"

	"github.com/dmpettyp/dorky"
)

type ImageGraphCommandHandlers struct {
}

// NewImageGraphCommandHandlers initializes the handlers struct that processes
// all ImageGraph Commands and registers all handlers with the provided
// message bus
func NewImageGraphCommandHandlers(
	messageBus *dorky.MessageBus,
) (
	*ImageGraphCommandHandlers,
	error,
) {
	handlers := &ImageGraphCommandHandlers{}

	err := dorky.RegisterCommandHandler(
		messageBus,
		handlers.HandleCreateImageGraphCommand,
	)

	if err != nil {
		return nil, fmt.Errorf("could not create image graph command handlers: %w", err)
	}

	return handlers, nil
}

func (h *ImageGraphCommandHandlers) HandleCreateImageGraphCommand(
	context.Context,
	*CreateImageGraphCommand,
) (
	[]dorky.Event,
	error,
) {
	return nil, nil
}
