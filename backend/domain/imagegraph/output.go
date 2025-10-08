package imagegraph

import "fmt"

type OutputName string

type OutputConnection struct {
	NodeID    NodeID
	InputName InputName
}

type Output struct {
	Name        OutputName
	ImageID     ImageID
	Connections map[OutputConnection]struct{}
}

func MakeOutput(name OutputName) Output {
	return Output{
		Name:        name,
		Connections: make(map[OutputConnection]struct{}),
	}
}

func (o *Output) Connect(nodeID NodeID, inputName InputName) error {
	oc := OutputConnection{
		NodeID:    nodeID,
		InputName: inputName,
	}

	_, ok := o.Connections[oc]

	if ok {
		return fmt.Errorf(
			"cannot connect output %q to node %q input %q: already connected",
			o.Name, nodeID, inputName,
		)
	}

	o.Connections[oc] = struct{}{}

	return nil
}

func (o *Output) IsConnected(nodeID NodeID, inputName InputName) bool {
	oc := OutputConnection{
		NodeID:    nodeID,
		InputName: inputName,
	}

	_, ok := o.Connections[oc]

	return ok
}

func (o *Output) Disconnect(nodeID NodeID, inputName InputName) error {
	oc := OutputConnection{
		NodeID:    nodeID,
		InputName: inputName,
	}

	_, ok := o.Connections[oc]

	if !ok {
		return fmt.Errorf(
			"cannot disconnect output %q from node %q input %q: not connected",
			o.Name, nodeID, inputName,
		)
	}

	delete(o.Connections, oc)

	return nil
}

func (o *Output) SetImage(imageID ImageID) {
	o.ImageID = imageID
}

func (o *Output) ResetImage() {
	o.ImageID = ImageID{}
}

func (o *Output) HasImage() bool {
	return !o.ImageID.IsNil()
}
