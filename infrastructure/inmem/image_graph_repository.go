package inmem

import (
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky/inmem"
)

type ImageGraph = imagegraph.ImageGraph
type ImageGraphRepository struct {
	inmem.Repository[*ImageGraph]
}

func NewImageGraphRepository() (*ImageGraphRepository, error) {
	identityEqualFn := func(a *ImageGraph, b *ImageGraph) bool {
		return a.ID == b.ID
	}

	constraintEqualFn := func(a *ImageGraph, b *ImageGraph) bool {
		return a.ID == b.ID
	}

	inmemRepository, err := inmem.CreateRepository(
		identityEqualFn,
		constraintEqualFn,
	)

	if err != nil {
		return nil, fmt.Errorf("could not create inmem ImageGraph repository: %w", err)
	}

	repo := &ImageGraphRepository{inmemRepository}

	return repo, nil
}

func (repo *ImageGraphRepository) Get(
	id imagegraph.ImageGraphID,
) (
	*ImageGraph,
	error,
) {
	return repo.FindOne(
		func(a *ImageGraph) bool { return a.ID == id },
	)
}
