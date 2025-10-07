package imagegraph

type State string

const (
	WaitingForInputs  State = "WaitingForInputs"
	GeneratingOutputs       = "GeneratingOutputs"
	OutputsGenerated        = "OutputsGenerated"
)

func (s State) Transitions() map[State][]State {
	return map[State][]State{
		WaitingForInputs:  {GeneratingOutputs},
		GeneratingOutputs: {OutputsGenerated, WaitingForInputs, GeneratingOutputs},
		OutputsGenerated:  {WaitingForInputs, GeneratingOutputs},
	}
}
