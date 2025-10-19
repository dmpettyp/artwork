package imagegraph

type NodeType int

const (
	NodeTypeNone NodeType = iota
	NodeTypeInput
	NodeTypeScale
	NodeTypeBlur
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
	inputs  []InputName
	outputs []OutputName
	fields  map[string]nodeConfigField
}

var nodeConfigs = map[NodeType]nodeConfig{
	NodeTypeInput: {
		outputs: []OutputName{"original"},
	},
	NodeTypeScale: {
		inputs:  []InputName{"original"},
		outputs: []OutputName{"scaled"},
		fields: map[string]nodeConfigField{
			"factor": {NodeConfigTypeFloat, true},
		},
	},
	NodeTypeBlur: {
		inputs:  []InputName{"original"},
		outputs: []OutputName{"blurred"},
		fields: map[string]nodeConfigField{
			"radius": {NodeConfigTypeInt, true},
		},
	},
}
