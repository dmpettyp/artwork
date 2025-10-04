package application

import "github.com/dmpettyp/artwork/domain/imagegraph"

type Repos struct {
	ImageGraphRepository ImageGraphRepository
}

type ImageGraphRepository interface {
	Add(*imagegraph.ImageGraph) error
	Get(imagegraph.ImageGraphID) (*imagegraph.ImageGraph, error)
}
