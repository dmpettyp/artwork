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

type Inputs map[InputName]*Input

func NewInputs(inputNames []InputName) (Inputs, error) {
	inputs := Inputs(make(map[InputName]*Input))

	for _, inputName := range inputNames {
		if err := inputs.Add(inputName); err != nil {
			return nil, fmt.Errorf("could not create inputs: %w", err)
		}
	}

	return inputs, nil
}

func (inputs Inputs) Add(name InputName) error {
	if _, ok := inputs[name]; ok {
		return fmt.Errorf("input named %q already exists", name)
	}
	input := MakeInput(name)
	inputs[name] = &input

	return nil
}

func (inputs Inputs) Exists(name InputName) bool {
	_, ok := inputs[name]
	return ok
}

func (inputs Inputs) Get(name InputName) (*Input, error) {
	input, ok := inputs[name]

	if !ok {
		return nil, fmt.Errorf("input %q does not exist", name)
	}

	return input, nil
}

func (inputs Inputs) ConnectFrom(inputName InputName, nodeID NodeID, outputName OutputName) error {
	input, ok := inputs[inputName]

	if !ok {
		return fmt.Errorf("input %q does not exist", inputName)
	}

	return input.Connect(nodeID, outputName)
}

func (inputs Inputs) IsConnected(inputName InputName) (bool, error) {
	input, ok := inputs[inputName]

	if !ok {
		return false, fmt.Errorf("input %q does not exist", inputName)
	}

	return input.Connected, nil
}

func (inputs Inputs) Disconnect(inputName InputName) (InputConnection, bool, error) {
	input, err := inputs.Get(inputName)

	if err != nil {
		return InputConnection{}, false, fmt.Errorf("could not disconnect input: %w", err)
	}

	// Store the connection before disconnecting
	oldConnection := input.InputConnection

	// Disconnect the input
	if err := input.Disconnect(); err != nil {
		return InputConnection{}, false, err
	}

	// Check if input had an image, reset it if so
	hadImage := input.HasImage()
	if hadImage {
		input.ResetImage()
	}

	return oldConnection, hadImage, nil
}

func (inputs Inputs) SetImage(inputName InputName, imageID ImageID) error {
	if imageID.IsNil() {
		return fmt.Errorf("cannot set input %q image to nil", inputName)
	}

	input, ok := inputs[inputName]

	if !ok {
		return fmt.Errorf("input %q does not exist", inputName)
	}

	input.SetImage(imageID)

	return nil
}

func (inputs Inputs) UnsetImage(inputName InputName) error {
	input, ok := inputs[inputName]

	if !ok {
		return fmt.Errorf("input %q does not exist", inputName)
	}

	input.ResetImage()

	return nil
}

func (inputs Inputs) AllSet() bool {
	for _, input := range inputs {
		if !input.Connected {
			return false
		}

		if input.ImageID.IsNil() {
			return false
		}
	}

	return true
}
