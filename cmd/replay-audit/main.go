// replay-audit is an OFFLINE read-only comparison against saved observations.
// It NEVER submits trading or simulated orders, and cannot invent settlements.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/twappleon/SkewLab/internal/paper"
	"github.com/twappleon/SkewLab/internal/replay"
)

func run(ledgerPath, glob string) (replay.Review, error) {
	if ledgerPath == "" || glob == "" {
		return replay.Review{}, errors.New("--ledger and --glob are required")
	}
	raw, err := os.ReadFile(ledgerPath)
	if err != nil {
		return replay.Review{}, fmt.Errorf("ledger file unavailable: %w", err)
	}
	var l paper.Ledger
	if err = json.Unmarshal(raw, &l); err != nil {
		return replay.Review{}, fmt.Errorf("ledger invalid: %w", err)
	}
	if !(l.Config.StartingCash > 0) || !(l.Config.StopLossPct > 0) {
		return replay.Review{}, errors.New("invalid ledger configuration")
	}
	paths, err := filepath.Glob(glob)
	if err != nil {
		return replay.Review{}, err
	}
	sort.Strings(paths)
	archive, err := replay.Load(paths)
	if err != nil {
		return replay.Review{}, err
	}
	review := replay.BuildReview(l.Positions, l.Config, archive)
	gate := replay.BuildEvidenceGate(&l, archive, review)
	review.Evidence = &gate
	return review, nil
}
func main() {
	ledger := flag.String("ledger", "", "path to an actual private paper-ledger.json")
	files := flag.String("glob", "", "quoted glob of original observations.jsonl segments")
	flag.Parse()
	review, err := run(*ledger, *files)
	if err != nil {
		fmt.Fprintln(os.Stderr, "REPLAY_NOT_VERIFIABLE:", err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err = enc.Encode(review); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
