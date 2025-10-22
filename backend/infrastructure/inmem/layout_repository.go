package inmem

import (
	"errors"
	"fmt"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
	"github.com/dmpettyp/dorky/inmem"
)

type LayoutRepository struct {
	inmem.Repository[*ui.Layout]
}

func NewLayoutRepository() (*LayoutRepository, error) {
	identityEqualFn := func(a *ui.Layout, b *ui.Layout) bool {
		return a.GraphID == b.GraphID
	}

	inmemRepository, err := inmem.CreateRepository(
		identityEqualFn,
		identityEqualFn,
	)

	if err != nil {
		return nil, fmt.Errorf("could not create inmem Layout repository: %w", err)
	}

	repo := &LayoutRepository{inmemRepository}

	return repo, nil
}

func (repo *LayoutRepository) Get(
	graphID imagegraph.ImageGraphID,
) (
	*ui.Layout,
	error,
) {
	result, err := repo.FindOne(
		func(a *ui.Layout) bool { return a.GraphID == graphID },
	)
	if err != nil {
		if errors.Is(err, inmem.ErrNotFound) {
			return nil, application.ErrLayoutNotFound
		}
		return nil, err
	}
	return result, nil
}
