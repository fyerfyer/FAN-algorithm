package test

import (
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/examples"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/algorithm"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
	"testing"
)

func TestSimpleFaultDetection(t *testing.T) {
	c := examples.CreateSimpleCircuit()
	faultSite := c.PrimaryInputs[0]
	faultValue := circuit.ZERO

	result := algorithm.FAN(c, faultSite, faultValue)

	if !result.Success {
		t.Errorf("Failed to find test pattern for simple stuck-at-0 fault")
	}
}

func TestMultipleDFrontierPaths(t *testing.T) {
	c := examples.CreateFanTestCircuit()
	// Get internal signal with multiple paths to outputs
	signal, _ := c.GetSignalByID("n1")
	faultValue := circuit.ONE

	result := algorithm.FAN(c, signal, faultValue)

	if !result.Success {
		t.Errorf("Failed to handle multiple D-frontier paths")
	}
	if len(result.DFrontier) == 0 {
		t.Errorf("D-frontier should not be empty")
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
