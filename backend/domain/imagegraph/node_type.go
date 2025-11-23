package imagegraph

import (
	"encoding/json"
)

type NodeType int

const (
	NodeTypeNone NodeType = iota
	NodeTypeInput
	NodeTypeBlur
	NodeTypeCrop
	NodeTypeOutput
	NodeTypeResize
	NodeTypeResizeMatch
	NodeTypePixelInflate
	NodeTypePaletteExtract
	NodeTypePaletteApply
)

func (nt NodeType) MarshalJSON() ([]byte, error) {
	str := NodeTypeMapper.FromWithDefault(nt, "unknown")
	return json.Marshal(str)
}

// NodeTypeConfig represents the configuration for a node type
type NodeTypeConfig struct {
	Inputs       []InputName
	Outputs      []OutputName
	NameRequired bool
	NewConfig    func() NodeConfig
}

// NodeTypeConfigs maps node types to their configurations
var NodeTypeConfigs = map[NodeType]NodeTypeConfig{
	NodeTypeInput: {
		Outputs:   []OutputName{"original"},
		NewConfig: func() NodeConfig { return NewNodeConfigInput() },
	},
	NodeTypeOutput: {
		Inputs:       []InputName{"input"},
		Outputs:      []OutputName{"final"},
		NameRequired: true,
		NewConfig:    func() NodeConfig { return NewNodeConfigOutput() },
	},
	NodeTypeCrop: {
		Inputs:    []InputName{"original"},
		Outputs:   []OutputName{"cropped"},
		NewConfig: func() NodeConfig { return NewNodeConfigCrop() },
	},
	NodeTypeBlur: {
		Inputs:    []InputName{"original"},
		Outputs:   []OutputName{"blurred"},
		NewConfig: func() NodeConfig { return NewNodeConfigBlur() },
	},
	NodeTypeResize: {
		Inputs:    []InputName{"original"},
		Outputs:   []OutputName{"resized"},
		NewConfig: func() NodeConfig { return NewNodeConfigResize() },
	},
	NodeTypeResizeMatch: {
		Inputs:    []InputName{"original", "size_match"},
		Outputs:   []OutputName{"resized"},
		NewConfig: func() NodeConfig { return NewNodeConfigResizeMatch() },
	},
	NodeTypePixelInflate: {
		Inputs:    []InputName{"original"},
		Outputs:   []OutputName{"inflated"},
		NewConfig: func() NodeConfig { return NewNodeConfigPixelInflate() },
	},
	NodeTypePaletteExtract: {
		Inputs:    []InputName{"source"},
		Outputs:   []OutputName{"palette"},
		NewConfig: func() NodeConfig { return NewNodeConfigPaletteExtract() },
	},
	NodeTypePaletteApply: {
		Inputs:    []InputName{"source", "palette"},
		Outputs:   []OutputName{"mapped"},
		NewConfig: func() NodeConfig { return NewNodeConfigPaletteApply() },
	},
}
