package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky/messagebus"
	"github.com/dmpettyp/dorky/messages"
)

type ImageGraphCommandHandlers struct {
	uow UnitOfWork
}

// NewImageGraphCommandHandlers initializes the handlers struct that processes
// all ImageGraph Commands and registers all handlers with the provided
// message bus
func NewImageGraphCommandHandlers(
	mb *messagebus.MessageBus,
	uow UnitOfWork,
) (
	*ImageGraphCommandHandlers,
	error,
) {
	handlers := &ImageGraphCommandHandlers{uow: uow}

	err := errors.Join(
		messagebus.RegisterCommandHandler(mb, handlers.HandleCreateImageGraphCommand),
		messagebus.RegisterCommandHandler(mb, handlers.HandleAddImageGraphNodeCommand),
		messagebus.RegisterCommandHandler(mb, handlers.HandleRemoveImageGraphNodeCommand),
		messagebus.RegisterCommandHandler(mb, handlers.HandleConnectImageGraphNodesCommand),
		messagebus.RegisterCommandHandler(mb, handlers.HandleDisconnectImageGraphNodesCommand),
		messagebus.RegisterCommandHandler(mb, handlers.HandleSetImageGraphNodeOutputImageCommand),
		messagebus.RegisterCommandHandler(mb, handlers.HandleUnsetImageGraphNodeOutputImageCommand),
		messagebus.RegisterCommandHandler(mb, handlers.HandleSetImageGraphNodePreviewCommand),
		messagebus.RegisterCommandHandler(mb, handlers.HandleUnsetImageGraphNodePreviewCommand),
		messagebus.RegisterCommandHandler(mb, handlers.HandleSetImageGraphNodeConfigCommand),
		messagebus.RegisterCommandHandler(mb, handlers.HandleSetImageGraphNodeNameCommand),
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
	[]messages.Event,
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
	[]messages.Event,
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
		)

		if err != nil {
			return fmt.Errorf("could not process AddImageGraphNodeCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		if command.Config != nil {
			err = ig.SetNodeConfig(command.NodeID, command.Config)
			if err != nil {
				return fmt.Errorf("could not process AddImageGraphNodeCommand for ImageGraph %q: %w", command.ImageGraphID, err)
			}
		}

		return nil
	})
}

func (h *ImageGraphCommandHandlers) HandleRemoveImageGraphNodeCommand(
	ctx context.Context,
	command *RemoveImageGraphNodeCommand,
) (
	[]messages.Event,
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
	[]messages.Event,
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
	[]messages.Event,
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
	[]messages.Event,
	error,
) {
	return h.uow.Run(ctx, func(repos *Repos) error {
		ig, err := repos.ImageGraphRepository.Get(command.ImageGraphID)

		if err != nil {
			return fmt.Errorf("could not process SetImageGraphNodeOutputImageCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		nodeVersion := command.NodeVersion
		if nodeVersion == 0 {
			node, ok := ig.Nodes.Get(command.NodeID)
			if !ok {
				return fmt.Errorf("could not process SetImageGraphNodeOutputImageCommand for ImageGraph %q: node %q not found", command.ImageGraphID, command.NodeID)
			}
			nodeVersion = node.Version
		}

		err = ig.SetNodeOutputImage(
			command.NodeID,
			command.OutputName,
			command.ImageID,
			nodeVersion,
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
	[]messages.Event,
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
	[]messages.Event,
	error,
) {
	return h.uow.Run(ctx, func(repos *Repos) error {
		ig, err := repos.ImageGraphRepository.Get(command.ImageGraphID)

		if err != nil {
			return fmt.Errorf("could not process SetImageGraphNodePreviewCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		nodeVersion := command.NodeVersion
		if nodeVersion == 0 {
			node, ok := ig.Nodes.Get(command.NodeID)
			if !ok {
				return fmt.Errorf("could not process SetImageGraphNodePreviewCommand for ImageGraph %q: node %q not found", command.ImageGraphID, command.NodeID)
			}
			nodeVersion = node.Version
		}

		err = ig.SetNodePreview(
			command.NodeID,
			command.ImageID,
			nodeVersion,
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
	[]messages.Event,
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

func (h *ImageGraphCommandHandlers) HandleSetImageGraphNodeConfigCommand(
	ctx context.Context,
	command *SetImageGraphNodeConfigCommand,
) (
	[]messages.Event,
	error,
) {
	return h.uow.Run(ctx, func(repos *Repos) error {
		ig, err := repos.ImageGraphRepository.Get(command.ImageGraphID)

		if err != nil {
			return fmt.Errorf("could not process SetImageGraphNodeConfigCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		if command.Config != nil {
			err = ig.SetNodeConfig(command.NodeID, command.Config)
			if err != nil {
				return fmt.Errorf("could not process SetImageGraphNodeConfigCommand for ImageGraph %q: %w", command.ImageGraphID, err)
			}
		}

		return nil
	})
}

func (h *ImageGraphCommandHandlers) HandleSetImageGraphNodeNameCommand(
	ctx context.Context,
	command *SetImageGraphNodeNameCommand,
) (
	[]messages.Event,
	error,
) {
	return h.uow.Run(ctx, func(repos *Repos) error {
		ig, err := repos.ImageGraphRepository.Get(command.ImageGraphID)

		if err != nil {
			return fmt.Errorf("could not process SetImageGraphNodeNameCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		err = ig.SetNodeName(command.NodeID, command.Name)

		if err != nil {
			return fmt.Errorf("could not process SetImageGraphNodeNameCommand for ImageGraph %q: %w", command.ImageGraphID, err)
		}

		return nil
	})
}
