package postgres

import (
	"database/sql"
	"errors"

	"github.com/dmpettyp/artwork/application"
)

// wrapImageGraphNotFound wraps sql.ErrNoRows as application.ErrImageGraphNotFound
func wrapImageGraphNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return application.ErrImageGraphNotFound
	}
	return err
}

// wrapLayoutNotFound wraps sql.ErrNoRows as application.ErrLayoutNotFound
func wrapLayoutNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return application.ErrLayoutNotFound
	}
	return err
}

// wrapViewportNotFound wraps sql.ErrNoRows as application.ErrViewportNotFound
func wrapViewportNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return application.ErrViewportNotFound
	}
	return err
}
