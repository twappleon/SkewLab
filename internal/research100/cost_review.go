package research100
import("time")
type CostResearchReview struct{
 Status string `json:"status"`
 FeeTransition FeeTransitionReview `json:"fee_transition"`
 Slippage SlippageSensitivity `json:"slippage"`
 ExtraReserve FeeReserveGrid `json:"extra_reserve"`
 RoundTrips RoundTripStressGrid `json:"roundtrip_stress"`
 StopHeadroom StopHeadroomReview `json:"stop_headroom"`
 RequiredWinRate QuoteBreakevenReview `json:"required_win_rate"`
 SettledHypothesis []SettlementCounterfactual `json:"settled_hypothesis"`
 HypotheticalDay HypotheticalDayLoss `json:"hypothetical_daily_loss"`
 RiskEnvelope StakeEnvelope `json:"risk_envelope"`
 TradingAuthorized bool `json:"trading_authorized"`
}
func ReviewCostResearch(d Dataset,asOf time.Time)CostResearchReview{
 change:=time.Date(2026,11,20,23,59,0,0,time.UTC)
 out:=CostResearchReview{TradingAuthorized:false,Status:"MODELED_COSTS_ONLY_NOT_EXECUTABLE"}
 out.FeeTransition=CompareFeeTransition(.5,change.Add(-time.Second),change)
 out.Slippage=AssessSlippageSensitivity(.5,asOf)
 out.ExtraReserve=AssessFeeReserveGrid(.5,asOf)
 out.RoundTrips=AssessRoundTripStress(d,asOf)
 out.StopHeadroom=AssessStopHeadroom(d,asOf,5)
 out.RequiredWinRate=AssessQuoteBreakEven(d,asOf)
 out.SettledHypothesis=AssessHypotheticalSettlement(d,asOf)
 out.HypotheticalDay=AssessHypotheticalDailyLoss(d,asOf)
 out.RiskEnvelope=CheckStakeEnvelope([]float64{5})
 if out.StopHeadroom.AlreadyExceedsStop>0{out.Status="MODELED_FRICTION_EXCEEDS_LEGACY_STOP"}
 return out
}
