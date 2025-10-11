package application

import (
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
)

type Repos struct {
	ImageGraphRepository ImageGraphRepository
	UIMetadataRepository UIMetadataRepository
}

type ImageGraphRepository interface {
	Add(*imagegraph.ImageGraph) error
	Get(imagegraph.ImageGraphID) (*imagegraph.ImageGraph, error)
}

type UIMetadataRepository interface {
	Get(graphID imagegraph.ImageGraphID) (*ui.UIMetadata, error)
	Add(metadata *ui.UIMetadata) error
}
