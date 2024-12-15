package und

// State is
type State int

const (
	StateUndefined = State(1 << iota)
	StateNull
	StateDefined
)
