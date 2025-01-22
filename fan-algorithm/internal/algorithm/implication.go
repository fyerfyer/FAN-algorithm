package algorithm

import (
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/pkg/types"
	"time"
)

// Implication performs forward and backward implication
func Implication(c *circuit.Circuit, assignment types.Assignment) types.TestResult {
	result := types.NewTestResult()
	result.Implications = append(result.Implications, assignment)

	queue := []types.Assignment{assignment}
	processed := make(map[string]bool)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if processed[current.Signal.ID] {
			continue
		}
		processed[current.Signal.ID] = true

		// Forward implication
		if !forwardImplicate(current.Signal, &queue, result) {
			result.Success = false
			result.Error = types.ErrInconsistency
			return *result
		}

		// Backward implication
		if !backwardImplicate(current.Signal, &queue, result) {
			result.Success = false
			result.Error = types.ErrInconsistency
			return *result
		}

		result.Stats.Implications++
	}

	result.Success = true
	return *result
}

func forwardImplicate(signal *circuit.Signal, queue *[]types.Assignment, result *types.TestResult) bool {
	for _, fanout := range signal.Fanouts {
		if fanout.FanIn == nil {
			continue
		}

		gate := fanout.FanIn
		if gate.Evaluate() {
			newValue := gate.Output.GetValue()
			if !signal.IsCompatible(newValue) {
				return false
			}

			assignment := types.Assignment{
				Signal:    gate.Output,
				Value:     newValue,
				Reason:    types.IMPLICATION,
				Level:     result.CircuitState.DecisionLevel,
				TimeStamp: time.Now(),
			}

			*queue = append(*queue, assignment)
			result.Implications = append(result.Implications, assignment)
			result.CircuitState.SignalValues[gate.Output] = newValue
		}
	}
	return true
}

func backwardImplicate(signal *circuit.Signal, queue *[]types.Assignment, result *types.TestResult) bool {
	if signal.FanIn == nil {
		return true
	}

	gate := signal.FanIn
	var implResult bool

	switch gate.Type {
	case circuit.AND:
		implResult = backwardImplicateAND(gate, signal, queue, result)
	case circuit.OR:
		implResult = backwardImplicateOR(gate, signal, queue, result)
	case circuit.NOT:
		implResult = backwardImplicateNOT(gate, signal, queue, result)
	}

	return implResult
}

func backwardImplicateAND(gate *circuit.Gate, signal *circuit.Signal, queue *[]types.Assignment, result *types.TestResult) bool {
	// Logic for AND gate backward implication
	// Similar updates for other gate implication functions...
	if gate.Output.GetValue() == circuit.ONE {
		for _, input := range gate.Inputs {
			if input.IsUnknown() {
				assignment := types.Assignment{
					Signal:    input,
					Value:     circuit.ONE,
					Reason:    types.IMPLICATION,
					Level:     result.CircuitState.DecisionLevel,
					TimeStamp: time.Now(),
				}

				if !signal.IsCompatible(circuit.ONE) {
					return false
				}

				*queue = append(*queue, assignment)
				result.Implications = append(result.Implications, assignment)
				result.CircuitState.SignalValues[input] = circuit.ONE
			}
		}
	}
	return true
}

func backwardImplicateOR(gate *circuit.Gate, signal *circuit.Signal, queue *[]types.Assignment, result *types.TestResult) bool {
	if gate.Output.GetValue() == circuit.ZERO {
		for _, input := range gate.Inputs {
			if input.IsUnknown() {
				assignment := types.Assignment{
					Signal:    input,
					Value:     circuit.ZERO,
					Reason:    types.IMPLICATION,
					Level:     result.CircuitState.DecisionLevel,
					TimeStamp: time.Now(),
				}

				if !signal.IsCompatible(circuit.ZERO) {
					return false
				}

				*queue = append(*queue, assignment)
				result.Implications = append(result.Implications, assignment)
				result.CircuitState.SignalValues[input] = circuit.ZERO
			}
		}
	}
	return true
}

func backwardImplicateNOT(gate *circuit.Gate, signal *circuit.Signal, queue *[]types.Assignment, result *types.TestResult) bool {
	input := gate.Inputs[0]
	if input.IsUnknown() {
		var impliedValue circuit.SignalValue
		switch gate.Output.GetValue() {
		case circuit.ZERO:
			impliedValue = circuit.ONE
		case circuit.ONE:
			impliedValue = circuit.ZERO
		default:
			return true
		}

		assignment := types.Assignment{
			Signal:    input,
			Value:     impliedValue,
			Reason:    types.IMPLICATION,
			Level:     result.CircuitState.DecisionLevel,
			TimeStamp: time.Now(),
		}

		if !signal.IsCompatible(impliedValue) {
			return false
		}

		*queue = append(*queue, assignment)
		result.Implications = append(result.Implications, assignment)
		result.CircuitState.SignalValues[input] = impliedValue
	}
	return true
}
