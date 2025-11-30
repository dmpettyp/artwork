package application

import (
	"context"
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky/messagebus"
)

type NodeUpdater struct {
	messageBus *messagebus.MessageBus
}

func NewNodeUpdater(messageBus *messagebus.MessageBus) *NodeUpdater {
	return &NodeUpdater{
		messageBus: messageBus,
	}
}

func (s *NodeUpdater) SetNodeOutputImage(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	outputName imagegraph.OutputName,
	imageID imagegraph.ImageID,
	nodeVersion imagegraph.NodeVersion,
) error {
	cmd := NewSetImageGraphNodeOutputImageCommand(
		imageGraphID,
		nodeID,
		outputName,
		imageID,
		nodeVersion,
	)

	err := s.messageBus.HandleCommand(ctx, cmd)

	if err != nil {
		return fmt.Errorf("could not set node output image: %w", err)
	}

	return nil
}

func (s *NodeUpdater) SetNodePreviewImage(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	imageID imagegraph.ImageID,
	nodeVersion imagegraph.NodeVersion,
) error {
	cmd := NewSetImageGraphNodePreviewCommand(
		imageGraphID,
		nodeID,
		imageID,
		nodeVersion,
	)

	err := s.messageBus.HandleCommand(ctx, cmd)

	if err != nil {
		return fmt.Errorf("could not set node preview image: %w", err)
	}

	return nil
}

func (s *NodeUpdater) SetNodeConfig(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	config imagegraph.NodeConfig,
) error {
	cmd := NewSetImageGraphNodeConfigCommand(
		imageGraphID,
		nodeID,
		config,
	)

	if err := s.messageBus.HandleCommand(ctx, cmd); err != nil {
		return fmt.Errorf("could not set node config: %w", err)
	}

	return nil
}
