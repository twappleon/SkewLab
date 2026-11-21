package main

import (
	"encoding/json"
	"github.com/twappleon/SkewLab/internal/paper"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadOnlyCLIGateBlocksSmallSample(t *testing.T) {
	dir := t.TempDir()
	l := paper.New(paper.DefaultConfig())
	b, _ := json.Marshal(l)
	path := filepath.Join(dir, "ledger.json")
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 19, 1, 0, 0, 0, time.UTC)
	snap := []byte(`{"time":"` + now.Format(time.RFC3339) + `","btc":90000,"source":"Coinbase","markets":[]}` + "\n")
	obs := filepath.Join(dir, "observations.jsonl")
	if err := os.WriteFile(obs, snap, 0600); err != nil {
		t.Fatal(err)
	}
	result, err := run(path, obs)
	if err != nil {
		t.Fatal(err)
	}
	if result.Evidence == nil || result.Evidence.TradingAuthorized || result.Evidence.Status != "RESEARCH_BLOCKED" {
		t.Fatalf("unjustified release: %+v", result.Evidence)
	}
}
