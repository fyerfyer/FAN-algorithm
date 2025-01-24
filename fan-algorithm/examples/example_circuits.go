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

	g1 := circuit.NewGate("g1", circuit.AND, []*circuit.Signal{in1, in2}, n6, c)
	g2 := circuit.NewGate("g2", circuit.AND, []*circuit.Signal{in3, in4}, n7, c)
	g3 := circuit.NewGate("g3", circuit.AND, []*circuit.Signal{n6, in3}, n8, c)
	g4 := circuit.NewGate("g4", circuit.AND, []*circuit.Signal{n7, in5}, n9, c)
	g5 := circuit.NewGate("g5", circuit.AND, []*circuit.Signal{n8, n7}, out10, c)
	g6 := circuit.NewGate("g6", circuit.AND, []*circuit.Signal{n9, n8}, out11, c)

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

	// Create signals with clear IDs
	in1 := circuit.NewSignal("in1")
	in2 := circuit.NewSignal("in2")
	n1 := circuit.NewSignal("n1")
	out1 := circuit.NewSignal("out1")

	// Mark I/O
	in1.MarkAsPrimary()
	in2.MarkAsPrimary()
	out1.MarkAsPrimary()

	// Add to circuit
	c.AddPrimaryInput(in1)
	c.AddPrimaryInput(in2)
	c.AddPrimaryOutput(out1)

	// Create simple test path: AND -> NOT
	g1 := circuit.NewGate("g1", circuit.AND, []*circuit.Signal{in1, in2}, n1, c)
	g2 := circuit.NewGate("g2", circuit.NOT, []*circuit.Signal{n1}, out1, c)

	c.AddGate(g1)
	c.AddGate(g2)

	// Set up connections
	n1.FanIn = g1
	n1.AddFanout(out1)
	out1.FanIn = g2

	c.Signals = []*circuit.Signal{in1, in2, n1, out1}

	return c
}

// Add a new example circuit specifically for FAN algorithm features
func CreateFanTestCircuit() *circuit.Circuit {
	c := circuit.NewCircuit()

	// Create signals with unique sensitization path structure
	in1 := circuit.NewSignal("in1")
	in2 := circuit.NewSignal("in2")
	in3 := circuit.NewSignal("in3")
	n1 := circuit.NewSignal("n1")
	n2 := circuit.NewSignal("n2") // Signal requiring unique sensitization
	n3 := circuit.NewSignal("n3")
	out1 := circuit.NewSignal("out1")
	out2 := circuit.NewSignal("out2")

	// Set primary I/O
	in1.MarkAsPrimary()
	in2.MarkAsPrimary()
	in3.MarkAsPrimary()
	out1.MarkAsPrimary()
	out2.MarkAsPrimary()

	c.AddPrimaryInput(in1)
	c.AddPrimaryInput(in2)
	c.AddPrimaryInput(in3)
	c.AddPrimaryOutput(out1)
	c.AddPrimaryOutput(out2)

	// Create gates forming unique sensitization paths
	g1 := circuit.NewGate("g1", circuit.AND, []*circuit.Signal{in1, in2}, n1, c)
	g2 := circuit.NewGate("g2", circuit.OR, []*circuit.Signal{n1, in3}, n2, c)
	g3 := circuit.NewGate("g3", circuit.AND, []*circuit.Signal{n2}, n3, c)
	g4 := circuit.NewGate("g4", circuit.OR, []*circuit.Signal{n3}, out1, c)
	g5 := circuit.NewGate("g5", circuit.AND, []*circuit.Signal{n2}, out2, c)

	c.Gates = append(c.Gates, g1, g2, g3, g4, g5)
	c.Signals = append(c.Signals, in1, in2, in3, n1, n2, n3, out1, out2)

	// Create mandatory paths through n2
	n2.AddFanout(n3)
	n2.AddFanout(out2)
	n3.AddFanout(out1)

	c.IdentifyBoundAndHeadLines()
	return c
}
