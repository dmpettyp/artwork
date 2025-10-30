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

type nodeConfigField struct {
	name      string
	fieldType NodeConfigFieldType
	required  bool
	options   []string
}

type nodeTypeConfig struct {
	nodeType     NodeType
	inputs       []InputName
	outputs      []OutputName
	nameRequired bool
	fields       []nodeConfigField
	validate     func(NodeConfig) error
}

// nodeTypeConfigs defines all node type configurations in order
var nodeTypeConfigs = []nodeTypeConfig{
	{
		nodeType: NodeTypeInput,
		outputs:  []OutputName{"original"},
	},
	{
		nodeType:     NodeTypeOutput,
		inputs:       []InputName{"input"},
		outputs:      []OutputName{"final"},
		nameRequired: true,
	},
	{
		nodeType: NodeTypeCrop,
		inputs:   []InputName{"original"},
		outputs:  []OutputName{"cropped"},
		fields: []nodeConfigField{
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
		nodeType: NodeTypeBlur,
		inputs:   []InputName{"original"},
		outputs:  []OutputName{"blurred"},
		fields: []nodeConfigField{
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
		nodeType: NodeTypeResize,
		inputs:   []InputName{"original"},
		outputs:  []OutputName{"resized"},
		fields: []nodeConfigField{
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
		nodeType: NodeTypeResizeMatch,
		inputs:   []InputName{"original", "size_match"},
		outputs:  []OutputName{"resized"},
		fields: []nodeConfigField{
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
func getNodeTypeConfig(nt NodeType) *nodeTypeConfig {
	for i := range nodeTypeConfigs {
		if nodeTypeConfigs[i].nodeType == nt {
			return &nodeTypeConfigs[i]
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
	fieldMap := make(map[string]nodeConfigField)
	for _, field := range nodeTypeConfig.fields {
		fieldMap[field.name] = field
	}

	// Validate required fields are present
	for _, field := range nodeTypeConfig.fields {
		if field.required {
			if !nodeConfig.Exists(field.name) {
				return fmt.Errorf("required field %q is missing", field.name)
			}
		}
	}

	// Validate field types and reject unknown fields
	for key, value := range nodeConfig {
		fieldDef, exists := fieldMap[key]
		if !exists {
			return fmt.Errorf("unknown field %q", key)
		}

		switch fieldDef.fieldType {
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
			for _, opt := range fieldDef.options {
				if str == opt {
					validOption = true
					break
				}
			}
			if !validOption {
				return fmt.Errorf("field %q has invalid value %q, must be one of: %v", key, str, fieldDef.options)
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
	return cfg.inputs
}

// OutputNames returns the ordered list of output names for this node type
func (nt NodeType) OutputNames() []OutputName {
	cfg := getNodeTypeConfig(nt)
	if cfg == nil {
		return []OutputName{}
	}
	return cfg.outputs
}

func (nt NodeType) NameRequired() bool {
	cfg := getNodeTypeConfig(nt)
	if cfg == nil {
		return false
	}
	return cfg.nameRequired
}

// NodeTypeSchemaField represents a single field's schema information
type NodeTypeSchemaField struct {
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Options  []string `json:"options,omitempty"`
}

// NodeTypeSchema represents the complete schema for a node type
type NodeTypeSchema struct {
	Inputs       []string                       `json:"inputs"`
	Outputs      []string                       `json:"outputs"`
	NameRequired bool                           `json:"name_required"`
	Fields       map[string]NodeTypeSchemaField `json:"fields"`
}

// GetFieldTypeString converts a NodeConfigFieldType to its string representation
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

// GetSchema returns the schema for this node type
func (nt NodeType) GetSchema() NodeTypeSchema {
	cfg := getNodeTypeConfig(nt)
	if cfg == nil {
		return NodeTypeSchema{}
	}

	// Convert inputs
	inputs := make([]string, len(cfg.inputs))
	for i, input := range cfg.inputs {
		inputs[i] = string(input)
	}

	// Convert outputs
	outputs := make([]string, len(cfg.outputs))
	for i, output := range cfg.outputs {
		outputs[i] = string(output)
	}

	// Convert fields
	fields := make(map[string]NodeTypeSchemaField)
	for _, field := range cfg.fields {
		fields[field.name] = NodeTypeSchemaField{
			Type:     field.fieldType.String(),
			Required: field.required,
			Options:  field.options,
		}
	}

	return NodeTypeSchema{
		Inputs:       inputs,
		Outputs:      outputs,
		NameRequired: cfg.nameRequired,
		Fields:       fields,
	}
}

// NodeTypeSchemaEntry represents a node type with its schema
type NodeTypeSchemaEntry struct {
	NodeType NodeType       `json:"node_type"`
	Schema   NodeTypeSchema `json:"schema"`
}

// GetAllNodeTypeSchemas returns schemas for all node types in order
func GetAllNodeTypeSchemas() []NodeTypeSchemaEntry {
	schemas := make([]NodeTypeSchemaEntry, 0, len(nodeTypeConfigs))

	for _, cfg := range nodeTypeConfigs {
		schemas = append(schemas, NodeTypeSchemaEntry{
			NodeType: cfg.nodeType,
			Schema:   cfg.nodeType.GetSchema(),
		})
	}

	return schemas
}
