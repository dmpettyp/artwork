package inmem

import (
	"context"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
)

type UIMetadataViews struct {
	repo *UIMetadataRepository
}

func NewUIMetadataViews(repo *UIMetadataRepository) *UIMetadataViews {
	return &UIMetadataViews{repo: repo}
}

func (view *UIMetadataViews) Get(
	_ context.Context,
	graphID imagegraph.ImageGraphID,
) (
	*ui.UIMetadata,
	error,
) {
	result, err := view.repo.Get(graphID)
	if err != nil {
		return nil, err
	}
	return result.Clone(), nil
}
