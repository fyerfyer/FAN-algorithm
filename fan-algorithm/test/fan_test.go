package test

import (
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/examples"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/algorithm"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
	"testing"
)

func TestSimpleFaultDetection(t *testing.T) {
	c := examples.CreateSimpleCircuit()

	// Verify circuit structure
	if len(c.PrimaryInputs) == 0 {
		t.Fatal("Circuit has no primary inputs")
	}

	faultSite := c.PrimaryInputs[0]
	faultValue := circuit.ZERO

	// Run FAN algorithm
	result := algorithm.FAN(c, faultSite, faultValue)

	// Debug logging
	t.Logf("Circuit state after test:")
	t.Logf("- Primary inputs: %d", len(c.PrimaryInputs))
	t.Logf("- Primary outputs: %d", len(c.PrimaryOutputs))
	t.Logf("- Total signals: %d", len(c.Signals))
	t.Logf("- D-frontier size: %d", len(result.DFrontier))

	if !result.Success {
		t.Errorf("Failed to find test pattern for stuck-at-0 fault at %s", faultSite.ID)
		t.Logf("Final circuit state:")
		for _, s := range c.Signals {
			t.Logf("Signal %s = %s", s.ID, valueToString(s.GetValue()))
		}
		t.Logf("D-Frontier gates: %d", len(result.DFrontier))
	}
}

func TestMultipleDFrontierPaths(t *testing.T) {
	c := examples.CreateFanTestCircuit()

	signal, err := c.GetSignalByID("n1")
	if err != nil { // Fix error check
		t.Fatalf("Could not find signal n1: %v", err)
	}

	faultValue := circuit.ONE

	// Run FAN with debug
	result := algorithm.FAN(c, signal, faultValue)

	// Enhanced debug output
	t.Logf("Test results:")
	t.Logf("- Success: %v", result.Success)
	t.Logf("- D-frontier size: %d", len(result.DFrontier))
	t.Logf("- Decisions made: %d", result.Stats.Decisions)
	t.Logf("- Backtrack count: %d", result.Stats.Backtracks)

	if !result.Success {
		t.Error("Failed to handle multiple D-frontier paths")
	}
	if len(result.DFrontier) == 0 {
		t.Error("D-frontier should not be empty")
		printCircuitState(t, c)
	}
}

func TestUniqueSensitization(t *testing.T) {
	c := examples.CreateFanTestCircuit()
	// Get signal that requires unique sensitization
	signal, _ := c.GetSignalByID("n3")
	faultValue := circuit.ZERO

	result := algorithm.FAN(c, signal, faultValue)

	if !result.Success {
		t.Errorf("Failed to handle unique sensitization case")
	}
}

// Helper functions
func valueToString(v circuit.SignalValue) string {
	switch v {
	case circuit.ZERO:
		return "0"
	case circuit.ONE:
		return "1"
	case circuit.D:
		return "D"
	case circuit.D_BAR:
		return "D'"
	case circuit.X:
		return "X"
	default:
		return "?"
	}
}

func printCircuitState(t *testing.T, c *circuit.Circuit) {
	t.Log("Circuit state:")
	for _, g := range c.Gates {
		inputs := make([]string, len(g.Inputs))
		for i, in := range g.Inputs {
			inputs[i] = valueToString(in.GetValue())
		}
		t.Logf("Gate %s: inputs=%v, output=%s",
			g.ID,
			inputs,
			valueToString(g.Output.GetValue()))
	}
}
