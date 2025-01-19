// example_circuits.go
package examples

import (
    "github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
)

// CreateC17Circuit creates the ISCAS-85 C17 benchmark circuit
func CreateC17Circuit() *circuit.Circuit {
    c := circuit.NewCircuit()

    // Create primary inputs
    in1 := circuit.NewSignal("1")
    in2 := circuit.NewSignal("2")
    in3 := circuit.NewSignal("3")
    in4 := circuit.NewSignal("4")
    in5 := circuit.NewSignal("5")
    
    // Mark as primary inputs
    in1.MarkAsPrimary()
    in2.MarkAsPrimary()
    in3.MarkAsPrimary()
    in4.MarkAsPrimary()
    in5.MarkAsPrimary()
    
    // Add to circuit
    c.AddPrimaryInput(in1)
    c.AddPrimaryInput(in2)
    c.AddPrimaryInput(in3)
    c.AddPrimaryInput(in4)
    c.AddPrimaryInput(in5)

    // Create internal signals
    n6 := circuit.NewSignal("6")
    n7 := circuit.NewSignal("7")
    n8 := circuit.NewSignal("8")
    n9 := circuit.NewSignal("9")
    out10 := circuit.NewSignal("10")
    out11 := circuit.NewSignal("11")

    // Mark outputs as primary
    out10.MarkAsPrimary()
    out11.MarkAsPrimary()
    
    // Add primary outputs to circuit
    c.AddPrimaryOutput(out10)
    c.AddPrimaryOutput(out11)

    // Create NAND gates (C17 uses only NAND gates)
    // We'll implement NAND using AND and NOT gates for simplicity
    
    // First level gates
    g1 := circuit.NewGate("g1", circuit.AND, []*circuit.Signal{in1, in2}, n6)
    g2 := circuit.NewGate("g2", circuit.AND, []*circuit.Signal{in3, in4}, n7)
    
    // Second level gates
    g3 := circuit.NewGate("g3", circuit.AND, []*circuit.Signal{n6, in3}, n8)
    g4 := circuit.NewGate("g4", circuit.AND, []*circuit.Signal{n7, in5}, n9)
    
    // Output gates
    g5 := circuit.NewGate("g5", circuit.AND, []*circuit.Signal{n8, n7}, out10)
    g6 := circuit.NewGate("g6", circuit.AND, []*circuit.Signal{n9, n8}, out11)

    // Add gates to circuit
    c.AddGate(g1)
    c.AddGate(g2)
    c.AddGate(g3)
    c.AddGate(g4)
    c.AddGate(g5)
    c.AddGate(g6)

    // Set up fanout connections
    in1.AddFanout(n6)
    in2.AddFanout(n6)
    in3.AddFanout(n7)
    in3.AddFanout(n8)
    in4.AddFanout(n7)
    in5.AddFanout(n9)
    n6.AddFanout(n8)
    n7.AddFanout(n9)
    n7.AddFanout(out10)
    n8.AddFanout(out10)
    n8.AddFanout(out11)
    n9.AddFanout(out11)

    // Identify bound and head lines
    c.IdentifyBoundAndHeadLines()

    return c
}

// CreateSimpleCircuit creates a simple circuit for testing
func CreateSimpleCircuit() *circuit.Circuit {
    c := circuit.NewCircuit()

    // Create signals
    in1 := circuit.NewSignal("in1")
    in2 := circuit.NewSignal("in2")
    in3 := circuit.NewSignal("in3")
    
    mid1 := circuit.NewSignal("mid1")
    mid2 := circuit.NewSignal("mid2")
    
    out := circuit.NewSignal("out")

    // Mark primary I/O
    in1.MarkAsPrimary()
    in2.MarkAsPrimary()
    in3.MarkAsPrimary()
    out.MarkAsPrimary()

    // Add to circuit
    c.AddPrimaryInput(in1)
    c.AddPrimaryInput(in2)
    c.AddPrimaryInput(in3)
    c.AddPrimaryOutput(out)

    // Create gates
    g1 := circuit.NewGate("g1", circuit.AND, []*circuit.Signal{in1, in2}, mid1)
    g2 := circuit.NewGate("g2", circuit.OR, []*circuit.Signal{in3, mid1}, mid2)
    g3 := circuit.NewGate("g3", circuit.NOT, []*circuit.Signal{mid2}, out)

    // Add gates to circuit
    c.AddGate(g1)
    c.AddGate(g2)
    c.AddGate(g3)

    // Set up fanout connections
    in1.AddFanout(mid1)
    in2.AddFanout(mid1)
    in3.AddFanout(mid2)
    mid1.AddFanout(mid2)
    mid2.AddFanout(out)

    // Identify bound and head lines
    c.IdentifyBoundAndHeadLines()

    return c
}