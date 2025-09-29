package imagegraph

import "fmt"

type InputName string

type InputConnection struct {
	NodeID     NodeID
	OutputName OutputName
}

type Input struct {
	Name            InputName
	ImageID         ImageID
	Connected       bool
	InputConnection InputConnection
}

func MakeInput(name InputName) Input {
	return Input{
		Name: name,
	}
}

func (i *Input) Connect(nodeID NodeID, outputName OutputName) error {
	if i.Connected {
		return fmt.Errorf("input %q was already connected", i.Name)
	}

	i.Connected = true
	i.InputConnection.NodeID = nodeID
	i.InputConnection.OutputName = outputName

	return nil
}

func (i *Input) Disconnect() error {
	if !i.Connected {
		return fmt.Errorf("input %q is not connected", i.Name)
	}

	i.Connected = false
	i.InputConnection = InputConnection{}

	return nil
}

func (i *Input) SetImage(imageID ImageID) {
	i.ImageID = imageID
}

func (i *Input) ResetImage() {
	i.ImageID = ImageID{}
}

func (i *Input) HasImage() bool {
	return !i.ImageID.IsNil()
}
