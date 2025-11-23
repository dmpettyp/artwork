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

// NodeTypeConfig represents the configuration for a node type
type NodeTypeConfig struct {
	NodeType     NodeType
	Inputs       []InputName
	Outputs      []OutputName
	NameRequired bool
	NewConfig    func() NodeConfig
}

// NodeTypeConfigs defines all node type configurations in order
var NodeTypeConfigs = []NodeTypeConfig{
	{
		NodeType:  NodeTypeInput,
		Outputs:   []OutputName{"original"},
		NewConfig: func() NodeConfig { return NewNodeConfigInput() },
	},
	{
		NodeType:     NodeTypeOutput,
		Inputs:       []InputName{"input"},
		Outputs:      []OutputName{"final"},
		NameRequired: true,
		NewConfig:    func() NodeConfig { return NewNodeConfigOutput() },
	},
	{
		NodeType:  NodeTypeCrop,
		Inputs:    []InputName{"original"},
		Outputs:   []OutputName{"cropped"},
		NewConfig: func() NodeConfig { return NewNodeConfigCrop() },
	},
	{
		NodeType:  NodeTypeBlur,
		Inputs:    []InputName{"original"},
		Outputs:   []OutputName{"blurred"},
		NewConfig: func() NodeConfig { return NewNodeConfigBlur() },
	},
	{
		NodeType:  NodeTypeResize,
		Inputs:    []InputName{"original"},
		Outputs:   []OutputName{"resized"},
		NewConfig: func() NodeConfig { return NewNodeConfigResize() },
	},
	{
		NodeType:  NodeTypeResizeMatch,
		Inputs:    []InputName{"original", "size_match"},
		Outputs:   []OutputName{"resized"},
		NewConfig: func() NodeConfig { return NewNodeConfigResizeMatch() },
	},
	{
		NodeType:  NodeTypePixelInflate,
		Inputs:    []InputName{"original"},
		Outputs:   []OutputName{"inflated"},
		NewConfig: func() NodeConfig { return NewNodeConfigPixelInflate() },
	},
	{
		NodeType:  NodeTypePaletteExtract,
		Inputs:    []InputName{"source"},
		Outputs:   []OutputName{"palette"},
		NewConfig: func() NodeConfig { return NewNodeConfigPaletteExtract() },
	},
	{
		NodeType:  NodeTypePaletteApply,
		Inputs:    []InputName{"source", "palette"},
		Outputs:   []OutputName{"mapped"},
		NewConfig: func() NodeConfig { return NewNodeConfigPaletteApply() },
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
