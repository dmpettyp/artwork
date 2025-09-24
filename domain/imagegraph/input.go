package imagegraph

type InputName string

type Input struct {
	Name             InputName
	ImageID          ImageID
	SourceNodeID     NodeID
	SourceOutputName OutputName
}

func (i *Input) Connect(nodeID NodeID, outputName OutputName) {
	i.SourceNodeID = nodeID
	i.SourceOutputName = outputName
}

func (i *Input) Disconnect() {
	i.SourceNodeID = NodeID{}
	i.SourceOutputName = ""
}

func (i *Input) SetImage(imageID ImageID) {
	i.ImageID = imageID
}

func (i *Input) ResetImage() {
	i.ImageID = ImageID{}
}

func (i *Input) IsConnected() bool {
	return !i.SourceNodeID.IsNil()
}

func (i *Input) HasImage() bool {
	return !i.ImageID.IsNil()
}
