// Package edge collects pre-resolution, price-aware BTC market observations.
// It never sends orders. Outcomes are labeled only after official Gamma resolution.
package edge

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/twappleon/SkewLab/internal/fees"
	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/storage"
	"github.com/twappleon/SkewLab/internal/strategy"
)

const MinEvidence = 100
const MaxRecords = 50000
const maxAge = 12 * time.Second

// Candidate is a SINGLE pre-settlement observation per market. The final outcome
// and resolved timestamp are populated only after official result verification.
type Candidate struct {
	Slug          string    `json:"slug"`
	ConditionID   string    `json:"condition_id"`
	WindowSeconds int       `json:"window_seconds"`
	Side          string    `json:"side"`
	CapturedAt    time.Time `json:"captured_at"`
	MarketEnd     time.Time `json:"market_end"`
	Ask           float64   `json:"ask"`
	Bid           float64   `json:"bid"`
	AskDepth      float64   `json:"ask_depth"`
	BidDepth      float64   `json:"bid_depth"`
	QuoteAt       time.Time `json:"quote_at"`
	MomentumPct   float64   `json:"momentum_pct"`
	Skew          float64   `json:"skew"`
	Winner        string    `json:"winner,omitempty"`
	ResolvedAt    time.Time `json:"resolved_at,omitempty"`
	OutcomeSource string    `json:"outcome_source,omitempty"`
}

type RecordStatus struct {
	Total               int    `json:"total"`
	Resolved            int    `json:"resolved"`
	Pending             int    `json:"pending"`
	MinPerCohort        int    `json:"min_per_cohort"`
	PaperEntriesEnabled bool   `json:"paper_entries_enabled"`
	ResearchOnly        bool   `json:"research_only"`
	LastError           string `json:"last_error,omitempty"`
}

type Recorder struct {
	mu      sync.RWMutex
	path    string
	records []Candidate
	bySlug  map[string]int
	err     error
}

func NewRecorder(dir string) *Recorder {
	r := &Recorder{path: filepath.Join(dir, "edge-research.json"), bySlug: map[string]int{}}
	b, e := os.ReadFile(r.path)
	if os.IsNotExist(e) {
		return r
	}
	if e != nil {
		r.err = fmt.Errorf("research file unreadable: %w", e)
		return r
	}
	if e = json.Unmarshal(b, &r.records); e != nil {
		r.err = fmt.Errorf("research file corrupt: %w", e)
		r.records = nil
		return r
	}
	if len(r.records) > MaxRecords {
		r.err = errors.New("research archive exceeds maximum capacity")
		return r
	}
	for i, c := range r.records {
		if c.Slug == "" || c.ConditionID == "" || c.CapturedAt.IsZero() {
			r.err = errors.New("research archive contains invalid identity")
			return r
		}
		if _, exists := r.bySlug[c.Slug]; exists {
			r.err = errors.New("research archive contains duplicate market")
			return r
		}
		r.bySlug[c.Slug] = i
	}
	return r
}
func (r *Recorder) save() error {
	b, e := json.Marshal(r.records)
	if e != nil {
		return e
	}
	return storage.Atomic(r.path, b)
}
func validQuote(q polymarket.Quote, now time.Time) bool {
	if !(q.Ask > 0 && q.Ask < 1 && q.Bid > 0 && q.Bid <= q.Ask && q.AskDepth > 0 && q.BidDepth > 0) {
		return false
	}
	if q.UpdatedAt.IsZero() || q.UpdatedAt.After(now.Add(3*time.Second)) || now.Sub(q.UpdatedAt) > maxAge {
		return false
	}
	return true
}
func finite(n float64) bool { return !math.IsNaN(n) && !math.IsInf(n, 0) }

// Capture observes a candidate despite PAPER_ENTRIES_ENABLED being disabled.
// It must NOT infer outcome from spot price and does NOT book a fill or profit.
func (r *Recorder) Capture(m polymarket.Market, s strategy.Result, now time.Time) (bool, error) {
	if r == nil {
		return false, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return false, r.err
	}
	if !s.Ready || (s.Direction != "UP" && s.Direction != "DOWN") || !m.Active || m.ConditionID == "" || m.Slug == "" || !now.Before(m.End.Add(-65*time.Second)) || !finite(s.MomentumPct) || !finite(s.Skew) {
		return false, nil
	}
	if m.Seconds != 300 && m.Seconds != 900 {
		return false, nil
	}
	if _, ok := r.bySlug[m.Slug]; ok {
		return false, nil
	}
	if len(r.records) >= MaxRecords {
		r.err = errors.New("edge research archive full; backup/archive before continuing")
		return false, r.err
	}
	q := m.Up
	if s.Direction == "DOWN" {
		q = m.Down
	}
	if !validQuote(q, now) {
		return false, nil
	}
	c := Candidate{Slug: m.Slug, ConditionID: m.ConditionID, WindowSeconds: m.Seconds, Side: s.Direction, CapturedAt: now, MarketEnd: m.End, Ask: q.Ask, Bid: q.Bid, AskDepth: q.AskDepth, BidDepth: q.BidDepth, QuoteAt: q.UpdatedAt, MomentumPct: s.MomentumPct, Skew: s.Skew}
	r.records = append(r.records, c)
	if e := r.save(); e != nil {
		r.records = r.records[:len(r.records)-1]
		r.err = fmt.Errorf("research save failed: %w", e)
		return false, r.err
	}
	r.bySlug[c.Slug] = len(r.records) - 1
	return true, nil
}

// Due returns pending candidates safely; the caller fetches Gamma outside locks.
func (r *Recorder) Due(now time.Time, limit int) []Candidate {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.err != nil {
		return nil
	}
	if limit < 1 || limit > 32 {
		limit = 8
	}
	result := make([]Candidate, 0, limit)
	for _, c := range r.records {
		if c.Winner != "" || c.MarketEnd.IsZero() || now.Before(c.MarketEnd.Add(30*time.Second)) {
			continue
		}
		result = append(result, c)
		if len(result) >= limit {
			break
		}
	}
	return result
}
func (r *Recorder) Resolve(slug, conditionID, winner string, now time.Time) error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return r.err
	}
	i, ok := r.bySlug[slug]
	if !ok {
		return errors.New("unknown research candidate")
	}
	c := &r.records[i]
	if c.ConditionID != conditionID || (winner != "UP" && winner != "DOWN") || now.Before(c.MarketEnd) {
		return errors.New("resolution identity or time mismatch")
	}
	if c.Winner != "" {
		if c.Winner != winner {
			return errors.New("contradictory resolved outcome")
		}
		return nil
	}
	c.Winner = winner
	c.ResolvedAt = now
	c.OutcomeSource = "Polymarket Gamma resolved market, not independent Chainlink proof"
	if e := r.save(); e != nil {
		c.Winner = ""
		c.ResolvedAt = time.Time{}
		c.OutcomeSource = ""
		r.err = fmt.Errorf("resolution journal write failure: %w", e)
		return r.err
	}
	return nil
}
func (r *Recorder) Snapshot() []Candidate {
	if r == nil {
		return []Candidate{}
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]Candidate{}, r.records...)
}
func (r *Recorder) Status() RecordStatus {
	st := RecordStatus{MinPerCohort: MinEvidence, ResearchOnly: true, PaperEntriesEnabled: false}
	if r == nil {
		return st
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	st.Total = len(r.records)
	for _, c := range r.records {
		if c.Winner != "" {
			st.Resolved++
		} else {
			st.Pending++
		}
	}
	if r.err != nil {
		st.LastError = "research evidence unavailable; check private service logs"
	}
	return st
}

// Decision estimates a *settle-to-expiry* research edge. It does NOT justify
// the legacy strategy's early exit or give permission to place paper/live orders.
type Decision struct {
	Status               string   `json:"status"`
	Samples              int      `json:"samples"`
	Winners              int      `json:"winners"`
	WilsonLower95        *float64 `json:"wilson_lower_95,omitempty"`
	EntryCostPerShare    float64  `json:"entry_cost_per_share"`
	WorstCaseExpectedROI *float64 `json:"lower_bound_expected_roi,omitempty"`
	Explanation          string   `json:"explanation"`
}

func priceBucket(p float64) int {
	v := int(math.Floor(p * 5))
	if v < 0 {
		return 0
	}
	if v > 4 {
		return 4
	}
	return v
}
func wilsonLower(wins, total int) float64 {
	if total <= 0 {
		return 0
	}
	p := float64(wins) / float64(total)
	n := float64(total)
	z := 1.96
	den := 1 + z*z/n
	center := p + z*z/(2*n)
	distance := z * math.Sqrt(p*(1-p)/n+z*z/(4*n*n))
	return math.Max(0, (center-distance)/den)
}
func EvaluatePast(history []Candidate, now time.Time, window int, side string, ask, slippage, extraBps float64) Decision {
	out := Decision{Status: "NO_EDGE_EVIDENCE", Explanation: "Insufficient independently resolved, time-prior samples; research-only."}
	if now.IsZero() || !(side == "UP" || side == "DOWN") || (window != 300 && window != 900) || ask <= 0 || ask >= 1 || !finite(ask) || slippage < 0 || !finite(slippage) {
		out.Status = "INVALID_INPUT"
		return out
	}
	px := ask + slippage
	if px >= 1 {
		out.Status = "INVALID_INPUT"
		return out
	}
	out.EntryCostPerShare = fees.BufferedCost(1, px, extraBps, now)
	if !finite(out.EntryCostPerShare) || out.EntryCostPerShare <= 0 {
		out.Status = "INVALID_INPUT"
		return out
	}
	bucket := priceBucket(ask)
	seen := map[string]bool{}
	for _, c := range history {
		if c.WindowSeconds != window || c.Side != side || priceBucket(c.Ask) != bucket || c.Winner == "" || c.ResolvedAt.IsZero() || !c.ResolvedAt.Before(now) || !c.CapturedAt.Before(c.ResolvedAt) || c.Slug == "" || seen[c.Slug] {
			continue
		}
		seen[c.Slug] = true
		out.Samples++
		if c.Winner == side {
			out.Winners++
		}
	}
	if out.Samples < MinEvidence {
		return out
	}
	lower := wilsonLower(out.Winners, out.Samples)
	ev := lower/out.EntryCostPerShare - 1
	out.WilsonLower95 = &lower
	out.WorstCaseExpectedROI = &ev
	if ev < .02 {
		out.Status = "NO_CONSERVATIVE_EDGE"
		out.Explanation = "The lower 95% win-rate bound does not clear the 2% entry-cost buffer. Observational research only."
		return out
	}
	out.Status = "RESEARCH_CANDIDATE_ONLY"
	out.Explanation = "Historical confidence bound clears modeled entry costs under settle-to-expiry assumptions; requires independent prospective validation, calibration and execution checks. No trading authorization."
	return out
}

// WalkForward evaluates each resolved candidate using ONLY results resolved
// before the candidate was captured. Observations are clustered by market ID.
type WalkForward struct {
	Samples                    int     `json:"samples"`
	Resolved                   int     `json:"resolved"`
	EligibleResearchCandidates int     `json:"eligible_research_candidates"`
	EligibleWinners            int     `json:"eligible_winners"`
	SettleToExpiryPaperPnL     float64 `json:"settle_to_expiry_hypothetical_pnl_usd"`
	Status                     string  `json:"status"`
}

func WalkForwardCheck(data []Candidate, slippage, extraBps float64) WalkForward {
	out := WalkForward{Samples: len(data), Status: "INSUFFICIENT_OUT_OF_SAMPLE_EVIDENCE"}
	items := append([]Candidate{}, data...)
	sort.Slice(items, func(i, j int) bool { return items[i].CapturedAt.Before(items[j].CapturedAt) })
	var prior []Candidate
	for _, c := range items {
		if c.Winner == "" || c.ResolvedAt.IsZero() {
			continue
		}
		out.Resolved++
		decision := EvaluatePast(prior, c.CapturedAt, c.WindowSeconds, c.Side, c.Ask, slippage, extraBps)
		if decision.Status == "RESEARCH_CANDIDATE_ONLY" {
			out.EligibleResearchCandidates++
			pay := 0.0
			if c.Winner == c.Side {
				pay = 1
				out.EligibleWinners++
			}
			out.SettleToExpiryPaperPnL += pay - decision.EntryCostPerShare
		}
		// Even if c resolves after later market starts, it only contributes to a
		// future decision when resolved_at strictly precedes that candidate's time.
		prior = append(prior, c)
	}
	out.SettleToExpiryPaperPnL = math.Round(out.SettleToExpiryPaperPnL*10000) / 10000
	if out.EligibleResearchCandidates >= 30 {
		out.Status = "RESEARCH_ONLY_NOT_VALIDATED_FOR_TRADING"
	}
	return out
}
