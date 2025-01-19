// fan.go
package algorithm

import (
    "github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/utils"
)

// FanResult represents the result of FAN algorithm
type FanResult struct {
    TestPattern map[*circuit.Signal]circuit.SignalValue // Final test pattern
    Success     bool                                    // Whether test generation succeeded
    DFrontier   []*circuit.Gate                        // Final D-frontier gates
}

// DecisionNode represents a node in the decision tree
type DecisionNode struct {
    Line        *circuit.Signal            // Signal line where decision was made
    Value       circuit.SignalValue        // Assigned value
    Alternative bool                       // Whether alternative value has been tried
    State       map[*circuit.Signal]circuit.SignalValue // Circuit state at this point
}

// FAN implements the FAN algorithm for test pattern generation
func FAN(c *circuit.Circuit, faultSite *circuit.Signal, faultValue circuit.SignalValue) *FanResult {
    result := &FanResult{
        TestPattern: make(map[*circuit.Signal]circuit.SignalValue),
        Success:     false,
        DFrontier:   make([]*circuit.Gate, 0),
    }

    decisionTree := make([]*DecisionNode, 0)
    
    saveInitialState(c)
    faultSite.SetValue(faultValue)
    
    for {
        // Perform implication for all current assignments
        initialAssignments := getCurrentAssignments(c)
        implicationFailed := false
        
        for _, assignment := range initialAssignments {
            implicationResult := Implication(c, assignment)
            if !implicationResult.Success {
                implicationFailed = true
                break
            }
        }
        
        if implicationFailed {
            if !backtrack(&decisionTree, c) {
                return result
            }
            continue
        }

        dFrontier := findDFrontier(c)
        result.DFrontier = dFrontier

        if isTestComplete(c, dFrontier) {
            result.Success = true
            saveTestPattern(c, result)
            return result
        }

        objectives := getObjectives(c, dFrontier)
        if len(objectives) == 0 {
            if !backtrack(&decisionTree, c) {
                return result
            }
            continue
        }

        backtraceResult := MultipleBacktrace(objectives, c)
        if len(backtraceResult.FinalObjectives) == 0 {
            if !backtrack(&decisionTree, c) {
                return result
            }
            continue
        }

        success := applyFinalObjective(backtraceResult, c, &decisionTree)
        if !success {
            if !backtrack(&decisionTree, c) {
                return result
            }
        }
    }
}

// getCurrentAssignments returns all non-X value assignments in the circuit
func getCurrentAssignments(c *circuit.Circuit) []utils.Assignment {
    assignments := make([]utils.Assignment, 0)
    for _, signal := range c.Signals {
        if !signal.IsUnknown() {
            assignments = append(assignments, utils.Assignment{
                Signal: signal,
                Value:  signal.GetValue(),
            })
        }
    }
    return assignments
}

// findDFrontier finds all gates in the D-frontier
func findDFrontier(c *circuit.Circuit) []*circuit.Gate {
    frontier := make([]*circuit.Gate, 0)
    for _, gate := range c.Gates {
        if isDFrontierGate(gate) {
            frontier = append(frontier, gate)
        }
    }
    return frontier
}

// isDFrontierGate checks if a gate is in the D-frontier
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

// isTestComplete checks if test generation is complete
func isTestComplete(c *circuit.Circuit, dFrontier []*circuit.Gate) bool {
    // Check if fault effect propagated to any primary output
    for _, output := range c.PrimaryOutputs {
        if output.IsFaulty() {
            return true
        }
    }
    return false
}

// getObjectives determines current objectives based on circuit state
func getObjectives(c *circuit.Circuit, dFrontier []*circuit.Gate) []*Objective {
    objectives := make([]*Objective, 0)

    // If D-frontier exists, create objectives for fault propagation
    if len(dFrontier) > 0 {
        gate := dFrontier[0] // Choose first gate in D-frontier
        // Create objectives for sensitizing the gate
        for _, input := range gate.Inputs {
            if !input.IsFaulty() {
                value := gate.GetNonControllingValue()
                objectives = append(objectives, CreateInitialObjective(input, value))
            }
        }
    }

    // Add objectives for any unjustified lines
    for _, signal := range c.Signals {
        if !signal.State.IsStable && !signal.IsUnknown() {
            objectives = append(objectives, CreateInitialObjective(signal, signal.GetValue()))
        }
    }

    return objectives
}

// applyFinalObjective applies a chosen final objective
func applyFinalObjective(result *BacktraceResult, c *circuit.Circuit, decisionTree *[]*DecisionNode) bool {
    if len(result.FinalObjectives) == 0 {
        return false
    }

    // Choose first objective (already sorted by controllability)
    obj := result.FinalObjectives[0]
    value := GetFinalValue(obj)

    // Create new decision node
    node := &DecisionNode{
        Line:        obj.Line,
        Value:       value,
        Alternative: false,
        State:       saveCircuitState(c),
    }
    *decisionTree = append(*decisionTree, node)

    // Apply the value
    obj.Line.SetValue(value)
    return true
}

// backtrack performs backtracking in the decision tree
func backtrack(decisionTree *[]*DecisionNode, c *circuit.Circuit) bool {
    for len(*decisionTree) > 0 {
        lastNode := (*decisionTree)[len(*decisionTree)-1]
        
        if !lastNode.Alternative {
            // Try alternative value
            lastNode.Alternative = true
            restoreCircuitState(c, lastNode.State)
            
            // Apply alternative value
            alternativeValue := getAlternativeValue(lastNode.Value)
            lastNode.Line.SetValue(alternativeValue)
            return true
        }
        
        // Remove last node and continue backtracking
        *decisionTree = (*decisionTree)[:len(*decisionTree)-1]
    }
    return false
}

// Helper functions for state management
func saveCircuitState(c *circuit.Circuit) map[*circuit.Signal]circuit.SignalValue {
    state := make(map[*circuit.Signal]circuit.SignalValue)
    for _, signal := range c.Signals {
        state[signal] = signal.GetValue()
    }
    return state
}

func restoreCircuitState(c *circuit.Circuit, state map[*circuit.Signal]circuit.SignalValue) {
    for signal, value := range state {
        signal.SetValue(value)
    }
}

func saveInitialState(c *circuit.Circuit) {
    for _, signal := range c.Signals {
        if !signal.IsPrimary {
            signal.SetValue(circuit.X)
        }
    }
}

func saveTestPattern(c *circuit.Circuit, result *FanResult) {
    for _, input := range c.PrimaryInputs {
        result.TestPattern[input] = input.GetValue()
    }
}

func getAlternativeValue(value circuit.SignalValue) circuit.SignalValue {
    if value == circuit.ONE {
        return circuit.ZERO
    }
    return circuit.ONE
}