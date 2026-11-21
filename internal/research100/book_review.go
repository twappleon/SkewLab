package research100
import("time")
type BookReview struct{
 Status string `json:"status"`
 Gate BookGate `json:"gate"`
 Depth []DepthFloorAudit `json:"depth"`
 Spread []SpreadAudit `json:"spread"`
 Joint JointQuoteQuality `json:"joint_quality"`
 Overlap CrossWindowOverlap `json:"cross_window_overlap"`
 RoundTrip RoundTripStudy `json:"hypothetical_round_trip"`
 TradingAuthorized bool `json:"trading_authorized"`
}
// ReviewBookEvidence never claims market fills or new open positions.
// Top-of-book depth and quote-age filters are only necessary, not sufficient.
func ReviewBookEvidence(d Dataset,asOf time.Time)BookReview{
 r:=BookReview{TradingAuthorized:false}
 r.Gate=AssessBookGate(d,asOf)
 r.Depth=AssessDepthFloors(d,asOf,20,20)
 r.Spread=AssessObservedSpread(d,asOf)
 r.Joint=AssessJointQuoteQuality(d,asOf)
 r.Overlap=AssessCrossWindowOverlap(d,asOf)
 r.RoundTrip=AssessRoundTripFriction(d,asOf)
 r.Status="INSUFFICIENT_EXECUTION_EVIDENCE"
 if r.Gate.Observations>0&&r.Gate.PassedVisibleFilters>0{
  r.Status="OBSERVED_TOP_BOOK_PASSES_FILTERS_NOT_EXECUTABLE"
 }
 if r.RoundTrip.CostOnlyLossAboveFivePct>0{
  r.Status="ROUND_TRIP_COST_CAN_EXCEED_5_PERCENT_STOP"
 }
 return r
}
