package application

import (
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky"
)

type CreateImageGraphCommand struct {
	dorky.BaseCommand
	ImageGraphID imagegraph.ImageGraphID
	Name         string
}

func NewCreateImageGraphCommand(
	imageGraphID imagegraph.ImageGraphID,
	name string,
) *CreateImageGraphCommand {
	command := &CreateImageGraphCommand{
		ImageGraphID: imageGraphID,
		Name:         name,
	}
	command.Init("CreateImageGraphCommand")
	return command
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
) *AddImageGraphNodeCommand {
	command := &AddImageGraphNodeCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
		NodeType:     nodeType,
		Name:         name,
		Config:       config,
	}
	command.Init("AddImageGraphNodeCommand")
	return command
}

type RemoveImageGraphNodeCommand struct {
	dorky.BaseCommand
	ImageGraphID imagegraph.ImageGraphID
	NodeID       imagegraph.NodeID
}

func NewRemoveImageGraphNodeCommand(
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
) *RemoveImageGraphNodeCommand {
	command := &RemoveImageGraphNodeCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
	}
	command.Init("RemoveImageGraphNodeCommand")
	return command
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
) *ConnectImageGraphNodesCommand {
	command := &ConnectImageGraphNodesCommand{
		ImageGraphID: imageGraphID,
		FromNodeID:   fromNodeID,
		OutputName:   outputName,
		ToNodeID:     toNodeID,
		InputName:    inputName,
	}
	command.Init("ConnectImageGraphNodesCommand")
	return command
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
) *DisconnectImageGraphNodesCommand {
	command := &DisconnectImageGraphNodesCommand{
		ImageGraphID: imageGraphID,
		FromNodeID:   fromNodeID,
		OutputName:   outputName,
		ToNodeID:     toNodeID,
		InputName:    inputName,
	}
	command.Init("DisconnectImageGraphNodesCommand")
	return command
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
) *SetImageGraphNodeOutputImageCommand {
	command := &SetImageGraphNodeOutputImageCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
		OutputName:   outputName,
		ImageID:      imageID,
	}
	command.Init("SetImageGraphNodeOutputImageCommand")
	return command
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
) *UnsetImageGraphNodeOutputImageCommand {
	command := &UnsetImageGraphNodeOutputImageCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
		OutputName:   outputName,
	}
	command.Init("UnsetImageGraphNodeOutputImageCommand")
	return command
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
) *SetImageGraphNodePreviewCommand {
	command := &SetImageGraphNodePreviewCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
		ImageID:      imageID,
	}
	command.Init("SetImageGraphNodePreviewCommand")
	return command
}

type UnsetImageGraphNodePreviewCommand struct {
	dorky.BaseCommand
	ImageGraphID imagegraph.ImageGraphID
	NodeID       imagegraph.NodeID
}

func NewUnsetImageGraphNodePreviewCommand(
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
) *UnsetImageGraphNodePreviewCommand {
	command := &UnsetImageGraphNodePreviewCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
	}
	command.Init("UnsetImageGraphNodePreviewCommand")
	return command
}

type SetImageGraphNodeConfigCommand struct {
	dorky.BaseCommand
	ImageGraphID imagegraph.ImageGraphID
	NodeID       imagegraph.NodeID
	Config       string
}

func NewSetImageGraphNodeConfigCommand(
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	config string,
) *SetImageGraphNodeConfigCommand {
	command := &SetImageGraphNodeConfigCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
		Config:       config,
	}
	command.Init("SetImageGraphNodeConfigCommand")
	return command
}
