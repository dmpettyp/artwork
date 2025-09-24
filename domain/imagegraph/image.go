package imagegraph

import "github.com/dmpettyp/id"

type ImageID struct{ id.ID }

var NewImageID, MustNewImageID, ParseImageID = id.Intitalizers(
	func(id id.ID) ImageID { return ImageID{ID: id} },
)
