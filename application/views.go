package application

import (
	"context"

	"github.com/dmpettyp/artwork/domain/imagegraph"
)

type ImageGraphViews interface {
	Get(
		ctx context.Context,
		id imagegraph.ImageGraphID,
	) (
		*imagegraph.ImageGraph,
		error,
	)
}
