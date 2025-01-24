// circuit.go
package circuit

import (
	"errors"
	"fmt"
)

// Circuit represents the entire digital circuit
type Circuit struct {
	Gates          []*Gate   // All gates in the circuit
	Signals        []*Signal // All signals in the circuit
	PrimaryInputs  []*Signal // Primary input signals
	PrimaryOutputs []*Signal // Primary output signals
	HeadLines      []*Signal // Head lines in the circuit
}

// NewCircuit creates a new empty circuit
func NewCircuit() *Circuit {
	return &Circuit{
		Gates:          make([]*Gate, 0),
		Signals:        make([]*Signal, 0),
		PrimaryInputs:  make([]*Signal, 0),
		PrimaryOutputs: make([]*Signal, 0),
		HeadLines:      make([]*Signal, 0),
	}
}

// AddGate adds a new gate to the circuit
func (c *Circuit) AddGate(gate *Gate) {
	c.Gates = append(c.Gates, gate)

	// Update signal lists if they're not already included
	if !c.containsSignal(gate.Output) {
		c.Signals = append(c.Signals, gate.Output)
	}
	for _, input := range gate.Inputs {
		if !c.containsSignal(input) {
			c.Signals = append(c.Signals, input)
		}
	}
}

// AddPrimaryInput adds a new primary input to the circuit
func (c *Circuit) AddPrimaryInput(signal *Signal) {
	signal.MarkAsPrimary()
	c.PrimaryInputs = append(c.PrimaryInputs, signal)
	if !c.containsSignal(signal) {
		c.Signals = append(c.Signals, signal)
	}
}

// AddPrimaryOutput adds a new primary output to the circuit
func (c *Circuit) AddPrimaryOutput(signal *Signal) {
	signal.MarkAsPrimary()
	c.PrimaryOutputs = append(c.PrimaryOutputs, signal)
	if !c.containsSignal(signal) {
		c.Signals = append(c.Signals, signal)
	}
}

// containsSignal checks if a signal is already in the circuit
func (c *Circuit) containsSignal(signal *Signal) bool {
	for _, s := range c.Signals {
		if s.ID == signal.ID {
			return true
		}
	}
	return false
}

// IdentifyBoundAndHeadLines identifies bound and head lines in the circuit
func (c *Circuit) IdentifyBoundAndHeadLines() {
	// Reset head lines
	c.HeadLines = make([]*Signal, 0)

	// First identify bound lines through fanout points
	boundLines := make(map[*Signal]bool)
	for _, signal := range c.Signals {
		if len(signal.Fanouts) > 1 {
			// Mark all signals reachable from this fanout as bound
			for _, fanout := range signal.Fanouts {
				boundLines[fanout] = true
			}
		}
	}

	// Then identify head lines - free lines adjacent to bound lines
	for _, signal := range c.Signals {
		if !boundLines[signal] { // If signal is free
			// Check if it's adjacent to any bound line
			isHead := false
			for _, fanout := range signal.Fanouts {
				if boundLines[fanout] {
					isHead = true
					break
				}
			}
			if isHead {
				signal.IsHead = true
				c.HeadLines = append(c.HeadLines, signal)
			}
		}
	}
}

// markReachableSignalsAsBound marks all signals reachable from the given signal as bound
func (c *Circuit) markReachableSignalsAsBound(signal *Signal) {
	if signal.IsBound {
		return
	}

	signal.MarkAsBound()

	// Mark all signals in fanout paths
	for _, fanout := range signal.Fanouts {
		c.markReachableSignalsAsBound(fanout)
	}
}

// Simulate performs circuit simulation with current input values
func (c *Circuit) Simulate() error {
	// Initialize all non-primary-input signals to X
	for _, signal := range c.Signals {
		if !signal.IsPrimary {
			signal.SetValue(X)
		}
	}

	// Keep simulating until no more changes occur
	changed := true
	maxIterations := len(c.Gates) * 2 // Prevent infinite loops
	iterations := 0

	for changed && iterations < maxIterations {
		changed = false
		for _, gate := range c.Gates {
			if gate.Evaluate() {
				changed = true
			}
		}
		iterations++
	}

	if iterations == maxIterations {
		return errors.New("simulation did not converge")
	}
	return nil
}

// GetSignalByID returns a signal by its ID
func (c *Circuit) GetSignalByID(id string) (*Signal, error) {
	for _, signal := range c.Signals {
		if signal.ID == id {
			return signal, nil
		}
	}
	return nil, fmt.Errorf("signal with ID %s not found", id)
}

// PrintCircuitState prints the current state of all signals in the circuit
func (c *Circuit) PrintCircuitState() {
	fmt.Println("Circuit State:")
	fmt.Println("Primary Inputs:")
	for _, signal := range c.PrimaryInputs {
		fmt.Printf("  %s\n", signal.String())
	}

	fmt.Println("Primary Outputs:")
	for _, signal := range c.PrimaryOutputs {
		fmt.Printf("  %s\n", signal.String())
	}

	fmt.Println("Head Lines:")
	for _, signal := range c.HeadLines {
		fmt.Printf("  %s\n", signal.String())
	}
}

// ValidateCircuit performs basic circuit validation
func (c *Circuit) ValidateCircuit() error {
	// Check for unconnected signals
	for _, signal := range c.Signals {
		if !signal.IsPrimary && signal.FanIn == nil && len(signal.Fanouts) == 0 {
			return fmt.Errorf("unconnected signal found: %s", signal.ID)
		}
	}

	// Check for proper gate connections
	for _, gate := range c.Gates {
		if gate.Output == nil {
			return fmt.Errorf("gate %s has no output", gate.ID)
		}
		if len(gate.Inputs) == 0 {
			return fmt.Errorf("gate %s has no inputs", gate.ID)
		}
		if gate.Type == NOT && len(gate.Inputs) != 1 {
			return fmt.Errorf("NOT gate %s must have exactly one input", gate.ID)
		}
	}

	return nil
}

// FindMandatoryPaths finds paths that must be sensitized for fault propagation
func (c *Circuit) FindMandatoryPaths(from *Signal) []*Signal {
	paths := from.GetPathsToOutputs()
	if len(paths) == 0 {
		return nil
	}

	// Find signals that appear in all paths
	signalCount := make(map[*Signal]int)
	for _, path := range paths {
		seen := make(map[*Signal]bool)
		for _, signal := range path {
			if !seen[signal] {
				signalCount[signal]++
				seen[signal] = true
			}
		}
	}

	mandatory := make([]*Signal, 0)
	for signal, count := range signalCount {
		if count == len(paths) {
			mandatory = append(mandatory, signal)
		}
	}
	return mandatory
}

// InitializeControllability sets initial controllability values
func (c *Circuit) InitializeControllability() {
	for _, gate := range c.Gates {
		gate.Controllability = gate.CalculateControllability()
		gate.Output.Controllability = gate.Controllability
	}

	// Primary inputs are easiest to control
	for _, input := range c.PrimaryInputs {
		input.Controllability = 1
	}
}
