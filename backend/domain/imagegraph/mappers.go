package imagegraph

import "github.com/dmpettyp/dorky/mapper"

var NodeTypeMapper = mapper.MustNew[string, NodeType](
	"input", NodeTypeInput,
	"output", NodeTypeOutput,
	"crop", NodeTypeCrop,
	"blur", NodeTypeBlur,
	"resize", NodeTypeResize,
	"resize_match", NodeTypeResizeMatch,
	"pixel_inflate", NodeTypePixelInflate,
	"palette_extract", NodeTypePaletteExtract,
	"palette_apply", NodeTypePaletteApply,
	"palette_create", NodeTypePaletteCreate,
	"palette_edit", NodeTypePaletteEdit,
)

var NodeStateMapper = mapper.MustNew[string, NodeState](
	"waiting", Waiting,
	"generating", Generating,
	"generated", Generated,
)
