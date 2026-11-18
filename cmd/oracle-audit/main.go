package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/twappleon/SkewLab/internal/oracle"
	"os"
	"time"
)

func main() {
	file := flag.String("file", "", "Path to Chainlink TWAP 60s observations JSONL")
	startText := flag.String("start", "", "Market UTC start time, RFC3339")
	minutes := flag.Int("minutes", 5, "Market length: 5 or 15")
	official := flag.String("official", "", "Optional resolved Polymarket outcome UP or DOWN")
	flag.Parse()
	if *file == "" || *startText == "" {
		fmt.Fprintln(os.Stderr, "--file and --start are required")
		os.Exit(2)
	}
	start, err := time.Parse(time.RFC3339, *startText)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	f, err := os.Open(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	ticks := []oracle.Tick{}
	scan := bufio.NewScanner(f)
	scan.Buffer(make([]byte, 4096), 2<<20)
	for scan.Scan() {
		var t oracle.Tick
		if json.Unmarshal(scan.Bytes(), &t) == nil {
			ticks = append(ticks, t)
		}
	}
	if scan.Err() != nil {
		fmt.Fprintln(os.Stderr, scan.Err())
		os.Exit(1)
	}
	result, err := oracle.Compare(ticks, start, time.Duration(*minutes)*time.Minute)
	if err != nil {
		fmt.Fprintln(os.Stderr, "INCOMPLETE:", err)
		os.Exit(1)
	}
	output := map[string]any{"audit": result, "source": "captured Chainlink TWAP 60s, not official settlement", "official_provided": *official}
	if *official != "" {
		output["agrees_with_official"] = (result.Direction == *official)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(output)
	if *official != "" && result.Direction != *official {
		os.Exit(1)
	}
}
