package application

import "errors"

// ErrImageGraphNotFound is returned when an ImageGraph cannot be found
var ErrImageGraphNotFound = errors.New("image graph not found")

// ErrLayoutNotFound is returned when Layout cannot be found
var ErrLayoutNotFound = errors.New("layout not found")

// ErrViewportNotFound is returned when Viewport cannot be found
var ErrViewportNotFound = errors.New("viewport not found")
