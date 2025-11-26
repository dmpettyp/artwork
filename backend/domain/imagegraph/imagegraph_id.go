package imagegraph

import "github.com/dmpettyp/dorky/id"

type ImageGraphID struct{ id.ID }

var NewImageGraphID, MustNewImageGraphID, ParseImageGraphID = id.Create(
	func(id id.ID) ImageGraphID { return ImageGraphID{ID: id} },
)
