package postgres

import (
	"database/sql"
	"errors"

	"github.com/dmpettyp/artwork/backend/application"
)

// isNotFoundError checks if the error is a sql.ErrNoRows and converts it
// to the appropriate application-layer error
func isNotFoundError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return err // Will be wrapped by caller into specific NotFound error
	}
	return err
}

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
