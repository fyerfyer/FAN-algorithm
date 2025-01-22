package types

import (
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
	"time"
)

type AssignmentReason int

const (
	DECISION AssignmentReason = iota
	IMPLICATION
	UNIQUE_SENSITIZATION
	MANDATORY_PATH
)

type SensitizationPath struct {
	Gates   []*circuit.Gate
	Signals []*circuit.Signal
	Score   int
}

type ObjectiveType int

const (
	PROPAGATE ObjectiveType = iota
	JUSTIFY
)

func NewCircuitState() *CircuitState {
	return &CircuitState{
		SignalValues:        make(map[*circuit.Signal]circuit.SignalValue),
		DFrontier:           make([]DFrontierGate, 0),
		JustificationStatus: make(map[*circuit.Signal]bool),
		SensitizedPaths:     make([]*SensitizationPath, 0),
	}
}

func NewTestGenerationStats() *TestGenerationStats {
	return &TestGenerationStats{
		ExecutionTime: 0,
		SuccessRate:   0.0,
	}
}

// TestResult represents the result of a test generation attempt
type TestResult struct {
	Success      bool
	TestPattern  map[*circuit.Signal]circuit.SignalValue
	Implications []Assignment
	DFrontier    []DFrontierGate
	Decisions    []Decision
	Stats        *TestGenerationStats
	Error        TestGenerationError
	CircuitState *CircuitState
}

// Assignment with enhanced reason tracking
type Assignment struct {
	Signal    *circuit.Signal
	Value     circuit.SignalValue
	Reason    AssignmentReason
	Level     int       // Decision level where this assignment was made
	TimeStamp time.Time // When the assignment was made
}

// DFrontierGate with additional analysis info
type DFrontierGate struct {
	Gate           *circuit.Gate
	FaultyInput    *circuit.Signal
	Priority       int               // Priority for processing this gate
	BlockingInputs []*circuit.Signal // Inputs blocking fault propagation
}

// Decision with enhanced backtracking support
type Decision struct {
	Signal      *circuit.Signal
	Value       circuit.SignalValue
	Alternative bool
	Level       int
	Children    []Assignment // Implications resulting from this decision
	Score       int          // Decision quality score
	TimeStamp   time.Time
}

// BacktraceObjective with improved priority handling
type BacktraceObjective struct {
	Signal       *circuit.Signal
	Value        circuit.SignalValue
	ZeroCount    int
	OneCount     int
	Priority     int
	Cost         int                   // Cost of achieving this objective
	Dependencies []*BacktraceObjective // Other objectives this depends on
}

// CircuitState with enhanced state tracking
type CircuitState struct {
	SignalValues        map[*circuit.Signal]circuit.SignalValue
	DFrontier           []DFrontierGate
	DecisionLevel       int
	JustificationStatus map[*circuit.Signal]bool // Tracks justified signals
	SensitizedPaths     []*SensitizationPath
}

// Add new method to check state consistency
func (cs *CircuitState) IsConsistent() bool {
	for signal, value := range cs.SignalValues {
		if !signal.IsCompatible(value) {
			return false
		}
	}
	return true
}

// Add method to clone circuit state
func (cs *CircuitState) Clone() *CircuitState {
	newState := NewCircuitState()
	for signal, value := range cs.SignalValues {
		newState.SignalValues[signal] = value
	}
	newState.DecisionLevel = cs.DecisionLevel
	return newState
}

// TestGenerationConfig with additional options
type TestGenerationConfig struct {
	MaxDecisions           int
	MaxBacktracks          int
	UseUniqueSensitization bool
	UseDynamicBacktrace    bool
	TimeLimit              time.Duration
	PreferredHeadLines     []*circuit.Signal
	BacktraceStrategy      BacktraceStrategy
	PropagationStrategy    PropagationStrategy
}

// Add strategy enums
type BacktraceStrategy int
type PropagationStrategy int

const (
	STATIC_BACKTRACE BacktraceStrategy = iota
	DYNAMIC_BACKTRACE
	HYBRID_BACKTRACE
)

const (
	FORWARD_PROPAGATION PropagationStrategy = iota
	BACKWARD_PROPAGATION
	BIDIRECTIONAL_PROPAGATION
)

// Enhanced stats tracking
type TestGenerationStats struct {
	Decisions                    int
	Backtracks                   int
	Implications                 int
	BacktraceCount               int
	ExecutionTime                time.Duration
	MaxDecisionLevel             int
	SuccessRate                  float64
	AverageBacktracksPerDecision float64
}

// Add method to update stats
func (s *TestGenerationStats) Update(decision bool, backtrack bool) {
	if decision {
		s.Decisions++
	}
	if backtrack {
		s.Backtracks++
	}
	s.AverageBacktracksPerDecision = float64(s.Backtracks) / float64(s.Decisions)
}

// Enhanced error types
type TestGenerationError interface {
	Error() string
	Code() int
}

type testError struct {
	message string
	code    int
}

func (e *testError) Error() string { return e.message }
func (e *testError) Code() int     { return e.code }

// Predefined errors
var (
	ErrMaxDecisions  = &testError{"Maximum decisions exceeded", 1}
	ErrMaxBacktracks = &testError{"Maximum backtracks exceeded", 2}
	ErrInconsistency = &testError{"Value inconsistency detected", 3}
	ErrNoSolution    = &testError{"No solution exists", 4}
	ErrTimeout       = &testError{"Time limit exceeded", 5}
)

// Enhanced constructor functions
func NewTestResult() *TestResult {
	return &TestResult{
		Success:      false,
		TestPattern:  make(map[*circuit.Signal]circuit.SignalValue),
		Implications: make([]Assignment, 0),
		DFrontier:    make([]DFrontierGate, 0),
		Decisions:    make([]Decision, 0),
		Stats:        NewTestGenerationStats(),
		CircuitState: NewCircuitState(),
	}
}

func NewTestGenerationConfig() *TestGenerationConfig {
	return &TestGenerationConfig{
		MaxDecisions:           1000,
		MaxBacktracks:          1000,
		UseUniqueSensitization: true,
		UseDynamicBacktrace:    true,
		TimeLimit:              time.Minute * 5,
		BacktraceStrategy:      DYNAMIC_BACKTRACE,
		PropagationStrategy:    BIDIRECTIONAL_PROPAGATION,
	}
}

// Add helper functions for objective management
func CreateObjective(signal *circuit.Signal, objType ObjectiveType, value circuit.SignalValue) *BacktraceObjective {
	return &BacktraceObjective{
		Signal:   signal,
		Value:    value,
		Priority: calculateObjectivePriority(signal, value),
		Cost:     estimateObjectiveCost(signal, value),
	}
}

func calculateObjectivePriority(signal *circuit.Signal, value circuit.SignalValue) int {
	// Priority calculation based on circuit structure and value
	priority := 0
	if signal.IsPrimary {
		priority += 10
	}
	if signal.IsHead {
		priority += 5
	}
	if signal.IsFanoutPoint() {
		priority -= 3
	}
	return priority
}

func estimateObjectiveCost(signal *circuit.Signal, value circuit.SignalValue) int {
	// Estimate cost based on controllability
	cost := signal.Controllability
	if signal.FanIn != nil {
		cost += len(signal.FanIn.Inputs) * 2
	}
	return cost
}
