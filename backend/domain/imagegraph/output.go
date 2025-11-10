package imagegraph

import (
	"fmt"
	"maps"
	"slices"
)

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

type Outputs map[OutputName]*Output

func NewOutputs(outputNames []OutputName) (Outputs, error) {
	outputs := Outputs(make(map[OutputName]*Output))

	for _, outputName := range outputNames {
		if err := outputs.Add(outputName); err != nil {
			return nil, fmt.Errorf("could not create outputs: %w", err)
		}
	}

	return outputs, nil
}

func (outputs Outputs) Add(name OutputName) error {
	if _, ok := outputs[name]; ok {
		return fmt.Errorf("output named %q already exists", name)
	}
	output := MakeOutput(name)
	outputs[name] = &output

	return nil
}

func (outputs Outputs) Each(apply func(output *Output) error) error {
	for _, output := range outputs {
		if err := apply(output); err != nil {
			return err
		}
	}

	return nil
}

func (outputs Outputs) IsOutputConnectedTo(
	outputName OutputName,
	toNodeID NodeID,
	inputName InputName,
) (
	bool,
	error,
) {
	output, ok := outputs[outputName]

	if !ok {
		return false, fmt.Errorf("no output named %q exists", outputName)
	}

	return output.IsConnected(toNodeID, inputName), nil
}

func (outputs Outputs) GetImage(name OutputName) (ImageID, error) {
	output, ok := outputs[name]

	if !ok {
		return ImageID{}, fmt.Errorf("output %q doesn't exist", name)
	}

	return output.ImageID, nil
}

func (outputs Outputs) SetImage(
	outputName OutputName,
	imageID ImageID,
) error {
	if imageID.IsNil() {
		return fmt.Errorf("cannot set output %q to nil", outputName)
	}

	output, ok := outputs[outputName]

	if !ok {
		return fmt.Errorf("no output named %q exists", outputName)
	}

	output.SetImage(imageID)

	return nil
}

func (outputs Outputs) UnsetImage(outputName OutputName) (ImageID, error) {
	output, ok := outputs[outputName]

	if !ok {
		return ImageID{}, fmt.Errorf("no output named %q exists", outputName)
	}

	if output.ImageID.IsNil() {
		return ImageID{}, nil
	}

	oldImageID := output.ImageID

	output.ResetImage()

	return oldImageID, nil
}

func (outputs Outputs) Connections(
	outputName OutputName,
) (
	[]OutputConnection,
	error,
) {
	output, ok := outputs[outputName]

	if !ok {
		return nil, fmt.Errorf("no output named %q exists", outputName)
	}

	return slices.Collect(maps.Keys(output.Connections)), nil
}

func (outputs Outputs) ConnectTo(
	outputName OutputName,
	toNodeID NodeID,
	inputName InputName,
) error {
	output, ok := outputs[outputName]

	if !ok {
		return fmt.Errorf("no output named %q exists", outputName)
	}

	if err := output.Connect(toNodeID, inputName); err != nil {
		return fmt.Errorf("could not connect output: %w", err)
	}

	return nil
}

func (outputs Outputs) DisconnectFrom(
	outputName OutputName,
	toNodeID NodeID,
	inputName InputName,
) error {
	output, ok := outputs[outputName]

	if !ok {
		return fmt.Errorf("no output named %q exists", outputName)
	}

	if err := output.Disconnect(toNodeID, inputName); err != nil {
		return fmt.Errorf("could not disconnect output: %w", err)
	}

	return nil
}

func (outputs Outputs) AllSet() bool {
	for _, output := range outputs {
		if output.ImageID.IsNil() {
			return false
		}
	}

	return true
}
