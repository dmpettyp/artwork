package application

import (
	"context"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
)

type ImageGraphViews interface {
	Get(
		ctx context.Context,
		id imagegraph.ImageGraphID,
	) (
		*imagegraph.ImageGraph,
		error,
	)

	List(ctx context.Context) (
		[]*imagegraph.ImageGraph,
		error,
	)
}

type LayoutViews interface {
	Get(
		ctx context.Context,
		graphID imagegraph.ImageGraphID,
	) (
		*ui.Layout,
		error,
	)
}

type ViewportViews interface {
	Get(
		ctx context.Context,
		graphID imagegraph.ImageGraphID,
	) (
		*ui.Viewport,
		error,
	)
}
