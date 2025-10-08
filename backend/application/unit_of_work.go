package application

import (
	"context"

	"github.com/dmpettyp/dorky"
)

type UnitOfWork interface {
	Run(context.Context, func(repos *Repos) error) ([]dorky.Event, error)
}
