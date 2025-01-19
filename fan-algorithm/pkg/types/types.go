// types.go
package types

import (
    "github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
)

// TestResult represents the result of a test generation attempt
type TestResult struct {
    Success      bool                                    // Whether test generation was successful
    TestPattern  map[*circuit.Signal]circuit.SignalValue // Input assignments that detect the fault
    Implications []Assignment                            // List of implications made during test generation
    DFrontier    []DFrontierGate                        // Final D-frontier gates
    Decisions    []Decision                             // List of decisions made during test generation
}

// Assignment represents a value assignment to a signal
type Assignment struct {
    Signal *circuit.Signal      // The signal being assigned
    Value  circuit.SignalValue  // The value being assigned
    Reason AssignmentReason     // Reason for the assignment
}

// AssignmentReason represents why a value was assigned
type AssignmentReason int

const (
    DECISION AssignmentReason = iota    // Direct decision by algorithm
    IMPLICATION                         // Result of implication
    FAULT_INJECTION                     // Fault injection
    BACKTRACE                          // Result of backtrace
)

// DFrontierGate represents a gate in the D-frontier
type DFrontierGate struct {
    Gate       *circuit.Gate    // The gate in the D-frontier
    FaultyInput *circuit.Signal // Input signal carrying D or D'
}

// Decision represents a decision made during test generation
type Decision struct {
    Signal      *circuit.Signal      // Signal where decision was made
    Value       circuit.SignalValue  // Value assigned
    Alternative bool                 // Whether alternative value has been tried
    Level       int                  // Decision level in the tree
}

// BacktraceObjective represents an objective for backtrace
type BacktraceObjective struct {
    Signal    *circuit.Signal      // Target signal
    Value     circuit.SignalValue  // Desired value
    ZeroCount int                 // Number of times 0 is required
    OneCount  int                 // Number of times 1 is required
    Priority  int                 // Priority of this objective
}

// SensitizationPath represents a path that needs to be sensitized
type SensitizationPath struct {
    Gates     []*circuit.Gate    // Gates in the path
    Sensitized bool              // Whether path is currently sensitized
}

// CircuitState represents a snapshot of circuit state
type CircuitState struct {
    SignalValues map[*circuit.Signal]circuit.SignalValue
    DFrontier    []DFrontierGate
}

// TestGenerationConfig holds configuration for test generation
type TestGenerationConfig struct {
    MaxDecisions        int     // Maximum number of decisions allowed
    MaxBacktracks       int     // Maximum number of backtracks allowed
    UseUniqueSensitization bool // Whether to use unique sensitization
    UseDynamicBacktrace bool   // Whether to use dynamic backtrace ordering
}

// ObjectiveType represents types of objectives in test generation
type ObjectiveType int

const (
    PROPAGATE ObjectiveType = iota  // Propagate fault effect
    JUSTIFY                         // Justify a signal value
    SENSITIZE                       // Sensitize a path
)

// Objective represents a test generation objective
type Objective struct {
    Type      ObjectiveType
    Signal    *circuit.Signal
    Value     circuit.SignalValue
    Priority  int
}

// BacktraceResult represents the result of a backtrace operation
type BacktraceResult struct {
    Success          bool
    FinalObjectives  []*BacktraceObjective
    HeadLines        []*circuit.Signal
}

// ImplicationResult represents the result of an implication operation
type ImplicationResult struct {
    Success      bool
    Implications []Assignment
    Consistent   bool
}

// TestGenerationStats holds statistics about test generation
type TestGenerationStats struct {
    Decisions       int     // Number of decisions made
    Backtracks      int     // Number of backtracks performed
    Implications    int     // Number of implications performed
    BacktraceCount  int     // Number of backtrace operations
    ExecutionTime   float64 // Execution time in seconds
}

// Error types for test generation
type TestGenerationError string

const (
    ERROR_MAX_DECISIONS  TestGenerationError = "Maximum decisions exceeded"
    ERROR_MAX_BACKTRACKS TestGenerationError = "Maximum backtracks exceeded"
    ERROR_INCONSISTENCY  TestGenerationError = "Value inconsistency detected"
    ERROR_NO_SOLUTION    TestGenerationError = "No solution exists"
)

func (e TestGenerationError) Error() string {
    return string(e)
}

// Helper functions

// NewTestResult creates a new TestResult with default values
func NewTestResult() *TestResult {
    return &TestResult{
        Success:      false,
        TestPattern:  make(map[*circuit.Signal]circuit.SignalValue),
        Implications: make([]Assignment, 0),
        DFrontier:    make([]DFrontierGate, 0),
        Decisions:    make([]Decision, 0),
    }
}

// NewCircuitState creates a new CircuitState
func NewCircuitState() *CircuitState {
    return &CircuitState{
        SignalValues: make(map[*circuit.Signal]circuit.SignalValue),
        DFrontier:    make([]DFrontierGate, 0),
    }
}

// NewTestGenerationConfig creates a new configuration with default values
func NewTestGenerationConfig() *TestGenerationConfig {
    return &TestGenerationConfig{
        MaxDecisions:           1000,
        MaxBacktracks:          1000,
        UseUniqueSensitization: true,
        UseDynamicBacktrace:    true,
    }
}

// NewTestGenerationStats creates new statistics with zero values
func NewTestGenerationStats() *TestGenerationStats {
    return &TestGenerationStats{
        Decisions:      0,
        Backtracks:     0,
        Implications:   0,
        BacktraceCount: 0,
        ExecutionTime:  0.0,
    }
}