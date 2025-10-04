package application

import (
	"context"
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky"
)

type ImageGraphCommandHandlers struct {
	uow UnitOfWork
}

// NewImageGraphCommandHandlers initializes the handlers struct that processes
// all ImageGraph Commands and registers all handlers with the provided
// message bus
func NewImageGraphCommandHandlers(
	messageBus *dorky.MessageBus,
	uow UnitOfWork,
) (
	*ImageGraphCommandHandlers,
	error,
) {
	handlers := &ImageGraphCommandHandlers{uow: uow}

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
	ctx context.Context,
	command *CreateImageGraphCommand,
) (
	[]dorky.Event,
	error,
) {
	return h.uow.Run(ctx, func(repos *Repos) error {
		ig, err := imagegraph.NewImageGraph(command.ImageGraphID, command.Name)

		if err != nil {
			return fmt.Errorf("could not process CreateImageGraphCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		err = repos.ImageGraphRepository.Add(ig)

		if err != nil {
			return fmt.Errorf("could not process CreateImageGraphCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		return nil
	})
}

func (h *ImageGraphCommandHandlers) HandleAddImageGraphNodeCommand(
	ctx context.Context,
	command *AddImageGraphNodeCommand,
) (
	[]dorky.Event,
	error,
) {
	return h.uow.Run(ctx, func(repos *Repos) error {
		ig, err := repos.ImageGraphRepository.Get(command.ImageGraphID)

		if err != nil {
			return fmt.Errorf("could not process AddImageGraphNodeCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		err = ig.AddNode(
			command.NodeID,
			command.NodeType,
			command.Name,
			command.Config,
		)

		if err != nil {
			return fmt.Errorf("could not process AddImageGraphNodeCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		return nil
	})
}

func (h *ImageGraphCommandHandlers) HandleRemoveImageGraphNodeCommand(
	ctx context.Context,
	command *RemoveImageGraphNodeCommand,
) (
	[]dorky.Event,
	error,
) {
	return h.uow.Run(ctx, func(repos *Repos) error {
		ig, err := repos.ImageGraphRepository.Get(command.ImageGraphID)

		if err != nil {
			return fmt.Errorf("could not process RemoveImageGraphNodeCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		err = ig.RemoveNode(command.NodeID)

		if err != nil {
			return fmt.Errorf("could not process RemoveImageGraphNodeCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		return nil
	})
}

func (h *ImageGraphCommandHandlers) HandleConnectImageGraphNodesCommand(
	ctx context.Context,
	command *ConnectImageGraphNodesCommand,
) (
	[]dorky.Event,
	error,
) {
	return h.uow.Run(ctx, func(repos *Repos) error {
		ig, err := repos.ImageGraphRepository.Get(command.ImageGraphID)

		if err != nil {
			return fmt.Errorf("could not process ConnectImageGraphNodesCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		err = ig.ConnectNodes(
			command.FromNodeID,
			command.OutputName,
			command.ToNodeID,
			command.InputName,
		)

		if err != nil {
			return fmt.Errorf("could not process ConnectImageGraphNodesCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		return nil
	})
}

func (h *ImageGraphCommandHandlers) HandleDisconnectImageGraphNodesCommand(
	ctx context.Context,
	command *DisconnectImageGraphNodesCommand,
) (
	[]dorky.Event,
	error,
) {
	return h.uow.Run(ctx, func(repos *Repos) error {
		ig, err := repos.ImageGraphRepository.Get(command.ImageGraphID)

		if err != nil {
			return fmt.Errorf("could not process DisconnectImageGraphNodesCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		err = ig.DisconnectNodes(
			command.FromNodeID,
			command.OutputName,
			command.ToNodeID,
			command.InputName,
		)

		if err != nil {
			return fmt.Errorf("could not process DisconnectImageGraphNodesCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		return nil
	})
}

func (h *ImageGraphCommandHandlers) HandleSetImageGraphNodeOutputImageCommand(
	ctx context.Context,
	command *SetImageGraphNodeOutputImageCommand,
) (
	[]dorky.Event,
	error,
) {
	return h.uow.Run(ctx, func(repos *Repos) error {
		ig, err := repos.ImageGraphRepository.Get(command.ImageGraphID)

		if err != nil {
			return fmt.Errorf("could not process SetImageGraphNodeOutputImageCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		err = ig.SetNodeOutputImage(
			command.NodeID,
			command.OutputName,
			command.ImageID,
		)

		if err != nil {
			return fmt.Errorf("could not process SetImageGraphNodeOutputImageCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		return nil
	})
}

func (h *ImageGraphCommandHandlers) HandleUnsetImageGraphNodeOutputImageCommand(
	ctx context.Context,
	command *UnsetImageGraphNodeOutputImageCommand,
) (
	[]dorky.Event,
	error,
) {
	return h.uow.Run(ctx, func(repos *Repos) error {
		ig, err := repos.ImageGraphRepository.Get(command.ImageGraphID)

		if err != nil {
			return fmt.Errorf("could not process UnsetImageGraphNodeOutputImageCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		err = ig.UnsetNodeOutputImage(
			command.NodeID,
			command.OutputName,
		)

		if err != nil {
			return fmt.Errorf("could not process UnsetImageGraphNodeOutputImageCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		return nil
	})
}

func (h *ImageGraphCommandHandlers) HandleSetImageGraphNodePreviewCommand(
	ctx context.Context,
	command *SetImageGraphNodePreviewCommand,
) (
	[]dorky.Event,
	error,
) {
	return h.uow.Run(ctx, func(repos *Repos) error {
		ig, err := repos.ImageGraphRepository.Get(command.ImageGraphID)

		if err != nil {
			return fmt.Errorf("could not process SetImageGraphNodePreviewCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		err = ig.SetNodePreview(
			command.NodeID,
			command.ImageID,
		)

		if err != nil {
			return fmt.Errorf("could not process SetImageGraphNodePreviewCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		return nil
	})
}

func (h *ImageGraphCommandHandlers) HandleUnsetImageGraphNodePreviewCommand(
	ctx context.Context,
	command *UnsetImageGraphNodePreviewCommand,
) (
	[]dorky.Event,
	error,
) {
	return h.uow.Run(ctx, func(repos *Repos) error {
		ig, err := repos.ImageGraphRepository.Get(command.ImageGraphID)

		if err != nil {
			return fmt.Errorf("could not process UnsetImageGraphNodePreviewCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		err = ig.UnsetNodePreview(command.NodeID)

		if err != nil {
			return fmt.Errorf("could not process UnsetImageGraphNodePreviewCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		return nil
	})
}
