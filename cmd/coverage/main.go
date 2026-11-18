// coverage measures recorded real-market observation completeness. It does not
// estimate profitability, nor call synthetic demo data real.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/twappleon/SkewLab/internal/polymarket"
)

type Observation struct {
	Time    time.Time           `json:"time"`
	BTC     float64             `json:"btc"`
	Source  string              `json:"source"`
	Markets []polymarket.Market `json:"markets"`
}

type Report struct {
	First              time.Time      `json:"first_recorded_at"`
	Last               time.Time      `json:"last_recorded_at"`
	Sources            map[string]int `json:"spot_sources"`
	Records            int            `json:"valid_real_observations"`
	Invalid            int            `json:"invalid_or_synthetic_records"`
	FreshPairs         int            `json:"real_observations_with_fresh_market_pair"`
	LargeGaps          int            `json:"gaps_over_10_seconds"`
	LongestGapSeconds  float64        `json:"longest_gap_seconds"`
	SampleDurationDays float64        `json:"first_to_last_days"`
	CalendarDays       int            `json:"distinct_utc_calendar_days"`
	ByDay              map[string]int `json:"samples_by_utc_date"`
	Warnings           []string       `json:"limitations"`
}

func summarize(obs []Observation) Report {
	r := Report{Sources: map[string]int{}, ByDay: map[string]int{}, Warnings: []string{
		"Coverage is sampled record availability, not exchange uptime or profitability.",
		"Real BTC spot prices are not Chainlink TWAP settlement proof.",
		"Freshness is checked per saved observation; no actual maker fills or executable trades are implied.",
	}}
	var prev time.Time
	for _, o := range obs {
		if o.Time.IsZero() || o.BTC <= 0 || strings.Contains(strings.ToUpper(o.Source), "SYNTHETIC") || o.Source == "" {
			r.Invalid++
			continue
		}
		r.Records++
		r.Sources[o.Source]++
		r.ByDay[o.Time.UTC().Format("2006-01-02")]++
		if r.First.IsZero() || o.Time.Before(r.First) {
			r.First = o.Time
		}
		if o.Time.After(r.Last) {
			r.Last = o.Time
		}
		if !prev.IsZero() && o.Time.After(prev) {
			gap := o.Time.Sub(prev)
			if gap > 10*time.Second {
				r.LargeGaps++
			}
			if gap.Seconds() > r.LongestGapSeconds {
				r.LongestGapSeconds = gap.Seconds()
			}
		}
		if !prev.IsZero() && o.Time.Before(prev) {
			r.Warnings = append(r.Warnings, "Out-of-order records were detected; check the source segments")
		}
		prev = o.Time
		fresh := false
		for _, m := range o.Markets {
			if m.Active && m.Up.Bid > 0 && m.Up.Ask > m.Up.Bid && m.Down.Bid > 0 && m.Down.Ask > m.Down.Bid &&
				m.Up.BidDepth > 0 && m.Up.AskDepth > 0 && m.Down.BidDepth > 0 && m.Down.AskDepth > 0 &&
				!m.Up.UpdatedAt.IsZero() && !m.Down.UpdatedAt.IsZero() &&
				o.Time.Sub(m.Up.UpdatedAt) <= 12*time.Second && o.Time.Sub(m.Down.UpdatedAt) <= 12*time.Second &&
				m.Up.UpdatedAt.Before(o.Time.Add(3*time.Second)) && m.Down.UpdatedAt.Before(o.Time.Add(3*time.Second)) {
				fresh = true
				break
			}
		}
		if fresh {
			r.FreshPairs++
		}
	}
	r.CalendarDays = len(r.ByDay)
	if !r.First.IsZero() {
		r.SampleDurationDays = r.Last.Sub(r.First).Hours() / 24
	}
	return r
}

func readFiles(pattern string) ([]Observation, error) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, errors.New("no files found: " + pattern)
	}
	// Files rotate to observations.jsonl.<UTC timestamp>; sorting by names
	// puts current observations.jsonl first, so sort parsed records afterwards.
	var out []Observation
	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		s := bufio.NewScanner(f)
		s.Buffer(make([]byte, 64*1024), 5<<20)
		for s.Scan() {
			var o Observation
			if e := json.Unmarshal(s.Bytes(), &o); e != nil {
				out = append(out, Observation{})
				continue
			}
			out = append(out, o)
		}
		err = s.Err()
		_ = f.Close()
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Time.Before(out[j].Time) })
	return out, nil
}

func main() {
	pattern := flag.String("glob", "./data/observations.jsonl*", "real observation segments (glob)")
	flag.Parse()
	observations, err := readFiles(*pattern)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	r := summarize(observations)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if r.Records == 0 {
		os.Exit(1)
	}
}
