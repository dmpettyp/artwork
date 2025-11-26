package application

import (
	"context"

	"github.com/dmpettyp/dorky/messages"
)

type UnitOfWork interface {
	Run(context.Context, func(repos *Repos) error) ([]messages.Event, error)
}
