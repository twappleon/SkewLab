package research100
import("time")
type ForwardResearchReview struct{
 Status string `json:"status"`
 ChronologicalFold ChronologicalFold `json:"chronological_fold"`
 HeldOutDays HeldOutDays `json:"heldout_days"`
 UTCRegimes []HourRegimeScore `json:"utc_hour_regimes"`
 WeekRegimes []WeekRegimeScore `json:"week_regimes"`
 MomentumBands []MomentumBand `json:"momentum_bands"`
 SignalCoherence SignalCoherence `json:"signal_coherence"`
 Abstentions AbstentionAudit `json:"abstentions"`
 Gate ForwardGate `json:"gate"`
 TradingAuthorized bool `json:"trading_authorized"`
}
// Seven-day cut is fixed relative to report time; genuine prospective
// evidence still needs a separately pre-registered cutoff and out-of-sample
// validation before any human would consider enabling even paper entries.
func ReviewForwardResearch(d Dataset,asOf time.Time)ForwardResearchReview{
 cutoff:=asOf.AddDate(0,0,-7)
 out:=ForwardResearchReview{TradingAuthorized:false}
 out.ChronologicalFold=AssessChronologicalFold(d,cutoff,asOf)
 out.HeldOutDays=AssessHeldOutDays(d,cutoff,asOf)
 out.UTCRegimes=AssessHourRegimeScore(d,asOf)
 out.WeekRegimes=AssessWeekRegimeScore(d,asOf)
 out.MomentumBands=AssessMomentumBands(d,asOf)
 out.SignalCoherence=AssessSignalCoherence(d,asOf)
 out.Abstentions=AssessAbstentions(d,asOf)
 out.Gate=CheckForwardGate(d,asOf)
 out.Status=out.Gate.Status
 return out
}
