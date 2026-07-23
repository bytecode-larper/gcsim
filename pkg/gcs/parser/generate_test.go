package parser

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestGeneratedParserUpToDate fails if gcsim_parser.go is not the fresh
// output of `go generate` / pigeon on gcsim.peg. Run:
//
//	go generate ./pkg/gcs/parser
func TestGeneratedParserUpToDate(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	peg := filepath.Join(dir, "gcsim.peg")
	committed := filepath.Join(dir, "gcsim_parser.go")
	if _, err := os.Stat(peg); err != nil {
		t.Skip("gcsim.peg not present")
	}
	want, err := os.ReadFile(committed)
	if err != nil {
		t.Fatalf("read committed parser: %v", err)
	}

	out := filepath.Join(t.TempDir(), "gcsim_parser.go")
	cmd := exec.Command("go", "tool", "pigeon", "-optimize-parser", "-o", out, peg)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("pigeon generate failed: %v\n%s", err, stderr.String())
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(want, got) {
		t.Fatalf("gcsim_parser.go is out of date with gcsim.peg; run: go generate ./pkg/gcs/parser")
	}
}
