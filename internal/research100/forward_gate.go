package research100
import("time")
type ForwardGate struct{
 Status string `json:"status"`
 ScoredResolved int `json:"scored_resolved"`
 ObservedPublishedDays int `json:"observed_published_days"`
 HypotheticalNetUSD float64 `json:"hypothetical_net_usd"`
 TradingAuthorized bool `json:"trading_authorized"`
}
func CheckForwardGate(d Dataset,asOf time.Time)ForwardGate{
 wf:=RunWalkForward(d,asOf)
 out:=ForwardGate{Status:"INSUFFICIENT_FORWARD_SAMPLES",ScoredResolved:wf.ScoredResolved,HypotheticalNetUSD:wf.HypotheticalNetUSD,TradingAuthorized:false}
 days:=map[string]bool{}
 for _,o:=range d.Items{
  if o.Winner!=""&&!o.ResolvedAt.After(asOf)&&!o.CapturedAt.After(asOf){days[o.CapturedAt.UTC().Format("2006-01-02")]=true}
 }
 out.ObservedPublishedDays=len(days)
 if out.ScoredResolved>=30&&out.ObservedPublishedDays>=10{out.Status="MANUAL_REVIEW_ONLY_NOT_TRADING_PERMISSION"}
 return out
}
