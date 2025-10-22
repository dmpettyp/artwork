package inmem

import (
	"errors"
	"fmt"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
	"github.com/dmpettyp/dorky/inmem"
)

type ViewportRepository struct {
	inmem.Repository[*ui.Viewport]
}

func NewViewportRepository() (*ViewportRepository, error) {
	identityEqualFn := func(a *ui.Viewport, b *ui.Viewport) bool {
		return a.GraphID == b.GraphID
	}

	inmemRepository, err := inmem.CreateRepository(
		identityEqualFn,
		identityEqualFn,
	)

	if err != nil {
		return nil, fmt.Errorf("could not create inmem Viewport repository: %w", err)
	}

	repo := &ViewportRepository{inmemRepository}

	return repo, nil
}

func (repo *ViewportRepository) Get(
	graphID imagegraph.ImageGraphID,
) (
	*ui.Viewport,
	error,
) {
	result, err := repo.FindOne(
		func(a *ui.Viewport) bool { return a.GraphID == graphID },
	)
	if err != nil {
		if errors.Is(err, inmem.ErrNotFound) {
			return nil, application.ErrViewportNotFound
		}
		return nil, err
	}
	return result, nil
}
