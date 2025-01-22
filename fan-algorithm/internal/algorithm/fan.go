package algorithm

import (
	"time"

	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/sensitization"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/pkg/types"
)

// FAN algorithm implementation with configuration and enhanced types
func FAN(c *circuit.Circuit, faultSite *circuit.Signal, faultValue circuit.SignalValue) *types.TestResult {
	result := types.NewTestResult()
	config := types.NewTestGenerationConfig()
	startTime := time.Now()

	// Initialize circuit state
	saveInitialState(c)
	faultSite.SetValue(faultValue)

	decisionTree := make([]*types.Decision, 0)

	for {
		// Check limits
		if result.Stats.Decisions >= config.MaxDecisions {
			result.Error = types.ErrMaxDecisions
			return result
		}
		if result.Stats.Backtracks >= config.MaxBacktracks {
			result.Error = types.ErrMaxBacktracks
			return result
		}
		if time.Since(startTime) > config.TimeLimit {
			result.Error = types.ErrTimeout
			return result
		}

		// Step 1: Implication
		implResult := Implication(c, types.Assignment{
			Signal:    faultSite,
			Value:     faultValue,
			Reason:    types.DECISION,
			Level:     result.CircuitState.DecisionLevel,
			TimeStamp: time.Now(),
		})

		if !implResult.Success {
			result.Stats.Update(false, true)
			if !backtrack(&decisionTree, c, result) {
				return result
			}
			continue
		}
		result.Implications = make([]types.Assignment, len(implResult.Implications))
		for i, impl := range implResult.Implications {
			result.Implications[i] = types.Assignment{
				Signal:    impl.Signal,
				Value:     impl.Value,
				Reason:    types.IMPLICATION,
				Level:     result.CircuitState.DecisionLevel,
				TimeStamp: time.Now(),
			}
		}

		// Step 2: Find D-frontier
		dFrontier := findDFrontier(c)
		result.DFrontier = convertToDFrontierGates(dFrontier)

		// Step 3: Check if test is complete
		if isTestComplete(c, dFrontier) {
			result.Success = true
			saveTestPattern(c, result)
			result.Stats.ExecutionTime = time.Since(startTime)
			return result
		}

		// Step 4: Handle unique sensitization if enabled
		if config.UseUniqueSensitization {
			if uniquePaths := findUniqueSensitizationPaths(c, dFrontier); len(uniquePaths) > 0 {
				if !handleUniqueSensitization(c, uniquePaths, &decisionTree, result) {
					result.Stats.Update(false, true)
					if !backtrack(&decisionTree, c, result) {
						return result
					}
					continue
				}
			}
		}

		// Step 5: Multiple backtrace
		objectives := createBacktraceObjectives(c, dFrontier)
		if len(objectives) == 0 {
			result.Stats.Update(false, true)
			if !backtrack(&decisionTree, c, result) {
				return result
			}
			continue
		}

		backtraceResult := MultipleBacktrace(objectives, c, config)
		if !handleBacktraceResult(backtraceResult, c, &decisionTree, result) {
			result.Stats.Update(false, true)
			if !backtrack(&decisionTree, c, result) {
				return result
			}
		}
		result.Stats.Update(true, false)
	}
}

// State management functions
func saveInitialState(c *circuit.Circuit) *types.CircuitState {
	state := types.NewCircuitState()
	for _, signal := range c.Signals {
		state.SignalValues[signal] = signal.GetValue()
	}
	return state
}

func saveTestPattern(c *circuit.Circuit, result *types.TestResult) {
	for _, signal := range c.PrimaryInputs {
		result.TestPattern[signal] = signal.GetValue()
	}
}

// Backtracking support
func backtrack(decisionTree *[]*types.Decision, c *circuit.Circuit, result *types.TestResult) bool {
	if len(*decisionTree) == 0 {
		return false
	}

	lastIdx := len(*decisionTree) - 1
	lastDecision := (*decisionTree)[lastIdx]

	if !lastDecision.Alternative {
		// Try alternative value
		lastDecision.Alternative = true
		lastDecision.Value = getOppositeValue(lastDecision.Value)
		lastDecision.TimeStamp = time.Now()

		// Reset circuit state and replay decisions
		resetCircuitState(c, *decisionTree, lastIdx)
		lastDecision.Signal.SetValue(lastDecision.Value)
		return true
	}

	// Remove last decision and try parent
	*decisionTree = (*decisionTree)[:lastIdx]
	return backtrack(decisionTree, c, result)
}

// D-frontier handling
func findDFrontier(c *circuit.Circuit) []*circuit.Gate {
	frontier := make([]*circuit.Gate, 0)
	for _, gate := range c.Gates {
		if isDFrontierGate(gate) {
			frontier = append(frontier, gate)
		}
	}
	return frontier
}

func isDFrontierGate(gate *circuit.Gate) bool {
	hasFaultyInput := false
	for _, input := range gate.Inputs {
		if input.IsFaulty() {
			hasFaultyInput = true
			break
		}
	}
	return hasFaultyInput && !gate.Output.IsFaulty()
}

// Type conversion helpers
func convertToDFrontierGates(gates []*circuit.Gate) []types.DFrontierGate {
	result := make([]types.DFrontierGate, len(gates))
	for i, gate := range gates {
		result[i] = types.DFrontierGate{
			Gate:           gate,
			FaultyInput:    findFaultyInput(gate),
			Priority:       calculateGatePriority(gate),
			BlockingInputs: findBlockingInputs(gate),
		}
	}
	return result
}

func findFaultyInput(gate *circuit.Gate) *circuit.Signal {
	for _, input := range gate.Inputs {
		if input.IsFaulty() {
			return input
		}
	}
	return nil
}

func findBlockingInputs(gate *circuit.Gate) []*circuit.Signal {
	blocking := make([]*circuit.Signal, 0)
	for _, input := range gate.Inputs {
		if !input.IsFaulty() && gate.IsControllingValue(input.GetValue()) {
			blocking = append(blocking, input)
		}
	}
	return blocking
}

// Change this function
func calculateGatePriority(gate *circuit.Gate) int {
	priority := 0
	// Get circuit instance from gate
	var c *circuit.Circuit
	if gate.Output != nil && gate.Output.FanIn != nil {
		// Get the circuit instance from any signal/gate
		// Need to add Circuit field to Gate struct
		c = gate.Circuit
	}

	if c != nil {
		pFinder := sensitization.NewPathFinder(c)
		paths := pFinder.FindUniqueSensitizationPaths([]*circuit.Gate{gate})
		if len(paths) > 0 {
			priority += 10 // Higher priority for gates with unique sensitization paths
		}
	}
	return priority
}

// Test completion check
func isTestComplete(c *circuit.Circuit, dFrontier []*circuit.Gate) bool {
	// Test is complete if D-frontier is empty and fault has propagated to output
	if len(dFrontier) > 0 {
		return false
	}

	for _, output := range c.PrimaryOutputs {
		if output.IsFaulty() {
			return true
		}
	}
	return false
}

// Objective creation
func createBacktraceObjectives(c *circuit.Circuit, dFrontier []*circuit.Gate) []*types.BacktraceObjective {
	objectives := make([]*types.BacktraceObjective, 0)

	for _, gate := range dFrontier {
		obj := types.CreateObjective(gate.Output, types.PROPAGATE, circuit.ONE)
		obj.Priority = calculateObjectivePriority(gate)
		objectives = append(objectives, obj)
	}

	return objectives
}

// State reset support
func resetCircuitState(c *circuit.Circuit, decisions []*types.Decision, upToIndex int) {
	// Reset all non-primary signals to X
	for _, signal := range c.Signals {
		if !signal.IsPrimary {
			signal.SetValue(circuit.X)
		}
	}

	// Replay decisions up to index
	for i := 0; i <= upToIndex; i++ {
		decisions[i].Signal.SetValue(decisions[i].Value)
	}
}

// Helper functions
func getOppositeValue(value circuit.SignalValue) circuit.SignalValue {
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

func calculateObjectivePriority(gate *circuit.Gate) int {
	priority := 0
	if gate.Output.IsHead {
		priority += 5 // Higher priority for head lines
	}
	if len(gate.Output.Fanouts) > 1 {
		priority -= 3 // Lower priority for fanout points
	}
	return priority
}

func handleUniqueSensitization(c *circuit.Circuit, paths []*circuit.Signal,
	decisionTree *[]*types.Decision, result *types.TestResult) bool {

	for _, signal := range paths {
		if signal.FanIn == nil {
			continue
		}

		value := signal.FanIn.GetNonControllingValue()
		decision := &types.Decision{
			Signal:    signal,
			Value:     value,
			Level:     result.CircuitState.DecisionLevel + 1,
			TimeStamp: time.Now(),
		}

		*decisionTree = append(*decisionTree, decision)
		signal.SetValue(value)

		result.CircuitState.DecisionLevel++
		result.Stats.Decisions++
	}

	return true
}

func handleBacktraceResult(backtraceResult *BacktraceResult, c *circuit.Circuit,
	decisionTree *[]*types.Decision, result *types.TestResult) bool {

	if len(backtraceResult.FinalObjectives) == 0 {
		return false
	}

	obj := backtraceResult.FinalObjectives[0]
	decision := &types.Decision{
		Signal:    obj.Signal,
		Value:     obj.Value,
		Level:     result.CircuitState.DecisionLevel + 1,
		TimeStamp: time.Now(),
	}

	*decisionTree = append(*decisionTree, decision)
	obj.Signal.SetValue(obj.Value)

	result.CircuitState.DecisionLevel++
	result.Stats.Decisions++

	return true
}

func findUniqueSensitizationPaths(c *circuit.Circuit, dFrontier []*circuit.Gate) []*circuit.Signal {
	pFinder := sensitization.NewPathFinder(c)
	paths := pFinder.FindUniqueSensitizationPaths(dFrontier)
	return pFinder.GetMandatorySignals(paths)
}
