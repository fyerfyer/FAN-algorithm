// helper.go
package utils

import (
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
)

// ObjectiveType represents the type of objective in FAN algorithm
type ObjectiveType int

const (
	PROPAGATE ObjectiveType = iota // Propagate D or D'
	JUSTIFY                        // Justify a signal value
)

// Objective represents a test generation objective
type Objective struct {
	Type      ObjectiveType
	Signal    *circuit.Signal
	Value     circuit.SignalValue
	ZeroCount int // Number of times 0 is required
	OneCount  int // Number of times 1 is required
}

// DFrontier represents a gate in the D-frontier
type DFrontier struct {
	Gate  *circuit.Gate
	Input *circuit.Signal // The input carrying D or D'
}

// Assignment represents a value assignment to a signal
type Assignment struct {
	Signal *circuit.Signal
	Value  circuit.SignalValue
}

// FindDFrontier finds all gates in the D-frontier
// D-frontier consists of gates with D/D' on input but not on output
func FindDFrontier(c *circuit.Circuit) []DFrontier {
	frontier := make([]DFrontier, 0)

	for _, gate := range c.Gates {
		// Check if any input has D/D'
		var faultyInput *circuit.Signal
		for _, input := range gate.Inputs {
			if input.IsFaulty() {
				faultyInput = input
				break
			}
		}

		// If found faulty input and output is not faulty, add to frontier
		if faultyInput != nil && !gate.Output.IsFaulty() {
			frontier = append(frontier, DFrontier{
				Gate:  gate,
				Input: faultyInput,
			})
		}
	}

	return frontier
}

// FindSensitizationPaths finds paths that must be sensitized from a gate to primary outputs
func FindSensitizationPaths(gate *circuit.Gate, c *circuit.Circuit) [][]*circuit.Gate {
	paths := make([][]*circuit.Gate, 0)
	visited := make(map[*circuit.Gate]bool)

	var dfs func(*circuit.Gate, []*circuit.Gate)
	dfs = func(current *circuit.Gate, path []*circuit.Gate) {
		if visited[current] {
			return
		}
		visited[current] = true

		newPath := append(path, current)

		// If output is primary output, we found a path
		if current.Output.IsPrimary {
			paths = append(paths, newPath)
			return
		}

		// Continue DFS through fanouts
		for _, fanout := range current.Output.Fanouts {
			if fanout.FanIn != nil {
				dfs(fanout.FanIn, newPath)
			}
		}
	}

	dfs(gate, make([]*circuit.Gate, 0))
	return paths
}

// CheckImplicationConsistency checks if a new assignment creates any inconsistency
func CheckImplicationConsistency(signal *circuit.Signal, value circuit.SignalValue) bool {
	return signal.IsCompatible(value)
}

// GetControllingInputs returns all inputs of a gate that have controlling values
func GetControllingInputs(gate *circuit.Gate) []*circuit.Signal {
	controllingInputs := make([]*circuit.Signal, 0)

	for _, input := range gate.Inputs {
		if gate.IsControllingValue(input.GetValue()) {
			controllingInputs = append(controllingInputs, input)
		}
	}

	return controllingInputs
}

// IsPathSensitized checks if a path is sensitized
func IsPathSensitized(path []*circuit.Gate) bool {
	for _, gate := range path {
		// Check if all non-faulty inputs have non-controlling values
		for _, input := range gate.Inputs {
			if !input.IsFaulty() && gate.IsControllingValue(input.GetValue()) {
				return false
			}
		}
	}
	return true
}

// GetUnassignedInputs returns all inputs of a gate that have unknown (X) values
func GetUnassignedInputs(gate *circuit.Gate) []*circuit.Signal {
	unassigned := make([]*circuit.Signal, 0)
	for _, input := range gate.Inputs {
		if input.IsUnknown() {
			unassigned = append(unassigned, input)
		}
	}
	return unassigned
}

// CreateObjective creates a new objective for the FAN algorithm
func CreateObjective(signal *circuit.Signal, objType ObjectiveType, value circuit.SignalValue) Objective {
	return Objective{
		Type:      objType,
		Signal:    signal,
		Value:     value,
		ZeroCount: 0,
		OneCount:  0,
	}
}

// calculatePathScore computes priority score for a sensitization path
func CalculatePathScore(gates []*circuit.Gate) int {
	score := 0
	for _, gate := range gates {
		// Higher score for gates with fewer inputs (easier to sensitize)
		score += 10 - len(gate.Inputs)
		// Bonus for gates close to primary outputs
		if gate.Output.IsPrimary {
			score += 5
		}
	}
	return score
}

// getAlternativeValue returns the opposite value
func GetAlternativeValue(value circuit.SignalValue) circuit.SignalValue {
	switch value {
	case circuit.ZERO:
		return circuit.ONE
	case circuit.ONE:
		return circuit.ZERO
	case circuit.D:
		return circuit.D_BAR
	case circuit.D_BAR:
		return circuit.D
	default:
		return circuit.X
	}
}

// IsPathSensitizable checks if a path can be sensitized
func IsPathSensitizable(path []*circuit.Gate) bool {
	for _, gate := range path {
		// Check controlling values on side inputs
		for _, input := range gate.Inputs {
			if input.IsFaulty() {
				continue
			}
			if gate.IsControllingValue(input.GetValue()) {
				return false
			}
		}
	}
	return true
}

// GetInputValues returns a slice of values for gate inputs
func GetInputValues(g *circuit.Gate) []circuit.SignalValue {
	values := make([]circuit.SignalValue, len(g.Inputs))
	for i, input := range g.Inputs {
		values[i] = input.GetValue()
	}
	return values
}
