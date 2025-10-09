package imagegraph

import "github.com/dmpettyp/id"

type ImageGraphID struct{ id.ID }

var NewImageGraphID, MustNewImageGraphID, ParseImageGraphID = id.Inititalizers(
	func(id id.ID) ImageGraphID { return ImageGraphID{ID: id} },
)
