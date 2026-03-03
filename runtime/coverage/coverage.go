package coverage

import (
	"fmt"
	"io"
	"os"
)

// Block represents a single instrumented statement.
type Block struct {
	FileName    string
	StartLine   int
	StartCol    int
	EndLine     int
	EndCol      int
	NumStmt     int
	Count       int
}

var blocks []Block

// Register adds a new coverage block and returns its ID.
func Register(fileName string, startLine, startCol, endLine, endCol, numStmt int) int {
	id := len(blocks)
	blocks = append(blocks, Block{
		FileName:  fileName,
		StartLine: startLine,
		StartCol:  startCol,
		EndLine:   endLine,
		EndCol:    endCol,
		NumStmt:   numStmt,
	})
	return id
}

// Hit records an execution of the block with the given ID.
func Hit(id int) {
	blocks[id].Count++
}

// Report writes a coverage summary to w.
func Report(w io.Writer) {
	if len(blocks) == 0 {
		return
	}
	total := 0
	covered := 0
	for _, b := range blocks {
		total += b.NumStmt
		if b.Count > 0 {
			covered += b.NumStmt
		}
	}
	pct := 0.0
	if total > 0 {
		pct = float64(covered) / float64(total) * 100
	}
	fmt.Fprintf(w, "coverage: %.1f%% of statements, nya~\n", pct)
}

// WriteProfile writes block data in Go-compatible coverage profile format.
// It appends to the file (the caller writes the "mode: set" header).
func WriteProfile(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, b := range blocks {
		count := 0
		if b.Count > 0 {
			count = 1
		}
		fmt.Fprintf(f, "%s:%d.%d,%d.%d %d %d\n",
			b.FileName, b.StartLine, b.StartCol, b.EndLine, b.EndCol,
			b.NumStmt, count)
	}
	return nil
}

// Reset clears all registered blocks. Used for testing.
func Reset() {
	blocks = nil
}

// Blocks returns the current block list. Used for testing.
func Blocks() []Block {
	return blocks
}
