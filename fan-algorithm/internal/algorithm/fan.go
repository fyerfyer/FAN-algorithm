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
	decisionTree := make([]*types.Decision, 0)
	config := types.NewTestGenerationConfig() // Add config

	// Initialize circuit
	resetCircuit(c)

	// Set fault site value
	faultSite.Value = circuit.D
	if faultValue == circuit.ONE {
		faultSite.Value = circuit.D_BAR
	}

	for {
		// Forward implication
		implResult := performImplication(c, &decisionTree, result)
		if !implResult {
			if !backtrack(&decisionTree, c, result) {
				break
			}
			continue
		}

		// Find D-frontier
		dFrontier := findDFrontier(c)
		result.DFrontier = convertToDFrontierGates(dFrontier) // Convert type

		// Check for completion
		if isTestComplete(c, dFrontier) {
			result.Success = true
			saveTestPattern(c, result)
			break
		}

		// Multiple backtrace when D-frontier exists
		objectives := createObjectives(c, dFrontier)
		if len(objectives) == 0 {
			if !backtrack(&decisionTree, c, result) {
				break
			}
			continue
		}

		backtraceResult := MultipleBacktrace(objectives, c, config)
		if !handleBacktraceResult(backtraceResult, c, &decisionTree, result) {
			if !backtrack(&decisionTree, c, result) {
				break
			}
		}
	}

	return result
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

// Update isDFrontierGate function
func isDFrontierGate(gate *circuit.Gate) bool {
	hasFaultyInput := false
	hasUnassignedInput := false

	for _, input := range gate.Inputs {
		if input.Value == circuit.D || input.Value == circuit.D_BAR {
			hasFaultyInput = true
		}
		if input.Value == circuit.X {
			hasUnassignedInput = true
		}
	}

	return hasFaultyInput && hasUnassignedInput && !gate.Output.IsFaulty()
}

// Type conversion helpers
func convertToDFrontierGates(gates []*circuit.Gate) []types.DFrontierGate {
	result := make([]types.DFrontierGate, len(gates))
	for i, gate := range gates {
		result[i] = types.DFrontierGate{
			Gate:        gate,
			FaultyInput: findFaultyInput(gate),
			Priority:    calculateGatePriority(gate),
		}
	}
	return result
}

func findFaultyInput(gate *circuit.Gate) *circuit.Signal {
	for _, input := range gate.Inputs {
		if input.Value == circuit.D || input.Value == circuit.D_BAR {
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
	// Test is complete if D or D' reaches any primary output
	for _, output := range c.PrimaryOutputs {
		if output.Value == circuit.D || output.Value == circuit.D_BAR {
			return true
		}
	}
	// Or if D-frontier is empty and all necessary assignments made
	return len(dFrontier) == 0 && allNecessaryAssignmentsMade(c)
}

// Objective creation
func createBacktraceObjectives(c *circuit.Circuit, dFrontier []types.DFrontierGate) []*types.BacktraceObjective {
	objectives := make([]*types.BacktraceObjective, 0)

	for _, dfGate := range dFrontier {
		obj := &types.BacktraceObjective{
			Signal:    dfGate.Gate.Output,
			Value:     circuit.ONE,
			Priority:  calculateObjectivePriority(dfGate.Gate),
			ZeroCount: 0,
			OneCount:  1,
		}
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

// Update handleBacktraceResult function
func handleBacktraceResult(backtraceResult *BacktraceResult, c *circuit.Circuit,
	decisionTree *[]*types.Decision, result *types.TestResult) bool {

	if len(backtraceResult.FinalObjectives) == 0 {
		return false
	}

	// Try each objective until one succeeds
	for _, obj := range backtraceResult.FinalObjectives {
		decision := &types.Decision{
			Signal:    obj.Signal,
			Value:     obj.Value,
			Level:     result.CircuitState.DecisionLevel + 1,
			TimeStamp: time.Now(),
		}

		// Apply decision
		obj.Signal.SetValue(obj.Value)

		// Check if decision leads to consistency
		implResult := Implication(c, types.Assignment{
			Signal: obj.Signal,
			Value:  obj.Value,
			Reason: types.DECISION,
			Level:  result.CircuitState.DecisionLevel + 1,
		})

		if implResult.Success {
			*decisionTree = append(*decisionTree, decision)
			result.CircuitState.DecisionLevel++
			result.Stats.Decisions++
			return true
		}

		// Reset if implication failed
		obj.Signal.SetValue(circuit.X)
	}

	return false
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

func findUniqueSensitizationPaths(c *circuit.Circuit, dFrontier []*circuit.Gate) []*circuit.Signal {
	pFinder := sensitization.NewPathFinder(c)
	paths := pFinder.FindUniqueSensitizationPaths(dFrontier)
	return pFinder.GetMandatorySignals(paths)
}

// Add performImplication function
func performImplication(c *circuit.Circuit, decisionTree *[]*types.Decision, result *types.TestResult) bool {
	changed := true
	for changed {
		changed = false
		for _, gate := range c.Gates {
			oldValue := gate.Output.Value
			newValue := evaluateGate(gate)

			if oldValue != newValue {
				if !gate.Output.IsCompatible(newValue) {
					return false
				}
				gate.Output.Value = newValue
				changed = true
			}
		}
	}
	return true
}

func evaluateGate(gate *circuit.Gate) circuit.SignalValue {
	switch gate.Type {
	case circuit.AND:
		return evaluateANDGate(gate)
	case circuit.OR:
		return evaluateORGate(gate)
	case circuit.NOT:
		return evaluateNOTGate(gate)
	default:
		return circuit.X
	}
}

func evaluateANDGate(gate *circuit.Gate) circuit.SignalValue {
	hasX := false
	hasZero := false
	hasD := false
	hasDBARR := false

	for _, input := range gate.Inputs {
		switch input.Value {
		case circuit.X:
			hasX = true
		case circuit.ZERO:
			hasZero = true
		case circuit.D:
			hasD = true
		case circuit.D_BAR:
			hasDBARR = true
		}
	}

	if hasZero {
		return circuit.ZERO
	}
	if hasX {
		return circuit.X
	}
	if hasD && hasDBARR {
		return circuit.ZERO
	}
	if hasD {
		return circuit.D
	}
	if hasDBARR {
		return circuit.D_BAR
	}
	return circuit.ONE
}

func evaluateORGate(gate *circuit.Gate) circuit.SignalValue {
	hasX := false
	hasOne := false
	hasD := false
	hasDBARR := false

	for _, input := range gate.Inputs {
		switch input.Value {
		case circuit.X:
			hasX = true
		case circuit.ONE:
			hasOne = true
		case circuit.D:
			hasD = true
		case circuit.D_BAR:
			hasDBARR = true
		}
	}

	if hasOne {
		return circuit.ONE
	}
	if hasX {
		return circuit.X
	}
	if hasD && hasDBARR {
		return circuit.ONE
	}
	if hasD {
		return circuit.D
	}
	if hasDBARR {
		return circuit.D_BAR
	}
	return circuit.ZERO
}

func evaluateNOTGate(gate *circuit.Gate) circuit.SignalValue {
	input := gate.Inputs[0]
	switch input.Value {
	case circuit.X:
		return circuit.X
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

// Add helper function to convert D-frontier types
func convertToCircuitGates(dFrontier []types.DFrontierGate) []*circuit.Gate {
	gates := make([]*circuit.Gate, len(dFrontier))
	for i, df := range dFrontier {
		gates[i] = df.Gate
	}
	return gates
}

func resetCircuit(c *circuit.Circuit) {
	for _, signal := range c.Signals {
		if !signal.IsPrimary {
			signal.Value = circuit.X
		}
	}
}

func createObjectives(c *circuit.Circuit, dFrontier []*circuit.Gate) []*types.BacktraceObjective {
	objectives := make([]*types.BacktraceObjective, 0)
	for _, gate := range dFrontier {
		obj := &types.BacktraceObjective{
			Signal:    gate.Output,
			Value:     circuit.ONE, // Non-controlling value for propagation
			Priority:  10,
			ZeroCount: 0,
			OneCount:  1,
		}
		objectives = append(objectives, obj)
	}
	return objectives
}

func allNecessaryAssignmentsMade(c *circuit.Circuit) bool {
	for _, signal := range c.Signals {
		if signal.Value == circuit.X && !signal.IsPrimary {
			return false
		}
	}
	return true
}
