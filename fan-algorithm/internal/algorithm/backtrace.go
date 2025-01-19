// backtrace.go
package algorithm

import (
    "sort"
    
    "github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
)

// Objective represents a backtrace objective with line and required counts
type Objective struct {
    Line *circuit.Signal // Target signal line
    N0   int            // Number of times value 0 is required
    N1   int            // Number of times value 1 is required
}

// BacktraceResult represents the result of multiple backtrace
type BacktraceResult struct {
    FinalObjectives []*Objective    // Final objectives at head lines
    HeadLines       []*circuit.Signal // Involved head lines
}

// NewObjective creates a new objective
func NewObjective(line *circuit.Signal, n0, n1 int) *Objective {
    return &Objective{
        Line: line,
        N0:   n0,
        N1:   n1,
    }
}

// MultipleBacktrace performs multiple backtrace from initial objectives
func MultipleBacktrace(initialObjectives []*Objective, c *circuit.Circuit) *BacktraceResult {
    result := &BacktraceResult{
        FinalObjectives: make([]*Objective, 0),
        HeadLines:       make([]*circuit.Signal, 0),
    }
    
    currentObjectives := initialObjectives
    processedLines := make(map[string]bool)
    
    // Continue until all objectives reach head lines
    for len(currentObjectives) > 0 {
        nextObjectives := make([]*Objective, 0)
        
        for _, obj := range currentObjectives {
            // Skip if already processed
            if processedLines[obj.Line.ID] {
                continue
            }
            processedLines[obj.Line.ID] = true
            
            // If current line is a head line, add to final objectives
            if obj.Line.IsHead {
                result.FinalObjectives = append(result.FinalObjectives, obj)
                if !contains(result.HeadLines, obj.Line) {
                    result.HeadLines = append(result.HeadLines, obj.Line)
                }
                continue
            }
            
            // If line has fanin gate, process the gate
            if obj.Line.FanIn != nil {
                newObjs := backtraceGate(obj.Line.FanIn, obj)
                nextObjectives = append(nextObjectives, newObjs...)
            }
        }
        
        currentObjectives = nextObjectives
    }
    
    // Sort final objectives by controllability
    sortObjectivesByControllability(result.FinalObjectives)
    
    return result
}

// backtraceGate processes backtrace through a gate
func backtraceGate(gate *circuit.Gate, objective *Objective) []*Objective {
    switch gate.Type {
    case circuit.AND:
        return backtraceANDGate(gate, objective)
    case circuit.OR:
        return backtraceORGate(gate, objective)
    case circuit.NOT:
        return backtraceNOTGate(gate, objective)
    default:
        return nil
    }
}

// backtraceANDGate handles backtrace through AND gate
func backtraceANDGate(gate *circuit.Gate, objective *Objective) []*Objective {
    results := make([]*Objective, 0)
    
    // For setting output to 0 (controlling value)
    if objective.N0 > 0 {
        easiestInput := findEasiestToControl(gate, circuit.ZERO)
        results = append(results, NewObjective(easiestInput, objective.N0, 0))
    }
    
    // For setting output to 1 (all inputs must be 1)
    if objective.N1 > 0 {
        for _, input := range gate.Inputs {
            results = append(results, NewObjective(input, 0, objective.N1))
        }
    }
    
    return results
}

// backtraceORGate handles backtrace through OR gate
func backtraceORGate(gate *circuit.Gate, objective *Objective) []*Objective {
    results := make([]*Objective, 0)
    
    // For setting output to 1 (controlling value)
    if objective.N1 > 0 {
        easiestInput := findEasiestToControl(gate, circuit.ONE)
        results = append(results, NewObjective(easiestInput, 0, objective.N1))
    }
    
    // For setting output to 0 (all inputs must be 0)
    if objective.N0 > 0 {
        for _, input := range gate.Inputs {
            results = append(results, NewObjective(input, objective.N0, 0))
        }
    }
    
    return results
}

// backtraceNOTGate handles backtrace through NOT gate
func backtraceNOTGate(gate *circuit.Gate, objective *Objective) []*Objective {
    // For NOT gate, swap N0 and N1
    return []*Objective{
        NewObjective(gate.Inputs[0], objective.N1, objective.N0),
    }
}

// findEasiestToControl finds the input that is easiest to control
func findEasiestToControl(gate *circuit.Gate, value circuit.SignalValue) *circuit.Signal {
    // Simple implementation: return first input
    // Could be enhanced with controllability measures
    return gate.Inputs[0]
}

// contains checks if a signal is in a slice
func contains(signals []*circuit.Signal, signal *circuit.Signal) bool {
    for _, s := range signals {
        if s.ID == signal.ID {
            return true
        }
    }
    return false
}

// sortObjectivesByControllability sorts objectives based on controllability
func sortObjectivesByControllability(objectives []*Objective) {
    sort.Slice(objectives, func(i, j int) bool {
        // Simple implementation: sort by total required assignments
        totalI := objectives[i].N0 + objectives[i].N1
        totalJ := objectives[j].N0 + objectives[j].N1
        return totalI < totalJ
    })
}

// CreateInitialObjective creates an initial objective for value assignment
func CreateInitialObjective(line *circuit.Signal, value circuit.SignalValue) *Objective {
    if value == circuit.ZERO {
        return NewObjective(line, 1, 0)
    }
    return NewObjective(line, 0, 1)
}

// GetFinalValue determines the final value for a head line based on objective
func GetFinalValue(obj *Objective) circuit.SignalValue {
    if obj.N0 > obj.N1 {
        return circuit.ZERO
    }
    if obj.N1 > obj.N0 {
        return circuit.ONE
    }
    // If equal, choose arbitrarily (could be enhanced with better heuristics)
    return circuit.ZERO
}