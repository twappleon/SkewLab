package main

import (
	"encoding/json"
	"github.com/twappleon/SkewLab/internal/paper"
	"os"
	"path/filepath"
	"testing"
)

func TestNoArchiveCannotInventResult(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "ledger.json")
	b, _ := json.Marshal(paper.New(paper.DefaultConfig()))
	if e := os.WriteFile(p, b, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := run(p, filepath.Join(dir, "missing-archive*")); e == nil {
		t.Fatal("missing archive accepted")
	}
	if _, e := run("", ""); e == nil {
		t.Fatal("missing ledger accepted")
	}
}
