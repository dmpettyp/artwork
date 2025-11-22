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

func AllNodeTypes() []NodeType {
	return []NodeType{
		NodeTypeInput,
		NodeTypeOutput,
		NodeTypeBlur,
		NodeTypeCrop,
		NodeTypeResize,
		NodeTypeResizeMatch,
		NodeTypePixelInflate,
		NodeTypePaletteExtract,
		NodeTypePaletteApply,
	}
}

type NodeConfigFieldType int

const (
	NodeConfigTypeNone NodeConfigFieldType = iota
	NodeConfigTypeString
	NodeConfigTypeInt
	NodeConfigTypeFloat
	NodeConfigTypeBool
	NodeConfigTypeOption
)

// NodeConfigField represents a single configuration field for a node type
type NodeConfigField struct {
	Name      string
	FieldType NodeConfigFieldType
	Required  bool
	Options   []string
	Default   any // Default value for the field (type depends on FieldType)
}

// NodeTypeConfig represents the configuration for a node type
type NodeTypeConfig struct {
	NodeType     NodeType
	Inputs       []InputName
	Outputs      []OutputName
	NameRequired bool
	Fields       []NodeConfigField
}

// NodeTypeConfigs defines all node type configurations in order
var NodeTypeConfigs = []NodeTypeConfig{
	{
		NodeType: NodeTypeInput,
		Outputs:  []OutputName{"original"},
	},
	{
		NodeType:     NodeTypeOutput,
		Inputs:       []InputName{"input"},
		Outputs:      []OutputName{"final"},
		NameRequired: true,
	},
	{
		NodeType: NodeTypeCrop,
		Inputs:   []InputName{"original"},
		Outputs:  []OutputName{"cropped"},
		Fields: []NodeConfigField{
			{"left", NodeConfigTypeInt, false, nil, nil},
			{"right", NodeConfigTypeInt, false, nil, nil},
			{"top", NodeConfigTypeInt, false, nil, nil},
			{"bottom", NodeConfigTypeInt, false, nil, nil},
			{"aspect_ratio_width", NodeConfigTypeInt, false, nil, nil},
			{"aspect_ratio_height", NodeConfigTypeInt, false, nil, nil},
		},
	},
	{
		NodeType: NodeTypeBlur,
		Inputs:   []InputName{"original"},
		Outputs:  []OutputName{"blurred"},
		Fields: []NodeConfigField{
			{"radius", NodeConfigTypeInt, false, nil, 2},
		},
	},
	{
		NodeType: NodeTypeResize,
		Inputs:   []InputName{"original"},
		Outputs:  []OutputName{"resized"},
		Fields: []NodeConfigField{
			{"width", NodeConfigTypeInt, false, nil, nil},
			{"height", NodeConfigTypeInt, false, nil, nil},
			{"interpolation", NodeConfigTypeOption, true, []string{
				"NearestNeighbor",
				"Bilinear",
				"Bicubic",
				"MitchellNetravali",
				"Lanczos2",
				"Lanczos3",
			}, nil},
		},
	},
	{
		NodeType: NodeTypeResizeMatch,
		Inputs:   []InputName{"original", "size_match"},
		Outputs:  []OutputName{"resized"},
		Fields: []NodeConfigField{
			{"interpolation", NodeConfigTypeOption, true, []string{
				"NearestNeighbor",
				"Bilinear",
				"Bicubic",
				"MitchellNetravali",
				"Lanczos2",
				"Lanczos3",
			}, nil},
		},
	},
	{
		NodeType: NodeTypePixelInflate,
		Inputs:   []InputName{"original"},
		Outputs:  []OutputName{"inflated"},
		Fields: []NodeConfigField{
			{"width", NodeConfigTypeInt, true, nil, nil},
			{"line_width", NodeConfigTypeInt, true, nil, nil},
			{"line_color", NodeConfigTypeString, true, nil, nil},
		},
	},
	{
		NodeType: NodeTypePaletteExtract,
		Inputs:   []InputName{"source"},
		Outputs:  []OutputName{"palette"},
		Fields: []NodeConfigField{
			{"num_colors", NodeConfigTypeInt, true, nil, 16},
			{"cluster_by", NodeConfigTypeOption, true, []string{
				"RGB",
				"HSL",
			}, "RGB"},
		},
	},
	{
		NodeType: NodeTypePaletteApply,
		Inputs:   []InputName{"source", "palette"},
		Outputs:  []OutputName{"mapped"},
	},
}

// GetNodeTypeConfig returns the config for a given node type
func GetNodeTypeConfig(nt NodeType) *NodeTypeConfig {
	for i := range NodeTypeConfigs {
		if NodeTypeConfigs[i].NodeType == nt {
			return &NodeTypeConfigs[i]
		}
	}
	return nil
}

// InputNames returns the ordered list of input names for this node type
func (nt NodeType) InputNames() []InputName {
	cfg := GetNodeTypeConfig(nt)
	if cfg == nil {
		return []InputName{}
	}
	return cfg.Inputs
}

// OutputNames returns the ordered list of output names for this node type
func (nt NodeType) OutputNames() []OutputName {
	cfg := GetNodeTypeConfig(nt)
	if cfg == nil {
		return []OutputName{}
	}
	return cfg.Outputs
}

func (nt NodeType) NameRequired() bool {
	cfg := GetNodeTypeConfig(nt)
	if cfg == nil {
		return false
	}
	return cfg.NameRequired
}
