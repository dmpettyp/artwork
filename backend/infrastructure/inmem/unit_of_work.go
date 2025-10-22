package inmem

import (
	"fmt"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/dorky/inmem"
)

// UnitOfWork is an in-memory version of the service's UnitOfWork
// that uses lib.dorky's inmem.UnitOfWork to drive the uow lifecycle
type UnitOfWork struct {
	*inmem.UnitOfWork[*application.Repos]
	ImageGraphViews *ImageGraphViews
	LayoutViews     *LayoutViews
	ViewportViews   *ViewportViews
}

func NewUnitOfWork() (*UnitOfWork, error) {
	imageGraphRepository, err := NewImageGraphRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to create ImageGraph repository: %w", err)
	}

	layoutRepository, err := NewLayoutRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to create Layout repository: %w", err)
	}

	viewportRepository, err := NewViewportRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to create Viewport repository: %w", err)
	}

	repos := &application.Repos{
		ImageGraphRepository: imageGraphRepository,
		LayoutRepository:     layoutRepository,
		ViewportRepository:   viewportRepository,
	}

	uow := &UnitOfWork{
		UnitOfWork: inmem.NewUnitOfWork(
			repos,
			imageGraphRepository,
			layoutRepository,
			viewportRepository,
		),
		ImageGraphViews: NewImageGraphViews(imageGraphRepository),
		LayoutViews:     NewLayoutViews(layoutRepository),
		ViewportViews:   NewViewportViews(viewportRepository),
	}

	return uow, nil
}
