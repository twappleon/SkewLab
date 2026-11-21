package research100
import("time")
type ReplayEntryAudit struct{
 Observations int `json:"observations"`
 PricedEntryCandidates int `json:"priced_entry_candidates"`
 InfeasibleEntryCosts int `json:"infeasible_entry_costs"`
 Status string `json:"status"`
 TradingAuthorized bool `json:"trading_authorized"`
}
func AssessReplayEntryCandidates(d Dataset,asOf time.Time)ReplayEntryAudit{
 out:=ReplayEntryAudit{Status:"OBSERVATIONS_ONLY_NOT_FILLED_ORDERS",TradingAuthorized:false}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  out.Observations++
  c:=CostToResolution(o.Ask,.002,25,0,o.CapturedAt)
  if c.Status=="COST_MODEL_ONLY_NOT_EXECUTABLE"{out.PricedEntryCandidates++}else{out.InfeasibleEntryCosts++}
 }
 return out
}
