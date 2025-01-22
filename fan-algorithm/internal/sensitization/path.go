package sensitization

import (
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/circuit"
	"github.com/fyerfyer/FAN-algorithm/fan-algorithm/internal/utils"
)

// Path represents a sensitization path in the circuit
type Path struct {
	Gates   []*circuit.Gate
	Signals []*circuit.Signal
	Score   int // Path priority score
}

// PathFinder handles path analysis for sensitization
type PathFinder struct {
	Circuit *circuit.Circuit
}

func NewPathFinder(c *circuit.Circuit) *PathFinder {
	return &PathFinder{Circuit: c}
}

// FindUniqueSensitizationPaths finds paths that must be uniquely sensitized
func (pf *PathFinder) FindUniqueSensitizationPaths(dFrontier []*circuit.Gate) []*Path {
	if len(dFrontier) != 1 {
		return nil
	}

	paths := make([]*Path, 0)
	startGate := dFrontier[0]

	// Find all paths to primary outputs
	var dfs func(*circuit.Gate, []*circuit.Gate)
	dfs = func(gate *circuit.Gate, currentPath []*circuit.Gate) {
		if gate == nil {
			return
		}

		newPath := append(currentPath, gate)

		// If reached primary output
		if gate.Output.IsPrimary {
			path := &Path{
				Gates: append([]*circuit.Gate{}, newPath...),
				Score: utils.CalculatePathScore(newPath),
			}
			paths = append(paths, path)
			return
		}

		// Continue through fanouts
		for _, fanout := range gate.Output.Fanouts {
			if fanout.FanIn != nil {
				dfs(fanout.FanIn, newPath)
			}
		}
	}

	dfs(startGate, make([]*circuit.Gate, 0))
	return paths
}

// GetMandatorySignals finds signals that must be sensitized
func (pf *PathFinder) GetMandatorySignals(paths []*Path) []*circuit.Signal {
	if len(paths) == 0 {
		return nil
	}

	// Count signal occurrences across all paths
	signalCount := make(map[*circuit.Signal]int)
	for _, path := range paths {
		for _, gate := range path.Gates {
			signalCount[gate.Output]++
		}
	}

	// Find signals that appear in all paths
	mandatory := make([]*circuit.Signal, 0)
	for signal, count := range signalCount {
		if count == len(paths) {
			mandatory = append(mandatory, signal)
		}
	}

	return mandatory
}
