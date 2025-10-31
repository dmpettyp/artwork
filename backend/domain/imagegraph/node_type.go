package imagegraph

import "fmt"

type NodeType int

const (
	NodeTypeNone NodeType = iota
	NodeTypeInput
	NodeTypeBlur
	NodeTypeCrop
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
	NodeConfigTypeOption
)

// NodeConfigField represents a single configuration field for a node type
type NodeConfigField struct {
	Name      string
	FieldType NodeConfigFieldType
	Required  bool
	Options   []string
}

// NodeTypeConfig represents the configuration for a node type
type NodeTypeConfig struct {
	NodeType     NodeType
	Inputs       []InputName
	Outputs      []OutputName
	NameRequired bool
	Fields       []NodeConfigField
	validate     func(NodeConfig) error
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
			{"left", NodeConfigTypeInt, true, nil},
			{"right", NodeConfigTypeInt, true, nil},
			{"top", NodeConfigTypeInt, true, nil},
			{"bottom", NodeConfigTypeInt, true, nil},
			{"aspect_ratio_width", NodeConfigTypeInt, false, nil},
			{"aspect_ratio_height", NodeConfigTypeInt, false, nil},
		},
		validate: func(config NodeConfig) error {
			left := config["left"].(int)
			right := config["right"].(int)
			top := config["top"].(int)
			bottom := config["bottom"].(int)

			// All coordinates must be non-negative
			if left < 0 || right < 0 || top < 0 || bottom < 0 {
				return fmt.Errorf("crop coordinates must be non-negative")
			}

			// Rectangle must have positive width and height
			if left >= right {
				return fmt.Errorf("left must be less than right")
			}
			if top >= bottom {
				return fmt.Errorf("top must be less than bottom")
			}

			// Validate aspect ratio if specified
			aspectWidthVal, hasAspectWidth := config["aspect_ratio_width"]
			aspectHeightVal, hasAspectHeight := config["aspect_ratio_height"]

			// Both must be set together or both omitted
			if hasAspectWidth != hasAspectHeight {
				return fmt.Errorf("aspect_ratio_width and aspect_ratio_height must both be set or both omitted")
			}

			if hasAspectWidth && hasAspectHeight {
				aspectWidth := aspectWidthVal.(int)
				aspectHeight := aspectHeightVal.(int)

				// Both must be positive
				if aspectWidth <= 0 || aspectHeight <= 0 {
					return fmt.Errorf("aspect ratio values must be positive integers")
				}

				// Validate that crop dimensions match the aspect ratio (within rounding tolerance)
				cropWidth := right - left
				cropHeight := bottom - top

				// Calculate expected aspect ratio
				expectedRatio := float64(aspectWidth) / float64(aspectHeight)
				actualRatio := float64(cropWidth) / float64(cropHeight)

				// Allow 1% tolerance for rounding
				tolerance := 0.01
				if actualRatio < expectedRatio*(1-tolerance) || actualRatio > expectedRatio*(1+tolerance) {
					return fmt.Errorf("crop dimensions (%dx%d) do not match specified aspect ratio (%d:%d)", cropWidth, cropHeight, aspectWidth, aspectHeight)
				}
			}

			return nil
		},
	},
	{
		NodeType: NodeTypeBlur,
		Inputs:   []InputName{"original"},
		Outputs:  []OutputName{"blurred"},
		Fields: []NodeConfigField{
			{"radius", NodeConfigTypeInt, true, nil},
		},
		validate: func(config NodeConfig) error {
			radius := config["radius"].(int)
			if radius < 1 {
				return fmt.Errorf("radius must be at least 1")
			}
			if radius > 100 {
				return fmt.Errorf("radius must be 100 or less")
			}
			return nil
		},
	},
	{
		NodeType: NodeTypeResize,
		Inputs:   []InputName{"original"},
		Outputs:  []OutputName{"resized"},
		Fields: []NodeConfigField{
			{"width", NodeConfigTypeInt, false, nil},
			{"height", NodeConfigTypeInt, false, nil},
			{"interpolation", NodeConfigTypeOption, true, []string{
				"NearestNeighbor",
				"Bilinear",
				"Bicubic",
				"MitchellNetravali",
				"Lanczos2",
				"Lanczos3",
			}},
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
				w := width.(int)
				if w < 1 {
					return fmt.Errorf("width must be at least 1")
				}
				if w > 10000 {
					return fmt.Errorf("width must be 10000 or less")
				}
			}

			// Validate height if present
			if hasHeight {
				h := height.(int)
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
			}},
		},
	},
}

// getNodeTypeConfig returns the config for a given node type
func getNodeTypeConfig(nt NodeType) *NodeTypeConfig {
	for i := range NodeTypeConfigs {
		if NodeTypeConfigs[i].NodeType == nt {
			return &NodeTypeConfigs[i]
		}
	}
	return nil
}

func (nt NodeType) ValidateConfig(nodeConfig NodeConfig) error {
	nodeTypeConfig := getNodeTypeConfig(nt)
	if nodeTypeConfig == nil {
		return fmt.Errorf("node type %q does not have config", nt)
	}

	// Build a map for quick lookup
	fieldMap := make(map[string]NodeConfigField)
	for _, field := range nodeTypeConfig.Fields {
		fieldMap[field.Name] = field
	}

	// Validate required fields are present
	for _, field := range nodeTypeConfig.Fields {
		if field.Required {
			if !nodeConfig.Exists(field.Name) {
				return fmt.Errorf("required field %q is missing", field.Name)
			}
		}
	}

	// Validate field types and reject unknown fields
	for key, value := range nodeConfig {
		fieldDef, exists := fieldMap[key]
		if !exists {
			return fmt.Errorf("unknown field %q", key)
		}

		switch fieldDef.FieldType {
		case NodeConfigTypeString:
			if _, ok := value.(string); !ok {
				return fmt.Errorf("field %q must be a string", key)
			}
		case NodeConfigTypeInt:
			if num, ok := value.(float64); ok {
				if num != float64(int(num)) {
					return fmt.Errorf("field %q must be an integer", key)
				}
				nodeConfig[key] = int(num)
			} else if _, ok := value.(int); !ok {
				return fmt.Errorf("field %q must be an integer", key)
			}
		case NodeConfigTypeFloat:
			if _, ok := value.(float64); !ok {
				return fmt.Errorf("field %q must be a number", key)
			}
		case NodeConfigTypeBool:
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("field %q must be a boolean", key)
			}
		case NodeConfigTypeOption:
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("field %q must be a string", key)
			}
			// Check if value is in allowed options
			validOption := false
			for _, opt := range fieldDef.Options {
				if str == opt {
					validOption = true
					break
				}
			}
			if !validOption {
				return fmt.Errorf("field %q has invalid value %q, must be one of: %v", key, str, fieldDef.Options)
			}
		}
	}

	// Run custom validator once after all fields are validated
	if nodeTypeConfig.validate != nil {
		if err := nodeTypeConfig.validate(nodeConfig); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}
	}

	return nil
}

// InputNames returns the ordered list of input names for this node type
func (nt NodeType) InputNames() []InputName {
	cfg := getNodeTypeConfig(nt)
	if cfg == nil {
		return []InputName{}
	}
	return cfg.Inputs
}

// OutputNames returns the ordered list of output names for this node type
func (nt NodeType) OutputNames() []OutputName {
	cfg := getNodeTypeConfig(nt)
	if cfg == nil {
		return []OutputName{}
	}
	return cfg.Outputs
}

func (nt NodeType) NameRequired() bool {
	cfg := getNodeTypeConfig(nt)
	if cfg == nil {
		return false
	}
	return cfg.NameRequired
}

// String converts a NodeConfigFieldType to its string representation
func (ft NodeConfigFieldType) String() string {
	switch ft {
	case NodeConfigTypeString:
		return "string"
	case NodeConfigTypeInt:
		return "int"
	case NodeConfigTypeFloat:
		return "float"
	case NodeConfigTypeBool:
		return "bool"
	case NodeConfigTypeOption:
		return "option"
	default:
		return "unknown"
	}
}
