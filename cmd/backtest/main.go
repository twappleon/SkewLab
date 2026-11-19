package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/twappleon/SkewLab/internal/paper"
	"github.com/twappleon/SkewLab/internal/quantv2"
	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/strategy"
	"os"
	"time"
)

type observation struct {
	Time    time.Time           `json:"time"`
	BTC     float64             `json:"btc"`
	Source  string              `json:"source"`
	Markets []polymarket.Market `json:"markets"`
}

func main() {
	file := flag.String("file", "./data/observations.jsonl", "recorded observations, not synthetic bars")
	mode := flag.String("strategy", "baseline", "baseline or quantv2; simulation only")
	flag.Parse()
	if *mode != "baseline" && *mode != "quantv2" { fmt.Fprintln(os.Stderr, "unknown replay strategy"); os.Exit(2) }
	f, err := os.Open(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	cfg := paper.DefaultConfig()
	if *mode == "quantv2" { cfg = quantv2.Config() }
	l := paper.New(cfg)
	var ticks []strategy.Tick
	total := 0
	rejected := 0
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), 5<<20)
	for scanner.Scan() {
		var o observation
		if json.Unmarshal(scanner.Bytes(), &o) != nil || o.BTC <= 0 || o.Time.IsZero() {
			rejected++
			continue
		}
		if len(ticks) > 0 && ticks[len(ticks)-1].Source != o.Source {
			ticks = nil
		}
		ticks = append(ticks, strategy.Tick{Time: o.Time, Price: o.BTC, Source: o.Source})
		begin := 0
		for begin < len(ticks) && o.Time.Sub(ticks[begin].Time) > 65*time.Second {
			begin++
		}
		ticks = ticks[begin:]
		s := strategy.Calculate(ticks, o.Time, 0.02, 0.2)
		for _, m := range o.Markets {
			if m.Active {
				candidate := s
				if *mode == "quantv2" { candidate = quantv2.Gate(s,m,cfg,o.Time) }
				l.Evaluate(m, candidate, o.Time)
			}
		}
		l.RecordEquity(o.Time)
		total++
	}
	if scanner.Err() != nil {
		fmt.Fprintln(os.Stderr, scanner.Err())
		os.Exit(1)
	}
	if total == 0 {
		fmt.Fprintln(os.Stderr, "No valid observations. Collect live data first; the backtester does not invent prices.")
		os.Exit(1)
	}
	fmt.Printf("Replay strategy: %s (PAPER ONLY)\\n",*mode)
	fmt.Printf("Replay observations: %d (invalid: %d)\nStarting USD: %.2f\nPaper equity USD: %.2f\nRealized P&L USD: %.2f\nClosed trades: %d; wins: %d; open/unresolved: %d\nMax drawdown (observed): %.2f%%\n", total, rejected, l.Config.StartingCash, l.CurrentEquity(), l.RealizedPnL, l.TotalClosed, l.Wins, l.OpenPositions(), l.MaxDrawdownPct)
	fmt.Println("CAUTION: tick-sampled order books only; no queue-position fill model; no oracle settlement; stale markets not valued. NOT live trading proof.")
}
