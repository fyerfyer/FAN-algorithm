// main.go
package main

import (
    "fmt"
    "time"

    "github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/algorithm"
    "github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
    "github.com/fyerfyer/FAN-algorithm/fan-algorithm/examples"
)

func main() {
    // Create test circuits
    c17 := examples.CreateC17Circuit()
    simpleCircuit := examples.CreateSimpleCircuit()

    // Test C17 circuit
    fmt.Println("Testing C17 circuit...")
    testCircuit(c17)

    fmt.Println("\nTesting simple circuit...")
    testCircuit(simpleCircuit)
}

func testCircuit(c *circuit.Circuit) {
    // Print circuit information
    fmt.Printf("Circuit contains:\n")
    fmt.Printf("- %d gates\n", len(c.Gates))
    fmt.Printf("- %d signals\n", len(c.Signals))
    fmt.Printf("- %d primary inputs\n", len(c.PrimaryInputs))
    fmt.Printf("- %d primary outputs\n", len(c.PrimaryOutputs))
    fmt.Printf("- %d head lines\n", len(c.HeadLines))

    // Test each primary input for stuck-at faults
    for _, input := range c.PrimaryInputs {
        // Test stuck-at-0
        fmt.Printf("\nTesting stuck-at-0 fault at %s\n", input.ID)
        testFault(c, input, circuit.ZERO)

        // Test stuck-at-1
        fmt.Printf("\nTesting stuck-at-1 fault at %s\n", input.ID)
        testFault(c, input, circuit.ONE)
    }
}

func testFault(c *circuit.Circuit, faultSite *circuit.Signal, faultValue circuit.SignalValue) {
    start := time.Now()

    // Run FAN algorithm
    result := algorithm.FAN(c, faultSite, faultValue)

    duration := time.Since(start)

    // Print results
    if result.Success {
        fmt.Printf("Test pattern found in %v\n", duration)
        fmt.Println("Test pattern:")
        for input, value := range result.TestPattern {
            fmt.Printf("%s = %v\n", input.ID, value)
        }
    } else {
        fmt.Printf("No test pattern found after %v\n", duration)
    }

    // Print D-frontier information
    fmt.Printf("Final D-frontier size: %d\n", len(result.DFrontier))
}