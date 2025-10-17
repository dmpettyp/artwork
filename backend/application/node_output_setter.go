package application

import (
	"context"
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky"
)

type NodeOutputSetter struct {
	messageBus *dorky.MessageBus
}

func NewNodeOutputSetter(messageBus *dorky.MessageBus) *NodeOutputSetter {
	return &NodeOutputSetter{
		messageBus: messageBus,
	}
}

func (s *NodeOutputSetter) SetNodeOutputImage(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	outputName imagegraph.OutputName,
	imageID imagegraph.ImageID,
) error {
	cmd := NewSetImageGraphNodeOutputImageCommand(
		imageGraphID,
		nodeID,
		outputName,
		imageID,
	)

	err := s.messageBus.HandleCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("could not handle SetImageGraphNodeOutputImageCommand: %w", err)
	}

	return nil
}
