package replay

// PairQuality reports paired what-if differences only where BOTH policies
// had observable hypothetical exits on the same original position.
// Never infer uncertainty from the raw trade count as independent samples.
type PairQuality struct {
	Window             string   `json:"window"`
	MatchedPairs       int      `json:"matched_pairs"`
	IndependentUTCDays int      `json:"independent_utc_days"`
	NetDifferenceUSD   *float64 `json:"alternative_minus_legacy_usd,omitempty"`
	Status             string   `json:"status"`
	TradingAuthorized  bool     `json:"trading_authorized"`
}

func AssessPairs(r Review) []PairQuality {
	result := make([]PairQuality, 0, len(r.Groups))
	for _, g := range r.Groups {
		p := PairQuality{Window: g.Window, MatchedPairs: g.PairedModeledExits, IndependentUTCDays: g.IndependentUTCDays, Status: "PAIRED_SAMPLE_INSUFFICIENT", TradingAuthorized: false}
		if g.PairedModeledExits > 0 && finite(g.LegacyModeledPnLUSD) && finite(g.AlternativeModeledPnLUSD) {
			delta := cleanMoney(g.AlternativeModeledPnLUSD - g.LegacyModeledPnLUSD)
			p.NetDifferenceUSD = &delta
		}
		if g.PairedModeledExits >= 30 && g.IndependentUTCDays >= 10 {
			p.Status = "EXPLORATORY_PAIRED_SAMPLE_NOT_VALIDATED"
		}
		result = append(result, p)
	}
	return result
}
