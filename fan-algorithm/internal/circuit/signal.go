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
    ID        string      // Unique identifier for the signal
    State     SignalState // Current state of the signal
    IsBound   bool        // True if the signal is reachable from some fanout point
    IsHead    bool        // True if it's a free line adjacent to a bound line
    IsPrimary bool        // True if it's a primary input/output
    Fanouts   []*Signal   // List of signals this signal fans out to
    FanIn     *Gate       // Gate that drives this signal (nil for primary inputs)
}

// NewSignal creates a new signal with default values
func NewSignal(id string) *Signal {
    return &Signal{
        ID: id,
        State: SignalState{
            Value:    X,    // Initialize with unknown value
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
func (s *Signal) SetValue(value SignalValue) {
    s.State.Value = value
    s.State.IsStable = true
}

// GetValue returns the current value of the signal
func (s *Signal) GetValue() SignalValue {
    return s.State.Value
}

// IsUnknown checks if the signal value is unknown (X)
func (s *Signal) IsUnknown() bool {
    return s.State.Value == X
}

// IsFaulty checks if the signal carries a fault value (D or D')
func (s *Signal) IsFaulty() bool {
    return s.State.Value == D || s.State.Value == D_BAR
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
