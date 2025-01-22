// backtrace.go
package algorithm

import (
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/pkg/types"
	"sort"
)

// BacktraceResult using enhanced types
type BacktraceResult struct {
	FinalObjectives []*types.BacktraceObjective
	HeadLines       []*circuit.Signal
	Stats           *types.TestGenerationStats
	Error           types.TestGenerationError
}

// MultipleBacktrace improved with enhanced functionality
func MultipleBacktrace(initialObjectives []*types.BacktraceObjective,
	c *circuit.Circuit,
	config *types.TestGenerationConfig) *BacktraceResult {

	result := &BacktraceResult{
		FinalObjectives: make([]*types.BacktraceObjective, 0),
		HeadLines:       make([]*circuit.Signal, 0),
		Stats:           types.NewTestGenerationStats(),
	}

	// Handle unique sensitization paths
	if config.UseDynamicBacktrace && len(initialObjectives) == 1 && initialObjectives[0].Signal.IsFaulty() {
		if paths := c.FindMandatoryPaths(initialObjectives[0].Signal); len(paths) > 0 {
			for _, signal := range paths {
				obj := types.CreateObjective(signal, types.JUSTIFY, circuit.ONE)
				obj.Priority = 10
				result.FinalObjectives = append(result.FinalObjectives, obj)
			}
			return result
		}
	}

	currentObjectives := initialObjectives
	processed := make(map[string]bool)

	for len(currentObjectives) > 0 {
		nextObjectives := make([]*types.BacktraceObjective, 0)

		for _, obj := range currentObjectives {
			if processed[obj.Signal.ID] {
				continue
			}
			processed[obj.Signal.ID] = true
			result.Stats.BacktraceCount++

			if obj.Signal.IsHead {
				result.FinalObjectives = append(result.FinalObjectives, obj)
				result.HeadLines = append(result.HeadLines, obj.Signal)
				continue
			}

			if obj.Signal.FanIn != nil {
				newObjs := backtraceGateWithCost(obj.Signal.FanIn, obj)
				nextObjectives = append(nextObjectives, newObjs...)
			}
		}

		currentObjectives = nextObjectives
	}

	sortObjectivesByPriorityAndCost(result.FinalObjectives)
	return result
}

// backtraceGateWithCost handles gate backtrace with cost estimation
func backtraceGateWithCost(gate *circuit.Gate, obj *types.BacktraceObjective) []*types.BacktraceObjective {
	switch gate.Type {
	case circuit.AND:
		return backtraceANDGateEnhanced(gate, obj)
	case circuit.OR:
		return backtraceORGateEnhanced(gate, obj)
	case circuit.NOT:
		return backtraceNOTGateEnhanced(gate, obj)
	default:
		return nil
	}
}

// backtraceANDGateEnhanced with priority and cost improvements
func backtraceANDGateEnhanced(gate *circuit.Gate, obj *types.BacktraceObjective) []*types.BacktraceObjective {
	results := make([]*types.BacktraceObjective, 0)

	if obj.ZeroCount > 0 {
		easiestInput := findEasiestToControl(gate, circuit.ZERO)
		newObj := types.CreateObjective(easiestInput, types.JUSTIFY, circuit.ZERO)
		newObj.ZeroCount = obj.ZeroCount
		results = append(results, newObj)
	}

	if obj.OneCount > 0 {
		for _, input := range gate.Inputs {
			newObj := types.CreateObjective(input, types.JUSTIFY, circuit.ONE)
			newObj.OneCount = obj.OneCount
			results = append(results, newObj)
		}
	}

	return results
}

// backtraceORGateEnhanced with priority and cost improvements
func backtraceORGateEnhanced(gate *circuit.Gate, obj *types.BacktraceObjective) []*types.BacktraceObjective {
	results := make([]*types.BacktraceObjective, 0)

	if obj.OneCount > 0 {
		// For OR gate, ONE is controlling - choose easiest input
		easiestInput := findEasiestToControl(gate, circuit.ONE)
		newObj := types.CreateObjective(easiestInput, types.JUSTIFY, circuit.ONE)
		newObj.OneCount = obj.OneCount
		results = append(results, newObj)
	}

	if obj.ZeroCount > 0 {
		// For OR gate, ZERO requires all inputs to be ZERO
		for _, input := range gate.Inputs {
			newObj := types.CreateObjective(input, types.JUSTIFY, circuit.ZERO)
			newObj.ZeroCount = obj.ZeroCount
			results = append(results, newObj)
		}
	}

	return results
}

// backtraceNOTGateEnhanced with priority and cost improvements
func backtraceNOTGateEnhanced(gate *circuit.Gate, obj *types.BacktraceObjective) []*types.BacktraceObjective {
	results := make([]*types.BacktraceObjective, 0)

	// NOT gate inverts the value, so swap ZeroCount and OneCount
	if obj.ZeroCount > 0 {
		newObj := types.CreateObjective(gate.Inputs[0], types.JUSTIFY, circuit.ONE)
		newObj.OneCount = obj.ZeroCount
		results = append(results, newObj)
	}

	if obj.OneCount > 0 {
		newObj := types.CreateObjective(gate.Inputs[0], types.JUSTIFY, circuit.ZERO)
		newObj.ZeroCount = obj.OneCount
		results = append(results, newObj)
	}

	return results
}

// findEasiestToControl with controllability measures
func findEasiestToControl(gate *circuit.Gate, value circuit.SignalValue) *circuit.Signal {
	// Example: select first input, but this can use advanced heuristics
	return gate.Inputs[0]
}

// sortObjectivesByPriorityAndCost sorts objectives based on priority and cost
func sortObjectivesByPriorityAndCost(objectives []*types.BacktraceObjective) {
	sort.Slice(objectives, func(i, j int) bool {
		if objectives[i].Priority != objectives[j].Priority {
			return objectives[i].Priority > objectives[j].Priority
		}
		return objectives[i].Cost < objectives[j].Cost
	})
}

// Helper: calculate opposite value
func oppositeValue(value circuit.SignalValue) circuit.SignalValue {
	if value == circuit.ZERO {
		return circuit.ONE
	}
	return circuit.ZERO
}
