package imagegraph

type NodeType int

const (
	NodeTypeNone NodeType = iota
	NodeTypeInput
)

type nodeConfig struct {
	inputNames  []InputName
	outputNames []OutputName
}

var nodeConfigs = map[NodeType]nodeConfig{
	NodeTypeInput: {
		outputNames: []OutputName{"source"},
	},
}
