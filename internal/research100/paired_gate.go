package research100
import("time";"github.com/twappleon/SkewLab/internal/experiment")
type PairedEvidenceGate struct{
 PublishedSamples int `json:"published_samples"`
 DistinctPublishedUTCDays int `json:"distinct_published_utc_days"`
 FiveMinuteForward int `json:"five_minute_forward"`
 FifteenMinuteForward int `json:"fifteen_minute_forward"`
 Status string `json:"status"`
 TradingAuthorized bool `json:"trading_authorized"`
}
func CheckPairedEvidenceGate(items []experiment.Sample,asOf time.Time)PairedEvidenceGate{
 out:=PairedEvidenceGate{Status:"INSUFFICIENT_PAIRED_FORWARD_EVIDENCE",TradingAuthorized:false}
 days:=map[string]bool{}
 for _,s:=range items{
  if !ValidatePairedSnapshot(s,asOf).Valid||s.Winner==""||s.ResolvedAt.After(asOf){continue}
  out.PublishedSamples++
  days[s.CapturedAt.UTC().Format("2006-01-02")]=true
 }
 out.DistinctPublishedUTCDays=len(days)
 comparison:=experiment.Compare(items,asOf)
 for _,w:=range comparison.AllWindows{
  if w.Window=="5m"{out.FiveMinuteForward=w.ForwardResolved}
  if w.Window=="15m"{out.FifteenMinuteForward=w.ForwardResolved}
 }
 if out.PublishedSamples>=260&&out.DistinctPublishedUTCDays>=10&&out.FiveMinuteForward>=30&&out.FifteenMinuteForward>=30{
  out.Status="MANUAL_PROSPECTIVE_REVIEW_REQUIRED_NOT_TRADE_APPROVAL"
 }
 return out
}
