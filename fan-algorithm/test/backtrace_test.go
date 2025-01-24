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

	if result == nil {
		t.Fatal("MultipleBacktrace returned nil result")
	}

	// Success if we have either:
	// 1. Initial objective preserved
	// 2. New objectives generated
	// 3. Head lines found
	if len(result.FinalObjectives) == 0 && len(result.HeadLines) == 0 {
		t.Error("Multiple backtrace should produce either objectives or head lines")
		t.Logf("Signal: %s, Value: %v", signal.ID, initialObjective.Value)
	}

	// Verify objective properties
	for _, obj := range result.FinalObjectives {
		if obj.Signal == nil {
			t.Error("Objective signal should not be nil")
		}
	}
}

// Add helper test functions
func TestBacktraceHeadLines(t *testing.T) {
	c := examples.CreateFanTestCircuit()

	// Verify circuit structure first
	if len(c.HeadLines) == 0 {
		t.Fatal("Circuit should have head lines after initialization")
	}

	// Print head lines for debugging
	t.Logf("Found %d head lines:", len(c.HeadLines))
	for _, hl := range c.HeadLines {
		t.Logf("- Head line: %s", hl.ID)
	}

	// Get first head line
	headLine := c.HeadLines[0]
	if headLine == nil {
		t.Fatal("Could not find a head line in circuit")
	}

	initialObjective := &types.BacktraceObjective{
		Signal:    headLine,
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

	// Verify results
	if len(result.HeadLines) == 0 {
		t.Errorf("Expected head lines in backtrace result, got none. Initial signal: %s", headLine.ID)
	}

	// Additional validation
	foundInitial := false
	for _, hl := range result.HeadLines {
		if hl == headLine {
			foundInitial = true
			break
		}
	}
	if !foundInitial {
		t.Error("Initial head line should be in result head lines")
	}
}

func TestBacktraceWithUniqueSensitization(t *testing.T) {
	c := examples.CreateFanTestCircuit()

	// Use n2 which should require unique sensitization
	signal, _ := c.GetSignalByID("n2")
	if signal == nil {
		t.Fatal("Could not find test signal n2")
	}

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

	if result == nil {
		t.Fatal("Backtrace returned nil result")
	}

	// Check stats for decisions
	if result.Stats.Decisions == 0 {
		t.Error("Should make decisions during unique sensitization")
	}

	// Verify mandatory paths were found
	if len(result.FinalObjectives) == 0 {
		t.Error("Should find objectives for unique sensitization paths")
	}

	t.Logf("Decisions made: %d", result.Stats.Decisions)
	t.Logf("Final objectives found: %d", len(result.FinalObjectives))
}
