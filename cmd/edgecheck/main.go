// edgecheck is an offline, read-only walk-forward research tool.
// It consumes only the local evidence file and cannot submit orders.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/twappleon/SkewLab/internal/edge"
	"os"
)

func main() {
	file := flag.String("file", "", "edge-research.json, created on a persistent volume")
	slip := flag.Float64("slippage", .002, "modeled per-share entry slippage reserve")
	reserve := flag.Float64("extra-fee-bps", 25, "conservative entry fee reserve beyond official platform taker fee")
	flag.Parse()
	if *file == "" {
		fmt.Fprintln(os.Stderr, "--file required; no samples are fabricated")
		os.Exit(2)
	}
	if *slip < 0 || *reserve < 0 {
		fmt.Fprintln(os.Stderr, "cost assumptions must be nonnegative")
		os.Exit(2)
	}
	raw, err := os.ReadFile(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not read evidence:", err)
		os.Exit(1)
	}
	var samples []edge.Candidate
	if err = json.Unmarshal(raw, &samples); err != nil {
		fmt.Fprintln(os.Stderr, "invalid evidence JSON:", err)
		os.Exit(1)
	}
	report := edge.WalkForwardCheck(samples, *slip, *reserve)
	out := struct {
		PaperOnly         bool             `json:"paper_only"`
		TradingAuthorized bool             `json:"trading_authorized"`
		Research          edge.WalkForward `json:"research"`
		Warning           string           `json:"warning"`
	}{true, false, report, "Historical settle-to-expiry research only. Samples are not independently audited Chainlink outcomes; do not enable paper or live entries based on this report."}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err = enc.Encode(out); err != nil {
		os.Exit(1)
	}
}
