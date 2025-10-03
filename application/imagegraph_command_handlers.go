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

	err = dorky.RegisterCommandHandler(
		messageBus,
		handlers.HandleAddImageGraphNodeCommand,
	)

	if err != nil {
		return nil, fmt.Errorf("could not create image graph command handlers: %w", err)
	}

	err = dorky.RegisterCommandHandler(
		messageBus,
		handlers.HandleRemoveImageGraphNodeCommand,
	)

	if err != nil {
		return nil, fmt.Errorf("could not create image graph command handlers: %w", err)
	}

	err = dorky.RegisterCommandHandler(
		messageBus,
		handlers.HandleConnectImageGraphNodesCommand,
	)

	if err != nil {
		return nil, fmt.Errorf("could not create image graph command handlers: %w", err)
	}

	err = dorky.RegisterCommandHandler(
		messageBus,
		handlers.HandleDisconnectImageGraphNodesCommand,
	)

	if err != nil {
		return nil, fmt.Errorf("could not create image graph command handlers: %w", err)
	}

	err = dorky.RegisterCommandHandler(
		messageBus,
		handlers.HandleSetImageGraphNodeOutputImageCommand,
	)

	if err != nil {
		return nil, fmt.Errorf("could not create image graph command handlers: %w", err)
	}

	err = dorky.RegisterCommandHandler(
		messageBus,
		handlers.HandleUnsetImageGraphNodeOutputImageCommand,
	)

	if err != nil {
		return nil, fmt.Errorf("could not create image graph command handlers: %w", err)
	}

	err = dorky.RegisterCommandHandler(
		messageBus,
		handlers.HandleSetImageGraphNodePreviewCommand,
	)

	if err != nil {
		return nil, fmt.Errorf("could not create image graph command handlers: %w", err)
	}

	err = dorky.RegisterCommandHandler(
		messageBus,
		handlers.HandleUnsetImageGraphNodePreviewCommand,
	)

	if err != nil {
		return nil, fmt.Errorf("could not create image graph command handlers: %w", err)
	}

	return handlers, nil
}

func (h *ImageGraphCommandHandlers) HandleCreateImageGraphCommand(
	_ context.Context,
	command *CreateImageGraphCommand,
) (
	[]dorky.Event,
	error,
) {
	fmt.Println("creating a new image graph", command)
	return nil, nil
}

func (h *ImageGraphCommandHandlers) HandleAddImageGraphNodeCommand(
	context.Context,
	*AddImageGraphNodeCommand,
) (
	[]dorky.Event,
	error,
) {
	return nil, nil
}

func (h *ImageGraphCommandHandlers) HandleRemoveImageGraphNodeCommand(
	context.Context,
	*RemoveImageGraphNodeCommand,
) (
	[]dorky.Event,
	error,
) {
	return nil, nil
}

func (h *ImageGraphCommandHandlers) HandleConnectImageGraphNodesCommand(
	context.Context,
	*ConnectImageGraphNodesCommand,
) (
	[]dorky.Event,
	error,
) {
	return nil, nil
}

func (h *ImageGraphCommandHandlers) HandleDisconnectImageGraphNodesCommand(
	context.Context,
	*DisconnectImageGraphNodesCommand,
) (
	[]dorky.Event,
	error,
) {
	return nil, nil
}

func (h *ImageGraphCommandHandlers) HandleSetImageGraphNodeOutputImageCommand(
	context.Context,
	*SetImageGraphNodeOutputImageCommand,
) (
	[]dorky.Event,
	error,
) {
	return nil, nil
}

func (h *ImageGraphCommandHandlers) HandleUnsetImageGraphNodeOutputImageCommand(
	context.Context,
	*UnsetImageGraphNodeOutputImageCommand,
) (
	[]dorky.Event,
	error,
) {
	return nil, nil
}

func (h *ImageGraphCommandHandlers) HandleSetImageGraphNodePreviewCommand(
	context.Context,
	*SetImageGraphNodePreviewCommand,
) (
	[]dorky.Event,
	error,
) {
	return nil, nil
}

func (h *ImageGraphCommandHandlers) HandleUnsetImageGraphNodePreviewCommand(
	context.Context,
	*UnsetImageGraphNodePreviewCommand,
) (
	[]dorky.Event,
	error,
) {
	return nil, nil
}
