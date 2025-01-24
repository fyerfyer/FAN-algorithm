// signal.go
package circuit

// SignalValue represents possible values of a signal
type SignalValue int

const (
	ZERO SignalValue = iota
	ONE
	D
	D_BAR
	X
)

// SignalState represents the current state of a signal in the circuit
type SignalState struct {
	Value    SignalValue // Current value of the signal
	IsStable bool        // Whether the value is stable or might change
}

// Signal represents a signal line in the circuit
type Signal struct {
	ID               string      // Unique identifier for the signal
	State            SignalState // Current state of the signal
	IsBound          bool        // True if the signal is reachable from some fanout point
	IsHead           bool        // True if it's a free line adjacent to a bound line
	IsPrimary        bool        // True if it's a primary input/output
	Fanouts          []*Signal   // List of signals this signal fans out to
	FanIn            *Gate       // Gate that drives this signal (nil for primary inputs)
	ControllingValue SignalValue // The controlling value for its fanin gate
	Controllability  int         // Controllability metric
	IsFault          bool
	FaultType        SignalValue
	Value            SignalValue
}

// NewSignal creates a new signal with default values
func NewSignal(id string) *Signal {
	return &Signal{
		ID: id,
		State: SignalState{
			Value:    X, // Initialize with unknown value
			IsStable: false,
		},
		IsBound:   false,
		IsHead:    false,
		IsPrimary: false,
		Fanouts:   make([]*Signal, 0),
		FanIn:     nil,
	}
}

// SetValue sets the value of the signal and marks it as stable
func (s *Signal) SetValue(v SignalValue) bool {
	if s.IsFault {
		return true // Keep faulty value
	}
	if s.Value == v {
		return true
	}
	s.Value = v
	return true
}

// GetValue returns the current value of the signal
func (s *Signal) GetValue() SignalValue {
	if s.IsFault {
		if s.FaultType == ZERO {
			return D
		}
		return D_BAR
	}
	return s.Value
}

// IsUnknown checks if the signal value is unknown (X)
func (s *Signal) IsUnknown() bool {
	return s.State.Value == X
}

// IsFaulty checks if the signal carries a fault value (D or D')
func (s *Signal) IsFaulty() bool {
	val := s.GetValue()
	return val == D || val == D_BAR
}

// AddFanout adds a fanout connection to this signal
func (s *Signal) AddFanout(fanout *Signal) {
	s.Fanouts = append(s.Fanouts, fanout)
}

// SetFanIn sets the gate that drives this signal
func (s *Signal) SetFanIn(gate *Gate) {
	s.FanIn = gate
}

// HasFanout checks if the signal has any fanout points
func (s *Signal) HasFanout() bool {
	return len(s.Fanouts) > 0
}

// MarkAsHead marks this signal as a head line
// A head line is a free line adjacent to a bound line
func (s *Signal) MarkAsHead() {
	s.IsHead = true
}

// MarkAsBound marks this signal as bound
// A bound line is reachable from some fanout point
func (s *Signal) MarkAsBound() {
	s.IsBound = true
}

// MarkAsPrimary marks this signal as a primary input/output
func (s *Signal) MarkAsPrimary() {
	s.IsPrimary = true
}

// IsCompatible checks if the current signal value is compatible with a new value
func (s *Signal) IsCompatible(newValue SignalValue) bool {
	// If current value is unknown (X), any new value is compatible
	if s.State.Value == X {
		return true
	}

	// If new value is unknown (X), it's compatible with any current value
	if newValue == X {
		return true
	}

	// Check specific value compatibility
	switch s.State.Value {
	case ZERO:
		return newValue == ZERO || newValue == D_BAR
	case ONE:
		return newValue == ONE || newValue == D
	case D:
		return newValue == ONE || newValue == D
	case D_BAR:
		return newValue == ZERO || newValue == D_BAR
	default:
		return false
	}
}

// String returns a string representation of the signal
func (s *Signal) String() string {
	valueStr := map[SignalValue]string{
		ZERO:  "0",
		ONE:   "1",
		D:     "D",
		D_BAR: "D'",
		X:     "X",
	}

	return s.ID + "=" + valueStr[s.State.Value]
}

// Clone creates a deep copy of the signal (without fanout and fanin connections)
func (s *Signal) Clone() *Signal {
	// Clone without fanout and fanin connections
	return &Signal{
		ID: s.ID + "_clone",
		State: SignalState{
			Value:    s.State.Value,
			IsStable: s.State.IsStable,
		},
		IsBound:   s.IsBound,
		IsHead:    s.IsHead,
		IsPrimary: s.IsPrimary,
		Fanouts:   make([]*Signal, 0),
		FanIn:     nil,
	}
}

// GetReachableFanouts returns all fanout points reachable from this signal
func (s *Signal) GetReachableFanouts() []*Signal {
	visited := make(map[*Signal]bool)
	fanouts := make([]*Signal, 0)

	var dfs func(*Signal)
	dfs = func(curr *Signal) {
		if visited[curr] {
			return
		}
		visited[curr] = true
		if curr.HasFanout() {
			fanouts = append(fanouts, curr)
		}
		for _, f := range curr.Fanouts {
			dfs(f)
		}
	}
	dfs(s)
	return fanouts
}

// IsFanoutPoint checks if this signal is a fanout point
func (s *Signal) IsFanoutPoint() bool {
	return len(s.Fanouts) > 1
}

// GetPathsToOutputs finds all paths from this signal to primary outputs
func (s *Signal) GetPathsToOutputs() [][]*Signal {
	paths := make([][]*Signal, 0)
	visited := make(map[*Signal]bool)

	var dfs func(*Signal, []*Signal)
	dfs = func(curr *Signal, path []*Signal) {
		if visited[curr] {
			return
		}
		visited[curr] = true

		newPath := append(path, curr)
		if curr.IsPrimary && len(curr.Fanouts) == 0 {
			paths = append(paths, newPath)
			return
		}

		for _, f := range curr.Fanouts {
			visited[curr] = false // Allow multiple paths through fanout points
			dfs(f, newPath)
		}
	}

	dfs(s, make([]*Signal, 0))
	return paths
}

// SetFault sets the fault value of the signal
func (s *Signal) SetFault(value SignalValue) {
	s.IsFault = true
	s.FaultType = value
	if value == ZERO {
		s.Value = D
	} else {
		s.Value = D_BAR
	}
}
