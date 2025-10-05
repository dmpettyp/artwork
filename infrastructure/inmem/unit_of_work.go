package inmem

import (
	"fmt"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/dorky/inmem"
)

// UnitOfWork is an in-memory version of the service's UnitOfWork
// that uses lib.ddd's InmemUnitOfWorkDeprecated to drive the uow lifecycle
type UnitOfWork struct {
	*inmem.UnitOfWork[*application.Repos]
	// showingViews          *ShowingViews
}

func NewUnitOfWork() (*UnitOfWork, error) {
	imageGraphRepository, err := NewImageGraphRepository()

	if err != nil {
		return nil, fmt.Errorf("failed to create tour request repository: %w", err)
	}

	repos := &application.Repos{
		ImageGraphRepository: imageGraphRepository,
	}

	uow := &UnitOfWork{
		// showingViews:          NewShowingViews(showings),
		UnitOfWork: inmem.NewUnitOfWork(repos, imageGraphRepository),
	}

	return uow, nil
}

// func (uow *UnitOfWork) ShowingViews() (*ShowingViews, error) {
// 	return uow.showingViews, nil
// }
