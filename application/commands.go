package application

import (
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky"
)

type CreateImageGraphCommand struct {
	dorky.BaseCommand
	ImageGraphID imagegraph.ImageGraphID
	Name         string
}

func NewCreateImageGraphCommand(
	name string,
) (*CreateImageGraphCommand, error) {
	imageGraphID, err := imagegraph.NewImageGraphID()

	if err != nil {
		return nil, fmt.Errorf(
			"could not create new CreateImageGraphCommand: %w",
			err,
		)
	}
	command := &CreateImageGraphCommand{
		ImageGraphID: imageGraphID,
		Name:         name,
	}

	err = command.Init("CreateImageGraphCommand")

	if err != nil {
		return nil, fmt.Errorf(
			"could not create new CreateImageGraphCommand: %w",
			err,
		)
	}

	return command, nil
}

type AddImageGraphNodeCommand struct {
	dorky.BaseCommand
	ImageGraphID imagegraph.ImageGraphID
	NodeID       imagegraph.NodeID
	NodeType     imagegraph.NodeType
	Name         string
	Config       string
}

func NewAddImageGraphNodeCommand(
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	nodeType imagegraph.NodeType,
	name string,
	config string,
) (*AddImageGraphNodeCommand, error) {
	command := &AddImageGraphNodeCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
		NodeType:     nodeType,
		Name:         name,
		Config:       config,
	}

	err := command.Init("AddImageGraphNodeCommand")

	if err != nil {
		return nil, fmt.Errorf(
			"could not create new AddImageGraphNodeCommand: %w",
			err,
		)
	}

	return command, nil
}

type RemoveImageGraphNodeCommand struct {
	dorky.BaseCommand
	ImageGraphID imagegraph.ImageGraphID
	NodeID       imagegraph.NodeID
}

func NewRemoveImageGraphNodeCommand(
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
) (*RemoveImageGraphNodeCommand, error) {
	command := &RemoveImageGraphNodeCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
	}

	err := command.Init("RemoveImageGraphNodeCommand")

	if err != nil {
		return nil, fmt.Errorf(
			"could not create new RemoveImageGraphNodeCommand: %w",
			err,
		)
	}

	return command, nil
}

type ConnectImageGraphNodesCommand struct {
	dorky.BaseCommand
	ImageGraphID imagegraph.ImageGraphID
	FromNodeID   imagegraph.NodeID
	OutputName   imagegraph.OutputName
	ToNodeID     imagegraph.NodeID
	InputName    imagegraph.InputName
}

func NewConnectImageGraphNodesCommand(
	imageGraphID imagegraph.ImageGraphID,
	fromNodeID imagegraph.NodeID,
	outputName imagegraph.OutputName,
	toNodeID imagegraph.NodeID,
	inputName imagegraph.InputName,
) (*ConnectImageGraphNodesCommand, error) {
	command := &ConnectImageGraphNodesCommand{
		ImageGraphID: imageGraphID,
		FromNodeID:   fromNodeID,
		OutputName:   outputName,
		ToNodeID:     toNodeID,
		InputName:    inputName,
	}

	err := command.Init("ConnectImageGraphNodesCommand")

	if err != nil {
		return nil, fmt.Errorf(
			"could not create new ConnectImageGraphNodesCommand: %w",
			err,
		)
	}

	return command, nil
}

type DisconnectImageGraphNodesCommand struct {
	dorky.BaseCommand
	ImageGraphID imagegraph.ImageGraphID
	FromNodeID   imagegraph.NodeID
	OutputName   imagegraph.OutputName
	ToNodeID     imagegraph.NodeID
	InputName    imagegraph.InputName
}

func NewDisconnectImageGraphNodesCommand(
	imageGraphID imagegraph.ImageGraphID,
	fromNodeID imagegraph.NodeID,
	outputName imagegraph.OutputName,
	toNodeID imagegraph.NodeID,
	inputName imagegraph.InputName,
) (*DisconnectImageGraphNodesCommand, error) {
	command := &DisconnectImageGraphNodesCommand{
		ImageGraphID: imageGraphID,
		FromNodeID:   fromNodeID,
		OutputName:   outputName,
		ToNodeID:     toNodeID,
		InputName:    inputName,
	}

	err := command.Init("DisconnectImageGraphNodesCommand")

	if err != nil {
		return nil, fmt.Errorf(
			"could not create new DisconnectImageGraphNodesCommand: %w",
			err,
		)
	}

	return command, nil
}

type SetImageGraphNodeOutputImageCommand struct {
	dorky.BaseCommand
	ImageGraphID imagegraph.ImageGraphID
	NodeID       imagegraph.NodeID
	OutputName   imagegraph.OutputName
	ImageID      imagegraph.ImageID
}

func NewSetImageGraphNodeOutputImageCommand(
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	outputName imagegraph.OutputName,
	imageID imagegraph.ImageID,
) (*SetImageGraphNodeOutputImageCommand, error) {
	command := &SetImageGraphNodeOutputImageCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
		OutputName:   outputName,
		ImageID:      imageID,
	}

	err := command.Init("SetImageGraphNodeOutputImageCommand")

	if err != nil {
		return nil, fmt.Errorf(
			"could not create new SetImageGraphNodeOutputImageCommand: %w",
			err,
		)
	}

	return command, nil
}

type UnsetImageGraphNodeOutputImageCommand struct {
	dorky.BaseCommand
	ImageGraphID imagegraph.ImageGraphID
	NodeID       imagegraph.NodeID
	OutputName   imagegraph.OutputName
}

func NewUnsetImageGraphNodeOutputImageCommand(
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	outputName imagegraph.OutputName,
) (*UnsetImageGraphNodeOutputImageCommand, error) {
	command := &UnsetImageGraphNodeOutputImageCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
		OutputName:   outputName,
	}

	err := command.Init("UnsetImageGraphNodeOutputImageCommand")

	if err != nil {
		return nil, fmt.Errorf(
			"could not create new UnsetImageGraphNodeOutputImageCommand: %w",
			err,
		)
	}

	return command, nil
}

type SetImageGraphNodePreviewCommand struct {
	dorky.BaseCommand
	ImageGraphID imagegraph.ImageGraphID
	NodeID       imagegraph.NodeID
	ImageID      imagegraph.ImageID
}

func NewSetImageGraphNodePreviewCommand(
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	imageID imagegraph.ImageID,
) (*SetImageGraphNodePreviewCommand, error) {
	command := &SetImageGraphNodePreviewCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
		ImageID:      imageID,
	}

	err := command.Init("SetImageGraphNodePreviewCommand")

	if err != nil {
		return nil, fmt.Errorf(
			"could not create new SetImageGraphNodePreviewCommand: %w",
			err,
		)
	}

	return command, nil
}

type UnsetImageGraphNodePreviewCommand struct {
	dorky.BaseCommand
	ImageGraphID imagegraph.ImageGraphID
	NodeID       imagegraph.NodeID
}

func NewUnsetImageGraphNodePreviewCommand(
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
) (*UnsetImageGraphNodePreviewCommand, error) {
	command := &UnsetImageGraphNodePreviewCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
	}

	err := command.Init("UnsetImageGraphNodePreviewCommand")

	if err != nil {
		return nil, fmt.Errorf(
			"could not create new UnsetImageGraphNodePreviewCommand: %w",
			err,
		)
	}

	return command, nil
}
