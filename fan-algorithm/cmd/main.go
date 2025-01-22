package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/examples"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/algorithm"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/pkg/types"
)

// TestResult stores the results of testing a fault
type TestResult struct {
	CircuitName string
	FaultSite   string
	FaultValue  circuit.SignalValue
	Success     bool
	TestPattern map[*circuit.Signal]circuit.SignalValue
	DFrontier   []types.DFrontierGate
	Stats       *types.TestGenerationStats
	Duration    time.Duration
}

func main() {
	// Create all test circuits
	circuits := map[string]*circuit.Circuit{
		"C17 Benchmark": examples.CreateC17Circuit(),
		"Simple":        examples.CreateSimpleCircuit(),
		"FAN Test":      examples.CreateFanTestCircuit(),
	}

	// Test each circuit
	for name, c := range circuits {
		fmt.Printf("\n%s\n%s\n", name, strings.Repeat("=", len(name)))
		printCircuitInfo(c)
		testAllFaults(name, c)
	}
}

func printCircuitInfo(c *circuit.Circuit) {
	fmt.Printf("\nCircuit Structure:\n")
	fmt.Printf("- Gates: %d\n", len(c.Gates))
	fmt.Printf("- Signals: %d\n", len(c.Signals))
	fmt.Printf("- Primary Inputs: %d\n", len(c.PrimaryInputs))
	fmt.Printf("- Primary Outputs: %d\n", len(c.PrimaryOutputs))
	fmt.Printf("- Head Lines: %d\n", len(c.HeadLines))
}

func testAllFaults(circuitName string, c *circuit.Circuit) {
	results := make([]*TestResult, 0)

	// Test primary inputs
	for _, signal := range c.PrimaryInputs {
		// Test stuck-at-0
		results = append(results, testFault(circuitName, c, signal, circuit.ZERO))
		// Test stuck-at-1
		results = append(results, testFault(circuitName, c, signal, circuit.ONE))
	}

	// Test internal signals
	for _, signal := range c.Signals {
		if !signal.IsPrimary {
			results = append(results, testFault(circuitName, c, signal, circuit.ZERO))
			results = append(results, testFault(circuitName, c, signal, circuit.ONE))
		}
	}

	// Print summary
	printTestSummary(results)
}

func testFault(circuitName string, c *circuit.Circuit, faultSite *circuit.Signal, faultValue circuit.SignalValue) *TestResult {
	start := time.Now()
	algResult := algorithm.FAN(c, faultSite, faultValue)
	duration := time.Since(start)

	return &TestResult{
		CircuitName: circuitName,
		FaultSite:   faultSite.ID,
		FaultValue:  faultValue,
		Success:     algResult.Success,
		TestPattern: algResult.TestPattern,
		DFrontier:   algResult.DFrontier,
		Stats:       algResult.Stats,
		Duration:    duration,
	}
}

func printTestSummary(results []*TestResult) {
	fmt.Printf("\nTest Results Summary:\n")
	fmt.Printf("%-20s %-15s %-10s %-10s %-15s %-10s\n",
		"Fault Site", "Stuck-At", "Result", "Backtracks", "D-Frontier", "Time")
	fmt.Println(strings.Repeat("-", 80))

	totalTests := len(results)
	successfulTests := 0
	totalBacktracks := 0
	totalDuration := time.Duration(0)

	for _, r := range results {
		status := "FAIL"
		if r.Success {
			status = "PASS"
			successfulTests++
		}

		fmt.Printf("%-20s %-15d %-10s %-10d %-15d %v\n",
			r.FaultSite,
			r.FaultValue,
			status,
			r.Stats.Backtracks,
			len(r.DFrontier),
			r.Duration)

		totalBacktracks += r.Stats.Backtracks
		totalDuration += r.Duration
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("\nSummary Statistics:\n")
	fmt.Printf("- Total Faults Tested: %d\n", totalTests)
	fmt.Printf("- Testable Faults: %d (%.1f%%)\n",
		successfulTests, float64(successfulTests)*100/float64(totalTests))
	fmt.Printf("- Average Backtracks: %.2f\n",
		float64(totalBacktracks)/float64(totalTests))
	fmt.Printf("- Average Time per Test: %v\n",
		totalDuration/time.Duration(totalTests))
	fmt.Printf("- Total Test Time: %v\n\n", totalDuration)
}
