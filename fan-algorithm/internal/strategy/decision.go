package strategy

import (
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/utils"
)

// DecisionNode represents a node in the decision tree
type DecisionNode struct {
	Signal   *circuit.Signal
	Value    circuit.SignalValue
	Tried    bool // Whether alternative value has been tried
	Parent   *DecisionNode
	Children []*DecisionNode
}

// DecisionTree manages backtracking decisions
type DecisionTree struct {
	Root    *DecisionNode
	Current *DecisionNode
	Circuit *circuit.Circuit
}

func NewDecisionTree(c *circuit.Circuit) *DecisionTree {
	return &DecisionTree{
		Circuit: c,
	}
}

// AddDecision adds a new decision to the tree
func (dt *DecisionTree) AddDecision(signal *circuit.Signal, value circuit.SignalValue) *DecisionNode {
	node := &DecisionNode{
		Signal: signal,
		Value:  value,
		Tried:  false,
	}

	if dt.Root == nil {
		dt.Root = node
		dt.Current = node
	} else {
		node.Parent = dt.Current
		dt.Current.Children = append(dt.Current.Children, node)
		dt.Current = node
	}

	return node
}

// Backtrack returns to previous decision point and tries alternative
func (dt *DecisionTree) Backtrack() bool {
	if dt.Current == nil {
		return false
	}

	// Try alternative value if not tried yet
	if !dt.Current.Tried {
		dt.Current.Tried = true
		dt.Current.Value = utils.GetAlternativeValue(dt.Current.Value)
		return true
	}

	// Move to parent node
	if dt.Current.Parent != nil {
		dt.Current = dt.Current.Parent
		return dt.Backtrack()
	}

	return false
}

// Reset circuit to state before current decision
func (dt *DecisionTree) ResetToCurrentState() {
	// Reset all signals to X
	for _, signal := range dt.Circuit.Signals {
		if !signal.IsPrimary {
			signal.SetValue(circuit.X)
		}
	}

	// Replay decisions up to current node
	node := dt.Root
	for node != nil {
		node.Signal.SetValue(node.Value)
		if len(node.Children) > 0 {
			node = node.Children[0]
		} else {
			break
		}
	}
}
