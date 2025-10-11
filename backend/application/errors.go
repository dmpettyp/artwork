package application

import "errors"

// ErrImageGraphNotFound is returned when an ImageGraph cannot be found
var ErrImageGraphNotFound = errors.New("image graph not found")

// ErrUIMetadataNotFound is returned when UI Metadata cannot be found
var ErrUIMetadataNotFound = errors.New("ui metadata not found")
