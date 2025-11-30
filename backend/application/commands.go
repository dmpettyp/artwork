package application

import (
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
	"github.com/dmpettyp/dorky/messages"
)

type CreateImageGraphCommand struct {
	messages.BaseCommand
	ImageGraphID imagegraph.ImageGraphID `json:"image_graph_id"`
	Name         string                  `json:"name"`
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
	messages.BaseCommand
	ImageGraphID imagegraph.ImageGraphID `json:"image_graph_id"`
	NodeID       imagegraph.NodeID       `json:"node_id"`
	NodeType     imagegraph.NodeType     `json:"node_type"`
	Name         string                  `json:"name"`
	Config       imagegraph.NodeConfig   `json:"config"`
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
	messages.BaseCommand
	ImageGraphID imagegraph.ImageGraphID `json:"image_graph_id"`
	NodeID       imagegraph.NodeID       `json:"node_id"`
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
	messages.BaseCommand
	ImageGraphID imagegraph.ImageGraphID `json:"image_graph_id"`
	FromNodeID   imagegraph.NodeID       `json:"from_node_id"`
	OutputName   imagegraph.OutputName   `json:"output_name"`
	ToNodeID     imagegraph.NodeID       `json:"to_node_id"`
	InputName    imagegraph.InputName    `json:"input_name"`
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
	messages.BaseCommand
	ImageGraphID imagegraph.ImageGraphID `json:"image_graph_id"`
	FromNodeID   imagegraph.NodeID       `json:"from_node_id"`
	OutputName   imagegraph.OutputName   `json:"output_name"`
	ToNodeID     imagegraph.NodeID       `json:"to_node_id"`
	InputName    imagegraph.InputName    `json:"input_name"`
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
	messages.BaseCommand
	ImageGraphID imagegraph.ImageGraphID `json:"image_graph_id"`
	NodeID       imagegraph.NodeID       `json:"node_id"`
	OutputName   imagegraph.OutputName   `json:"output_name"`
	ImageID      imagegraph.ImageID      `json:"image_id"`
	NodeVersion  imagegraph.NodeVersion  `json:"node_version"`
}

func NewSetImageGraphNodeOutputImageCommand(
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	outputName imagegraph.OutputName,
	imageID imagegraph.ImageID,
	nodeVersion imagegraph.NodeVersion,
) *SetImageGraphNodeOutputImageCommand {
	command := &SetImageGraphNodeOutputImageCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
		OutputName:   outputName,
		ImageID:      imageID,
		NodeVersion:  nodeVersion,
	}
	command.Init("SetImageGraphNodeOutputImageCommand")
	return command
}

type UnsetImageGraphNodeOutputImageCommand struct {
	messages.BaseCommand
	ImageGraphID imagegraph.ImageGraphID `json:"image_graph_id"`
	NodeID       imagegraph.NodeID       `json:"node_id"`
	OutputName   imagegraph.OutputName   `json:"output_name"`
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
	messages.BaseCommand
	ImageGraphID imagegraph.ImageGraphID `json:"image_graph_id"`
	NodeID       imagegraph.NodeID       `json:"node_id"`
	ImageID      imagegraph.ImageID      `json:"image_id"`
	NodeVersion  imagegraph.NodeVersion  `json:"node_version"`
}

func NewSetImageGraphNodePreviewCommand(
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	imageID imagegraph.ImageID,
	nodeVersion imagegraph.NodeVersion,
) *SetImageGraphNodePreviewCommand {
	command := &SetImageGraphNodePreviewCommand{
		ImageGraphID: imageGraphID,
		NodeID:       nodeID,
		ImageID:      imageID,
		NodeVersion:  nodeVersion,
	}
	command.Init("SetImageGraphNodePreviewCommand")
	return command
}

type UnsetImageGraphNodePreviewCommand struct {
	messages.BaseCommand
	ImageGraphID imagegraph.ImageGraphID `json:"image_graph_id"`
	NodeID       imagegraph.NodeID       `json:"node_id"`
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
	messages.BaseCommand
	ImageGraphID imagegraph.ImageGraphID `json:"image_graph_id"`
	NodeID       imagegraph.NodeID       `json:"node_id"`
	Config       imagegraph.NodeConfig   `json:"config"`
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
	messages.BaseCommand
	ImageGraphID imagegraph.ImageGraphID `json:"image_graph_id"`
	NodeID       imagegraph.NodeID       `json:"node_id"`
	Name         string                  `json:"name"`
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
	messages.BaseCommand
	GraphID       imagegraph.ImageGraphID `json:"graph_id"`
	NodePositions []ui.NodePosition       `json:"node_positions"`
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
	messages.BaseCommand
	GraphID imagegraph.ImageGraphID `json:"graph_id"`
	Zoom    float64                 `json:"zoom"`
	PanX    float64                 `json:"pan_x"`
	PanY    float64                 `json:"pan_y"`
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
