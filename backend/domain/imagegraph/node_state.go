package imagegraph

import "encoding/json"

type NodeState int

const (
	Waiting NodeState = iota
	Generating
	Generated
)

func (s NodeState) MarshalJSON() ([]byte, error) {
	str := NodeStateMapper.FromWithDefault(s, "unknown")
	return json.Marshal(str)
}

func (s NodeState) Transitions() map[NodeState][]NodeState {
	return map[NodeState][]NodeState{
		Waiting:    {Generating, Waiting},
		Generating: {Generated, Waiting, Generating},
		Generated:  {Waiting, Generating, Generated},
	}
}

func AllNodeStates() []NodeState {
	return []NodeState{
		Waiting,
		Generating,
		Generated,
	}
}
