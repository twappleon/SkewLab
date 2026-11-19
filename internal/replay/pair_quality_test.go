package replay

import "testing"

func TestPairQualityRequiresDaysAndPairs(t *testing.T) {
	r := Review{Groups: []Group{{Window: "5m", PairedModeledExits: 100, IndependentUTCDays: 1, LegacyModeledPnLUSD: -20, AlternativeModeledPnLUSD: -10}, {Window: "15m", PairedModeledExits: 30, IndependentUTCDays: 11, LegacyModeledPnLUSD: -30, AlternativeModeledPnLUSD: -35}}}
	out := AssessPairs(r)
	if out[0].Status != "PAIRED_SAMPLE_INSUFFICIENT" || out[0].NetDifferenceUSD == nil || *out[0].NetDifferenceUSD != 10 || out[0].TradingAuthorized {
		t.Fatal("one-day overfitting passed")
	}
	if out[1].Status != "EXPLORATORY_PAIRED_SAMPLE_NOT_VALIDATED" || out[1].NetDifferenceUSD == nil || *out[1].NetDifferenceUSD != -5 || out[1].TradingAuthorized {
		t.Fatal("positive result fabricated or auto-trade enabled")
	}
}
