package imagegraph

type OutputName string

type Output struct {
	Name    OutputName
	ImageID ImageID
}

func (o *Output) SetImage(imageID ImageID) {
	o.ImageID = imageID
}

func (o *Output) ResetImage() {
	o.ImageID = ImageID{}
}
