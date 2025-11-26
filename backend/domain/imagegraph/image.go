package imagegraph

import "github.com/dmpettyp/dorky/id"

type ImageID struct{ id.ID }

var NewImageID, MustNewImageID, ParseImageID = id.Create(
	func(id id.ID) ImageID { return ImageID{ID: id} },
)
