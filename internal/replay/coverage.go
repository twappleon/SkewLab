package replay

import "strings"

// QuoteCoverage counts ONLY the snapshots actually present on disk. It does
// not claim uptime, expected market availability, or missing-market causation.
type QuoteCoverage struct {
	RecordedFrames         int    `json:"recorded_frames"`
	FramesWithActiveMarket int    `json:"frames_with_active_market"`
	ValidFreshPairs        int    `json:"valid_fresh_pairs"`
	InvalidOrStalePairs    int    `json:"invalid_or_stale_pairs"`
	MissingPairs           int    `json:"missing_pairs"`
	TradeEvidenceReady     bool   `json:"trade_evidence_ready"`
	Caveat                 string `json:"caveat"`
}

func AssessQuoteCoverage(frames []Frame) QuoteCoverage {
	c := QuoteCoverage{RecordedFrames: len(frames), TradeEvidenceReady: false, Caveat: "Only recorded frames are counted; unobserved markets and unavailable periods have no inferred values. Fresh quotes do not prove executable fills."}
	for _, f := range frames {
		present := false
		active := false
		valid := false
		for _, m := range f.Markets {
			if !strings.HasPrefix(m.Slug, "btc-updown-5m-") && !strings.HasPrefix(m.Slug, "btc-updown-15m-") {
				continue
			}
			present = true
			if !m.Active {
				continue
			}
			active = true
			if validQuote(m.Up, f.Time) && validQuote(m.Down, f.Time) {
				valid = true
				break
			}
		}
		if active {
			c.FramesWithActiveMarket++
		}
		if valid {
			c.ValidFreshPairs++
		} else if present {
			c.InvalidOrStalePairs++
		} else {
			c.MissingPairs++
		}
	}
	return c
}
