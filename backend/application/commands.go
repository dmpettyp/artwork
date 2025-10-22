package application

import (
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
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
	Config       imagegraph.NodeConfig
}

func NewAddImageGraphNodeCommand(
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	nodeType imagegraph.NodeType,
	name string,
	config imagegraph.NodeConfig,
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
	Config       imagegraph.NodeConfig
}

func NewSetImageGraphNodeConfigCommand(
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	config imagegraph.NodeConfig,
) *SetImageGraphNodeConfigCommand {
	command := &SetImageGraphNodeConfigCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
		Config:       config,
	}
	command.Init("SetImageGraphNodeConfigCommand")
	return command
}

type SetImageGraphNodeNameCommand struct {
	dorky.BaseCommand
	ImageGraphID imagegraph.ImageGraphID
	NodeID       imagegraph.NodeID
	Name         string
}

func NewSetImageGraphNodeNameCommand(
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	name string,
) *SetImageGraphNodeNameCommand {
	command := &SetImageGraphNodeNameCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
		Name:         name,
	}
	command.Init("SetImageGraphNodeNameCommand")
	return command
}

// Layout Commands

type UpdateLayoutCommand struct {
	dorky.BaseCommand
	GraphID       imagegraph.ImageGraphID
	NodePositions []ui.NodePosition
}

func NewUpdateLayoutCommand(
	graphID imagegraph.ImageGraphID,
	nodePositions []ui.NodePosition,
) *UpdateLayoutCommand {
	command := &UpdateLayoutCommand{
		GraphID:       graphID,
		NodePositions: nodePositions,
	}
	command.Init("UpdateLayoutCommand")
	return command
}

// Viewport Commands

type UpdateViewportCommand struct {
	dorky.BaseCommand
	GraphID imagegraph.ImageGraphID
	Zoom    float64
	PanX    float64
	PanY    float64
}

func NewUpdateViewportCommand(
	graphID imagegraph.ImageGraphID,
	zoom, panX, panY float64,
) *UpdateViewportCommand {
	command := &UpdateViewportCommand{
		GraphID: graphID,
		Zoom:    zoom,
		PanX:    panX,
		PanY:    panY,
	}
	command.Init("UpdateViewportCommand")
	return command
}
