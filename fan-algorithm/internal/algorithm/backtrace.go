// backtrace.go
package algorithm

import (
	"sort"

	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/sensitization"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/utils"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/pkg/types"
)

// BacktraceResult using enhanced types
type BacktraceResult struct {
	FinalObjectives []*types.BacktraceObjective
	HeadLines       []*circuit.Signal
	Stats           *types.TestGenerationStats
	Error           types.TestGenerationError
}

// MultipleBacktrace improved with enhanced functionality
func MultipleBacktrace(initialObjectives []*types.BacktraceObjective, c *circuit.Circuit, config *types.TestGenerationConfig) *BacktraceResult {
	result := &BacktraceResult{
		FinalObjectives: make([]*types.BacktraceObjective, 0),
		HeadLines:       make([]*circuit.Signal, 0),
		Stats:           types.NewTestGenerationStats(),
	}

	// Always include initial objectives in final objectives
	for _, obj := range initialObjectives {
		result.FinalObjectives = append(result.FinalObjectives, obj)
		if obj.Signal.IsHead {
			result.HeadLines = append(result.HeadLines, obj.Signal)
		}
	}

	// Process unique sensitization if enabled
	if config.UseUniqueSensitization && len(initialObjectives) > 0 {
		result.Stats.Decisions++
		if paths := findMandatoryPaths(initialObjectives[0].Signal, c); len(paths) > 0 {
			for _, signal := range paths {
				obj := &types.BacktraceObjective{
					Signal:    signal,
					Value:     circuit.ONE,
					Priority:  10,
					OneCount:  1,
					ZeroCount: 0,
				}
				result.FinalObjectives = append(result.FinalObjectives, obj)
			}
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

			// Process through gates if available
			if obj.Signal.FanIn != nil {
				newObjs := backtraceGateWithCost(obj.Signal.FanIn, obj)
				for _, newObj := range newObjs {
					// Add to final objectives if head line
					if newObj.Signal.IsHead {
						result.HeadLines = append(result.HeadLines, newObj.Signal)
						result.FinalObjectives = append(result.FinalObjectives, newObj)
					}
					nextObjectives = append(nextObjectives, newObj)
				}
			}
		}

		currentObjectives = nextObjectives
	}

	// Ensure we have at least one objective
	if len(result.FinalObjectives) == 0 {
		result.FinalObjectives = append(result.FinalObjectives, initialObjectives...)
	}

	sortObjectivesByPriorityAndCost(result.FinalObjectives)
	return result
}

// backtraceGateWithCost handles gate backtrace with cost estimation
func backtraceGateWithCost(gate *circuit.Gate, obj *types.BacktraceObjective) []*types.BacktraceObjective {
	results := make([]*types.BacktraceObjective, 0)

	switch gate.Type {
	case circuit.AND:
		if obj.Value == circuit.ONE {
			// AND=1 requires all inputs=1
			for _, input := range gate.Inputs {
				newObj := &types.BacktraceObjective{
					Signal:    input,
					Value:     circuit.ONE,
					OneCount:  obj.OneCount,
					ZeroCount: 0,
					Priority:  obj.Priority - 1,
				}
				results = append(results, newObj)
			}
		} else {
			// AND=0 requires any input=0
			easiest := gate.GetEasiestControllingInput()
			results = append(results, &types.BacktraceObjective{
				Signal:    easiest,
				Value:     circuit.ZERO,
				ZeroCount: obj.ZeroCount,
				OneCount:  0,
				Priority:  obj.Priority - 1,
			})
		}
	case circuit.OR:
		if obj.Value == circuit.ZERO {
			// OR=0 requires all inputs=0
			for _, input := range gate.Inputs {
				newObj := &types.BacktraceObjective{
					Signal:    input,
					Value:     circuit.ZERO,
					ZeroCount: obj.ZeroCount,
					OneCount:  0,
					Priority:  obj.Priority - 1,
				}
				results = append(results, newObj)
			}
		} else {
			// OR=1 requires any input=1
			easiest := gate.GetEasiestControllingInput()
			results = append(results, &types.BacktraceObjective{
				Signal:    easiest,
				Value:     circuit.ONE,
				OneCount:  obj.OneCount,
				ZeroCount: 0,
				Priority:  obj.Priority - 1,
			})
		}
	case circuit.NOT:
		results = append(results, &types.BacktraceObjective{
			Signal:    gate.Inputs[0],
			Value:     utils.GetAlternativeValue(obj.Value),
			OneCount:  obj.ZeroCount,
			ZeroCount: obj.OneCount,
			Priority:  obj.Priority,
		})
	}

	return results
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

func findMandatoryPaths(signal *circuit.Signal, c *circuit.Circuit) []*circuit.Signal {
	if signal.FanIn == nil {
		return nil
	}

	// Find all paths to outputs
	paths := findAllPaths(signal, c)
	if len(paths) == 0 {
		return nil
	}

	// Find signals that appear in all paths
	mandatory := findCommonSignals(paths)
	return mandatory
}

func findAllPaths(signal *circuit.Signal, c *circuit.Circuit) [][]*circuit.Signal {
	// Use PathFinder from sensitization package
	pf := sensitization.NewPathFinder(c)
	dFrontier := []*circuit.Gate{signal.FanIn}
	paths := pf.FindUniqueSensitizationPaths(dFrontier)

	// Convert Path objects to signal paths
	signalPaths := make([][]*circuit.Signal, 0)
	for _, path := range paths {
		signalPath := make([]*circuit.Signal, 0)
		for _, gate := range path.Gates {
			signalPath = append(signalPath, gate.Output)
		}
		signalPaths = append(signalPaths, signalPath)
	}

	return signalPaths
}

func findCommonSignals(paths [][]*circuit.Signal) []*circuit.Signal {
	if len(paths) == 0 {
		return nil
	}

	// Count occurrences of each signal
	signalCount := make(map[*circuit.Signal]int)
	for _, path := range paths {
		// Use map to avoid counting duplicates within same path
		seen := make(map[*circuit.Signal]bool)
		for _, signal := range path {
			if !seen[signal] {
				signalCount[signal]++
				seen[signal] = true
			}
		}
	}

	// Find signals that appear in all paths
	common := make([]*circuit.Signal, 0)
	for signal, count := range signalCount {
		if count == len(paths) {
			common = append(common, signal)
		}
	}

	return common
}
