package test

import (
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/examples"
	"testing"
)

func TestCircuitStructure(t *testing.T) {
	c := examples.CreateFanTestCircuit()

	if len(c.Gates) == 0 {
		t.Errorf("Circuit should have gates")
	}

	// Test fanout identification
	found := false
	for _, s := range c.Signals {
		if len(s.Fanouts) > 1 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Circuit should have at least one fanout point")
	}
}
