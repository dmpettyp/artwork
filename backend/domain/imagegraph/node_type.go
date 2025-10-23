package imagegraph

import "fmt"

type NodeType int

const (
	NodeTypeNone NodeType = iota
	NodeTypeInput
	NodeTypeBlur
	NodeTypeOutput
	NodeTypeResize
	NodeTypeResizeMatch
)

type NodeConfigFieldType int

const (
	NodeConfigTypeNone NodeConfigFieldType = iota
	NodeConfigTypeString
	NodeConfigTypeInt
	NodeConfigTypeFloat
	NodeConfigTypeBool
)

type nodeConfigField struct {
	fieldType NodeConfigFieldType
	required  bool
}

type nodeConfig struct {
	inputs   []InputName
	outputs  []OutputName
	fields   map[string]nodeConfigField
	validate func(NodeConfig) error
}

var nodeConfigs = map[NodeType]nodeConfig{
	NodeTypeInput: {
		outputs: []OutputName{"original"},
	},
	NodeTypeBlur: {
		inputs:  []InputName{"original"},
		outputs: []OutputName{"blurred"},
		fields: map[string]nodeConfigField{
			"radius": {NodeConfigTypeInt, true},
		},
		validate: func(config NodeConfig) error {
			radius := config["radius"].(float64)
			if radius < 1 {
				return fmt.Errorf("radius must be at least 1")
			}
			if radius > 100 {
				return fmt.Errorf("radius must be 100 or less")
			}
			return nil
		},
	},
	NodeTypeOutput: {
		inputs:  []InputName{"input"},
		outputs: []OutputName{"final"},
	},
	NodeTypeResize: {
		inputs:  []InputName{"original"},
		outputs: []OutputName{"resized"},
		fields: map[string]nodeConfigField{
			"width":  {NodeConfigTypeInt, false},
			"height": {NodeConfigTypeInt, false},
		},
		validate: func(config NodeConfig) error {
			width, hasWidth := config["width"]
			height, hasHeight := config["height"]

			// At least one of width or height must be set
			if !hasWidth && !hasHeight {
				return fmt.Errorf("at least one of width or height must be set")
			}

			// Validate width if present
			if hasWidth {
				w := width.(float64)
				if w < 1 {
					return fmt.Errorf("width must be at least 1")
				}
				if w > 10000 {
					return fmt.Errorf("width must be 10000 or less")
				}
			}

			// Validate height if present
			if hasHeight {
				h := height.(float64)
				if h < 1 {
					return fmt.Errorf("height must be at least 1")
				}
				if h > 10000 {
					return fmt.Errorf("height must be 10000 or less")
				}
			}

			return nil
		},
	},
	NodeTypeResizeMatch: {
		inputs:  []InputName{"original", "size_match"},
		outputs: []OutputName{"resized"},
	},
}
