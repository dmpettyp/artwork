package inmem

import (
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky/inmem"
)

type ImageGraphRepository struct {
	inmem.Repository[*imagegraph.ImageGraph]
}

func NewImageGraphRepository() (*ImageGraphRepository, error) {
	identityEqualFn := func(a *imagegraph.ImageGraph, b *imagegraph.ImageGraph) bool {
		return a.ID == b.ID
	}

	inmemRepository, err := inmem.CreateRepository(
		identityEqualFn,
		identityEqualFn,
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
	*imagegraph.ImageGraph,
	error,
) {
	return repo.FindOne(
		func(a *imagegraph.ImageGraph) bool { return a.ID == id },
	)
}
