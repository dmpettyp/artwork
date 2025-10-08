package inmem

import (
	"context"

	"github.com/dmpettyp/artwork/domain/imagegraph"
)

type ImageGraphViews struct {
	repo *ImageGraphRepository
}

func NewImageGraphViews(repo *ImageGraphRepository) *ImageGraphViews {
	return &ImageGraphViews{repo}
}

func (view *ImageGraphViews) Get(
	_ context.Context,
	id imagegraph.ImageGraphID,
) (
	*imagegraph.ImageGraph,
	error,
) {
	result, err := view.repo.Get(id)
	if err != nil {
		return nil, err
	}
	return result.Clone(), nil
}
