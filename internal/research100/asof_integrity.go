package research100
import("time")
type AsOfIntegrity struct{
 Observations int `json:"observations"`
 CapturedInFuture int `json:"captured_in_future"`
 PublishedLabels int `json:"published_labels"`
 FutureOutcomeLabels int `json:"future_outcome_labels"`
 PendingWithoutOutcome int `json:"pending_without_outcome"`
 Status string `json:"status"`
 TradingAuthorized bool `json:"trading_authorized"`
}
func AssessAsOfIntegrity(d Dataset,asOf time.Time)AsOfIntegrity{
 out:=AsOfIntegrity{Status:"RESEARCH_TIMELINE_ONLY_NOT_TRADING_APPROVAL",TradingAuthorized:false}
 for _,o:=range d.Items{
  out.Observations++
  if o.CapturedAt.After(asOf){out.CapturedInFuture++;continue}
  if o.Winner==""{out.PendingWithoutOutcome++;continue}
  if o.ResolvedAt.After(asOf){out.FutureOutcomeLabels++;continue}
  out.PublishedLabels++
 }
 if out.CapturedInFuture>0||out.FutureOutcomeLabels>0{out.Status="FUTURE_RELATIVE_TO_REPORT_TIME_EXCLUDED"}
 return out
}
