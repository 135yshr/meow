package mutation

import (
	"fmt"
	"io"
)

// Report writes the mutation testing results to the given writer.
func Report(w io.Writer, mutants []Mutant, results []RunResult) {
	killed := 0
	survived := 0

	for _, r := range results {
		if r.Killed {
			killed++
		} else {
			survived++
		}
	}

	total := len(results)
	score := float64(0)
	if total > 0 {
		score = float64(killed) / float64(total) * 100
	}

	fmt.Fprintf(w, "\n=== Mutation Test Results ===\n")
	fmt.Fprintf(w, "Total mutants: %d\n", total)
	fmt.Fprintf(w, "Killed: %d\n", killed)
	fmt.Fprintf(w, "Survived: %d\n", survived)
	fmt.Fprintf(w, "Mutation score: %.1f%%\n", score)

	if survived > 0 {
		fmt.Fprintf(w, "\n--- Surviving Mutants ---\n")
		for _, r := range results {
			if !r.Killed {
				for _, m := range mutants {
					if m.ID == r.ID {
						fmt.Fprintf(w, "  [%d] %s (%s)\n", m.ID, m.Description, m.Pos)
						break
					}
				}
			}
		}
	}
	fmt.Fprintln(w)
}
