package imagegen

import (
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky"
)

type imageStorage interface {
	Save(imageID imagegraph.ImageID, imageData []byte) error
	Get(imageID imagegraph.ImageID) ([]byte, error)
}

type ImageGen struct {
	imageStorage imageStorage
	messageBus   *dorky.MessageBus
}

func NewImageGen(
	imageStorage imageStorage,
	messageBus *dorky.MessageBus,
) *ImageGen {
	return &ImageGen{
		imageStorage: imageStorage,
		messageBus:   messageBus,
	}
}
