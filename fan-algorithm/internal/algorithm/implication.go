// implication.go
package algorithm

import (
    "github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
    "github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/utils"
)

// ImplicationResult represents the result of an implication
type ImplicationResult struct {
    Success      bool
    Implications []utils.Assignment
}

// Implication performs forward and backward implication from a given assignment
func Implication(c *circuit.Circuit, assignment utils.Assignment) ImplicationResult {
    result := ImplicationResult{
        Success:      true,
        Implications: make([]utils.Assignment, 0),
    }
    
    // Add initial assignment to implications
    result.Implications = append(result.Implications, assignment)
    
    // Keep track of processed signals to avoid loops
    processed := make(map[*circuit.Signal]bool)
    
    // Process queue for implications
    queue := []utils.Assignment{assignment}
    
    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]
        
        if processed[current.Signal] {
            continue
        }
        processed[current.Signal] = true
        
        // Forward implication
        forwardResult := forwardImplication(current.Signal, &queue)
        if !forwardResult.Success {
            result.Success = false
            return result
        }
        result.Implications = append(result.Implications, forwardResult.Implications...)
        
        // Backward implication
        backwardResult := backwardImplication(current.Signal, &queue)
        if !backwardResult.Success {
            result.Success = false
            return result
        }
        result.Implications = append(result.Implications, backwardResult.Implications...)
    }
    
    return result
}

// forwardImplication performs forward implication through fanout paths
func forwardImplication(signal *circuit.Signal, queue *[]utils.Assignment) ImplicationResult {
    result := ImplicationResult{
        Success:      true,
        Implications: make([]utils.Assignment, 0),
    }
    
    // Process each fanout
    for _, fanout := range signal.Fanouts {
        if fanout.FanIn == nil {
            continue
        }
        
        // Evaluate the gate
        gate := fanout.FanIn
        // oldValue := gate.Output.GetValue()
        if gate.Evaluate() {
            // If output changed, check consistency and add to queue
            newValue := gate.Output.GetValue()
            if !utils.CheckImplicationConsistency(gate.Output, newValue) {
                result.Success = false
                return result
            }
            
            *queue = append(*queue, utils.Assignment{
                Signal: gate.Output,
                Value:  newValue,
            })
            result.Implications = append(result.Implications, utils.Assignment{
                Signal: gate.Output,
                Value:  newValue,
            })
        }
    }
    
    return result
}

// backwardImplication performs backward implication through the driving gate
func backwardImplication(signal *circuit.Signal, queue *[]utils.Assignment) ImplicationResult {
    result := ImplicationResult{
        Success:      true,
        Implications: make([]utils.Assignment, 0),
    }

    // If signal has no driving gate, return
    if signal.FanIn == nil {
        return result
    }

    gate := signal.FanIn
    
    switch gate.Type {
    case circuit.AND:
        result = backwardImplicateAND(gate, signal, queue)
    case circuit.OR:
        result = backwardImplicateOR(gate, signal, queue)
    case circuit.NOT:
        result = backwardImplicateNOT(gate, signal, queue)
    }

    return result
}

// backwardImplicateAND handles backward implication for AND gates
func backwardImplicateAND(gate *circuit.Gate, signal *circuit.Signal, queue *[]utils.Assignment) ImplicationResult {
    result := ImplicationResult{
        Success:      true,
        Implications: make([]utils.Assignment, 0),
    }

    // If output is 1, all inputs must be 1
    if gate.Output.GetValue() == circuit.ONE {
        for _, input := range gate.Inputs {
            if input.IsUnknown() {
                if !utils.CheckImplicationConsistency(input, circuit.ONE) {
                    result.Success = false
                    return result
                }
                *queue = append(*queue, utils.Assignment{
                    Signal: input,
                    Value:  circuit.ONE,
                })
                result.Implications = append(result.Implications, utils.Assignment{
                    Signal: input,
                    Value:  circuit.ONE,
                })
            }
        }
    }

    // If output is 0 and all inputs except one are 1, the remaining input must be 0
    if gate.Output.GetValue() == circuit.ZERO {
        unknownInputs := utils.GetUnassignedInputs(gate)
        oneInputs := 0
        for _, input := range gate.Inputs {
            if input.GetValue() == circuit.ONE {
                oneInputs++
            }
        }

        if oneInputs == len(gate.Inputs)-1 && len(unknownInputs) == 1 {
            if !utils.CheckImplicationConsistency(unknownInputs[0], circuit.ZERO) {
                result.Success = false
                return result
            }
            *queue = append(*queue, utils.Assignment{
                Signal: unknownInputs[0],
                Value:  circuit.ZERO,
            })
            result.Implications = append(result.Implications, utils.Assignment{
                Signal: unknownInputs[0],
                Value:  circuit.ZERO,
            })
        }
    }

    return result
}

// backwardImplicateOR handles backward implication for OR gates
func backwardImplicateOR(gate *circuit.Gate, signal *circuit.Signal, queue *[]utils.Assignment) ImplicationResult {
    result := ImplicationResult{
        Success:      true,
        Implications: make([]utils.Assignment, 0),
    }

    // If output is 0, all inputs must be 0
    if gate.Output.GetValue() == circuit.ZERO {
        for _, input := range gate.Inputs {
            if input.IsUnknown() {
                if !utils.CheckImplicationConsistency(input, circuit.ZERO) {
                    result.Success = false
                    return result
                }
                *queue = append(*queue, utils.Assignment{
                    Signal: input,
                    Value:  circuit.ZERO,
                })
                result.Implications = append(result.Implications, utils.Assignment{
                    Signal: input,
                    Value:  circuit.ZERO,
                })
            }
        }
    }

    // If output is 1 and all inputs except one are 0, the remaining input must be 1
    if gate.Output.GetValue() == circuit.ONE {
        unknownInputs := utils.GetUnassignedInputs(gate)
        zeroInputs := 0
        for _, input := range gate.Inputs {
            if input.GetValue() == circuit.ZERO {
                zeroInputs++
            }
        }

        if zeroInputs == len(gate.Inputs)-1 && len(unknownInputs) == 1 {
            if !utils.CheckImplicationConsistency(unknownInputs[0], circuit.ONE) {
                result.Success = false
                return result
            }
            *queue = append(*queue, utils.Assignment{
                Signal: unknownInputs[0],
                Value:  circuit.ONE,
            })
            result.Implications = append(result.Implications, utils.Assignment{
                Signal: unknownInputs[0],
                Value:  circuit.ONE,
            })
        }
    }

    return result
}

// backwardImplicateNOT handles backward implication for NOT gates
func backwardImplicateNOT(gate *circuit.Gate, signal *circuit.Signal, queue *[]utils.Assignment) ImplicationResult {
    result := ImplicationResult{
        Success:      true,
        Implications: make([]utils.Assignment, 0),
    }

    input := gate.Inputs[0]
    if input.IsUnknown() {
        var impliedValue circuit.SignalValue
        switch gate.Output.GetValue() {
        case circuit.ZERO:
            impliedValue = circuit.ONE
        case circuit.ONE:
            impliedValue = circuit.ZERO
        case circuit.D:
            impliedValue = circuit.D_BAR
        case circuit.D_BAR:
            impliedValue = circuit.D
        default:
            return result
        }

        if !utils.CheckImplicationConsistency(input, impliedValue) {
            result.Success = false
            return result
        }

        *queue = append(*queue, utils.Assignment{
            Signal: input,
            Value:  impliedValue,
        })
        result.Implications = append(result.Implications, utils.Assignment{
            Signal: input,
            Value:  impliedValue,
        })
    }

    return result
}

// ImplicateSignal performs implication for a specific signal value assignment
func ImplicateSignal(c *circuit.Circuit, signal *circuit.Signal, value circuit.SignalValue) ImplicationResult {
    // Create initial assignment
    assignment := utils.Assignment{
        Signal: signal,
        Value:  value,
    }

    // Perform implication
    return Implication(c, assignment)
}