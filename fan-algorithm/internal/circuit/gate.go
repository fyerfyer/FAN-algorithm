// gate.go
package circuit

// GateType represents the type of logic gate
type GateType int

const (
    AND GateType = iota // AND gate
    OR                  // OR gate
    NOT                 // NOT gate
)

// Gate represents a logic gate in the circuit
type Gate struct {
    ID      string    // Unique identifier for the gate
    Type    GateType  // Type of the gate (AND, OR, NOT)
    Inputs  []*Signal // Input signals
    Output  *Signal   // Output signal
}

// NewGate creates a new gate with the specified type and signals
func NewGate(id string, gateType GateType, inputs []*Signal, output *Signal) *Gate {
    return &Gate{
        ID:     id,
        Type:   gateType,
        Inputs: inputs,
        Output: output,
    }
}

// Evaluate computes the output value based on input values
// Returns true if the output value changes
func (g *Gate) Evaluate() bool {
    oldValue := g.Output.GetValue()
    
    switch g.Type {
    case AND:
        g.Output.SetValue(g.evaluateAND())
    case OR:
        g.Output.SetValue(g.evaluateOR())
    case NOT:
        g.Output.SetValue(g.evaluateNOT())
    }
    
    return oldValue != g.Output.GetValue()
}

// evaluateAND implements AND gate logic including D-algorithm values
func (g *Gate) evaluateAND() SignalValue {
    hasX := false
    hasD := false
    // hasDBar := false
    
    // Check for any 0 inputs first (dominant value for AND)
    for _, input := range g.Inputs {
        if input.GetValue() == ZERO {
            return ZERO
        }
        if input.GetValue() == D_BAR {
            return ZERO
        }
        if input.GetValue() == X {
            hasX = true
        }
        if input.GetValue() == D {
            hasD = true
        }
    }
    
    // If we have any X and no 0s, result is X
    if hasX {
        return X
    }
    
    // If all inputs are 1 or D
    if hasD {
        return D
    }
    
    // If all inputs are 1
    return ONE
}

// evaluateOR implements OR gate logic including D-algorithm values
func (g *Gate) evaluateOR() SignalValue {
    hasX := false
    // hasD := false
    hasDBar := false
    
    // Check for any 1 inputs first (dominant value for OR)
    for _, input := range g.Inputs {
        if input.GetValue() == ONE {
            return ONE
        }
        if input.GetValue() == D {
            return ONE
        }
        if input.GetValue() == X {
            hasX = true
        }
        if input.GetValue() == D_BAR {
            hasDBar = true
        }
    }
    
    // If we have any X and no 1s, result is X
    if hasX {
        return X
    }
    
    // If all inputs are 0 or D'
    if hasDBar {
        return D_BAR
    }
    
    // If all inputs are 0
    return ZERO
}

// evaluateNOT implements NOT gate logic including D-algorithm values
func (g *Gate) evaluateNOT() SignalValue {
    input := g.Inputs[0].GetValue()
    switch input {
    case ZERO:
        return ONE
    case ONE:
        return ZERO
    case D:
        return D_BAR
    case D_BAR:
        return D
    case X:
        return X
    default:
        return X
    }
}

// IsControllingValue checks if the given value is a controlling value for the gate
func (g *Gate) IsControllingValue(value SignalValue) bool {
    switch g.Type {
    case AND:
        return value == ZERO || value == D_BAR
    case OR:
        return value == ONE || value == D
    default: // NOT gate has no controlling value
        return false
    }
}

// GetNonControllingValue returns the non-controlling value for the gate
func (g *Gate) GetNonControllingValue() SignalValue {
    switch g.Type {
    case AND:
        return ONE
    case OR:
        return ZERO
    default: // NOT gate
        return X
    }
}
