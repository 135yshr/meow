package mutation

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Runner executes mutation tests by running a pre-built binary with different
// MEOW_MUTANT environment variable values.
type Runner struct {
	BinaryPath  string
	TestTimeout time.Duration
}

// NewRunner creates a new mutation test runner.
func NewRunner(binaryPath string, timeout time.Duration) *Runner {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Runner{
		BinaryPath:  binaryPath,
		TestTimeout: timeout,
	}
}

// RunAll executes the test binary once per mutant.
// A mutant is "killed" if the test binary exits with non-zero status.
func (r *Runner) RunAll(mutants []Mutant) []RunResult {
	results := make([]RunResult, len(mutants))
	for i, m := range mutants {
		results[i] = RunResult{
			ID:     m.ID,
			Killed: r.runOne(m.ID),
		}
	}
	return results
}

// runOne runs the test binary with MEOW_MUTANT set to the given mutant ID.
// Returns true if the mutant was killed (test failed).
func (r *Runner) runOne(id MutantID) bool {
	cmd := exec.Command(r.BinaryPath)
	cmd.Env = append(os.Environ(), fmt.Sprintf("MEOW_MUTANT=%d", id))
	cmd.Stdout = nil
	cmd.Stderr = nil

	done := make(chan error, 1)
	if err := cmd.Start(); err != nil {
		return true // Build failure counts as killed
	}
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		return err != nil // Non-zero exit = killed
	case <-time.After(r.TestTimeout):
		cmd.Process.Kill()
		<-done
		return true // Timeout = killed
	}
}
