package mutation

import "github.com/135yshr/meow/pkg/token"

// MutantID uniquely identifies a mutation.
type MutantID int

// MutantKind categorizes the type of mutation.
type MutantKind int

const (
	ArithmeticSwap  MutantKind = iota // +↔-, *↔/
	ComparisonSwap                    // ==↔!=, <↔<=, >↔>=
	LogicalSwap                       // &&↔||
	NegationRemoval                   // -x→x, !x→x
	BoolFlip                          // yarn↔hairball
	IntBoundary                       // 0→1, nonzero→0
	StringEmpty                       // ""→"mutant", non-empty→""
	ConditionNegate                   // if(c)→if(!c)
	ReturnNil                         // bring x→bring catnap
	CatchRemove                       // expr ~> fallback → expr
	PipeRemove                        // xs |=| f → xs
)

// Mutant represents a single mutation that can be applied to and reverted from the AST.
type Mutant struct {
	ID          MutantID
	Description string
	Pos         token.Position
	Kind        MutantKind
	Apply       func()
	Undo        func()
}

// RunResult holds the outcome of running tests against a single mutant.
type RunResult struct {
	ID     MutantID
	Killed bool
}
