package test

import (
	"testing"

	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/examples"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/algorithm"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/pkg/types"
)

func TestMultipleBacktrace(t *testing.T) {
	c := examples.CreateFanTestCircuit()
	signal, _ := c.GetSignalByID("n1")

	// Create initial objective
	initialObjective := &types.BacktraceObjective{
		Signal:    signal,
		Value:     circuit.ONE,
		ZeroCount: 0,
		OneCount:  1,
		Priority:  10,
	}

	// Create config
	config := types.NewTestGenerationConfig()

	// Call MultipleBacktrace with proper arguments
	result := algorithm.MultipleBacktrace(
		[]*types.BacktraceObjective{initialObjective},
		c,
		config,
	)

	// Verify results
	if result == nil {
		t.Fatal("MultipleBacktrace returned nil result")
	}

	if len(result.FinalObjectives) == 0 {
		t.Error("Multiple backtrace should produce objectives")
	}

	// Test objective properties
	for _, obj := range result.FinalObjectives {
		if obj.Signal == nil {
			t.Error("Objective signal should not be nil")
		}
		if obj.Priority == 0 {
			t.Error("Objective should have non-zero priority")
		}
	}
}

// Add helper test functions
func TestBacktraceHeadLines(t *testing.T) {
	c := examples.CreateFanTestCircuit()
	signal, _ := c.GetSignalByID("n3") // Use a signal that requires head line processing

	initialObjective := &types.BacktraceObjective{
		Signal:    signal,
		Value:     circuit.ONE,
		ZeroCount: 0,
		OneCount:  1,
		Priority:  10,
	}

	config := types.NewTestGenerationConfig()
	result := algorithm.MultipleBacktrace(
		[]*types.BacktraceObjective{initialObjective},
		c,
		config,
	)

	if len(result.HeadLines) == 0 {
		t.Error("Should identify head lines during backtrace")
	}
}

func TestBacktraceWithUniqueSensitization(t *testing.T) {
	c := examples.CreateFanTestCircuit()
	signal, _ := c.GetSignalByID("n1")

	initialObjective := &types.BacktraceObjective{
		Signal:    signal,
		Value:     circuit.ONE,
		ZeroCount: 0,
		OneCount:  1,
		Priority:  10,
	}

	config := types.NewTestGenerationConfig()
	config.UseUniqueSensitization = true

	result := algorithm.MultipleBacktrace(
		[]*types.BacktraceObjective{initialObjective},
		c,
		config,
	)

	if result.Stats.Decisions <= 0 {
		t.Error("Should make decisions during unique sensitization")
	}
}
