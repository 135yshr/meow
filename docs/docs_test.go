// Package docs holds the written documentation. The Go file here is a test
// only: it checks that the copies the website serves still say what the
// documents in this directory say.
package docs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// paired names the documents that exist twice: once here, and once under
// website/content/doc for Hugo to serve.
var paired = []string{"spec", "reference", "stdlib", "internals"}

// TestTheWebsiteSaysWhatTheDocumentsSay checks each pair line by line.
//
// The two copies are kept by hand, and had drifted seventeen commits apart
// before this test existed — the website went on describing a language without
// scram, without a basket, and without any way to reach a Go package. Nothing
// says so when it happens: the site builds, the tests pass, and only a reader
// finds out.
//
// A copy differs from its document in exactly two ways, both deliberate. Hugo
// needs front matter where the document has its title, and the opening line is
// written for a search engine rather than for a reader who already arrived.
// Everything after that has to match.
func TestTheWebsiteSaysWhatTheDocumentsSay(t *testing.T) {
	for _, name := range paired {
		t.Run(name, func(t *testing.T) {
			doc := readLines(t, name+".md")
			site := readLines(t, filepath.Join("..", "website", "content", "doc", name+".md"))

			body, ok := afterFrontMatter(site)
			if !ok {
				t.Fatalf("website/content/doc/%s.md has no front matter", name)
			}
			// The document opens with its title; the copy opens with its own
			// first line, which is the one difference allowed after this.
			want := trimBlank(dropTitle(doc))
			got := trimBlank(body)
			if len(want) == 0 || len(got) == 0 {
				t.Fatalf("one of the pair is empty: docs=%d website=%d", len(want), len(got))
			}
			want, got = want[1:], got[1:]

			for i := range max(len(want), len(got)) {
				switch {
				case i >= len(want):
					t.Fatalf("website/content/doc/%s.md has a line docs/%s.md does not, at %d: %q\n"+
						"copy the document over the website's, keeping its front matter and opening line",
						name, name, i+2, got[i])
				case i >= len(got):
					t.Fatalf("docs/%s.md has a line the website copy does not, at %d: %q\n"+
						"copy the document over the website's, keeping its front matter and opening line",
						name, i+2, want[i])
				case want[i] != got[i]:
					t.Fatalf("docs/%s.md and its website copy differ at line %d:\n  docs:    %q\n  website: %q\n"+
						"copy the document over the website's, keeping its front matter and opening line",
						name, i+2, want[i], got[i])
				}
			}
		})
	}
}

func readLines(t *testing.T, path string) []string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}
	// A checkout on Windows may carry CRLF. Left alone, every line would keep
	// a trailing \r, the front matter's "---" would not be recognized, and the
	// test would report drift that is not there.
	text := strings.ReplaceAll(string(b), "\r\n", "\n")
	return strings.Split(strings.TrimRight(text, "\n"), "\n")
}

// afterFrontMatter drops Hugo's `---` block from the front of a copy.
func afterFrontMatter(lines []string) ([]string, bool) {
	if len(lines) == 0 || lines[0] != "---" {
		return nil, false
	}
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			return lines[i+1:], true
		}
	}
	return nil, false
}

// dropTitle drops a document's `# ` heading, which the copy has as front matter.
func dropTitle(lines []string) []string {
	if len(lines) > 0 && strings.HasPrefix(lines[0], "# ") {
		return lines[1:]
	}
	return lines
}

func trimBlank(lines []string) []string {
	for len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	return lines
}
