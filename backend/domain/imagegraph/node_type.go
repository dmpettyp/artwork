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
	NodeConfigTypeOption
)

type nodeConfigField struct {
	fieldType NodeConfigFieldType
	required  bool
	options   []string
}

type nodeTypeConfig struct {
	inputs       []InputName
	outputs      []OutputName
	nameRequired bool
	fields       map[string]nodeConfigField
	validate     func(NodeConfig) error
}

var nodeTypeConfigs = map[NodeType]nodeTypeConfig{
	NodeTypeInput: {
		outputs: []OutputName{"original"},
	},
	NodeTypeBlur: {
		inputs:  []InputName{"original"},
		outputs: []OutputName{"blurred"},
		fields: map[string]nodeConfigField{
			"radius": {NodeConfigTypeInt, true, nil},
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
	NodeTypeOutput: {
		inputs:       []InputName{"input"},
		outputs:      []OutputName{"final"},
		nameRequired: true,
	},
	NodeTypeResize: {
		inputs:  []InputName{"original"},
		outputs: []OutputName{"resized"},
		fields: map[string]nodeConfigField{
			"width":  {NodeConfigTypeInt, false, nil},
			"height": {NodeConfigTypeInt, false, nil},
			"interpolation": {NodeConfigTypeOption, true, []string{
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
	NodeTypeResizeMatch: {
		inputs:  []InputName{"original", "size_match"},
		outputs: []OutputName{"resized"},
		fields: map[string]nodeConfigField{
			"interpolation": {NodeConfigTypeOption, true, []string{
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

func (nt NodeType) ValidateConfig(nodeConfig NodeConfig) error {
	nodeTypeConfig, ok := nodeTypeConfigs[nt]
	if !ok {
		return fmt.Errorf("node type %q does not have config", nt)
	}

	// Validate required fields are present
	for fieldName, fieldDef := range nodeTypeConfig.fields {
		if fieldDef.required {
			if !nodeConfig.Exists(fieldName) {
				return fmt.Errorf("required field %q is missing", fieldName)
			}
		}
	}

	// Validate field types and reject unknown fields
	for key, value := range nodeConfig {
		fieldDef, exists := nodeTypeConfig.fields[key]
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
	cfg, ok := nodeTypeConfigs[nt]
	if !ok {
		return []InputName{}
	}
	return cfg.inputs
}

// OutputNames returns the ordered list of output names for this node type
func (nt NodeType) OutputNames() []OutputName {
	cfg, ok := nodeTypeConfigs[nt]
	if !ok {
		return []OutputName{}
	}
	return cfg.outputs
}

func (nt NodeType) NameRequired() bool {
	cfg, ok := nodeTypeConfigs[nt]
	if !ok {
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
	Inputs       []string                        `json:"inputs"`
	Outputs      []string                        `json:"outputs"`
	NameRequired bool                            `json:"name_required"`
	Fields       map[string]NodeTypeSchemaField  `json:"fields"`
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
	cfg, ok := nodeTypeConfigs[nt]
	if !ok {
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
	for fieldName, fieldDef := range cfg.fields {
		fields[fieldName] = NodeTypeSchemaField{
			Type:     fieldDef.fieldType.String(),
			Required: fieldDef.required,
			Options:  fieldDef.options,
		}
	}

	return NodeTypeSchema{
		Inputs:       inputs,
		Outputs:      outputs,
		NameRequired: cfg.nameRequired,
		Fields:       fields,
	}
}

// GetAllNodeTypeSchemas returns schemas for all node types
func GetAllNodeTypeSchemas() map[string]NodeTypeSchema {
	schemas := make(map[string]NodeTypeSchema)

	// Map of NodeType to string representation (matches nodeTypeMapper in HTTP layer)
	nodeTypes := []struct {
		nodeType NodeType
		name     string
	}{
		{NodeTypeInput, "input"},
		{NodeTypeBlur, "blur"},
		{NodeTypeOutput, "output"},
		{NodeTypeResize, "resize"},
		{NodeTypeResizeMatch, "resize_match"},
	}

	for _, nt := range nodeTypes {
		schemas[nt.name] = nt.nodeType.GetSchema()
	}

	return schemas
}
