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
	UIMetadataViews *UIMetadataViews
}

func NewUnitOfWork() (*UnitOfWork, error) {
	imageGraphRepository, err := NewImageGraphRepository()

	if err != nil {
		return nil, fmt.Errorf("failed to create ImageGraph repository: %w", err)
	}

	uiMetadataRepository := NewUIMetadataRepository()

	repos := &application.Repos{
		ImageGraphRepository: imageGraphRepository,
		UIMetadataRepository: uiMetadataRepository,
	}

	uow := &UnitOfWork{
		UnitOfWork: inmem.NewUnitOfWork(
			repos,
			imageGraphRepository,
		),
		ImageGraphViews: NewImageGraphViews(imageGraphRepository),
		UIMetadataViews: NewUIMetadataViews(uiMetadataRepository),
	}

	return uow, nil
}
