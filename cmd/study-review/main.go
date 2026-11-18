// study-review generates a local, read-only, zero-trading research verdict.
// It does not need network access, a wallet, or Railway credentials.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/twappleon/SkewLab/internal/experiment"
	"github.com/twappleon/SkewLab/internal/validation"
	"os"
	"time"
)

func main() {
	file := flag.String("file", "", "path to owner-controlled strategy-comparison-v1.json")
	flag.Parse()
	if *file == "" {
		fmt.Fprintln(os.Stderr, "--file is required")
		os.Exit(2)
	}
	b, e := os.ReadFile(*file)
	if e != nil {
		fmt.Fprintln(os.Stderr, "cannot read research archive:", e)
		os.Exit(1)
	}
	var samples []experiment.Sample
	if e = json.Unmarshal(b, &samples); e != nil {
		fmt.Fprintln(os.Stderr, "invalid research archive:", e)
		os.Exit(1)
	}
	report := validation.BuildReview(samples, time.Now().UTC())
	out := json.NewEncoder(os.Stdout)
	out.SetIndent("", "  ")
	if e = out.Encode(report); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
