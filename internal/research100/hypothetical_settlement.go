package research100
import("time")
type SettlementCounterfactual struct{
 Market string `json:"market"`
 Scored int `json:"scored"`
 Wins int `json:"wins"`
 ModeledNetUSD float64 `json:"modeled_net_usd"`
 Status string `json:"status"`
}
func AssessHypotheticalSettlement(d Dataset,asOf time.Time)[]SettlementCounterfactual{
 out:=[]SettlementCounterfactual{{Market:"5m",Status:"SELECTED_SIDE_COUNTERFACTUAL_NOT_FILLS"},{Market:"15m",Status:"SELECTED_SIDE_COUNTERFACTUAL_NOT_FILLS"}}
 for _,o:=range d.Items{
  if o.Winner==""||o.ResolvedAt.After(asOf)||o.CapturedAt.After(asOf){continue}
  c:=CostToResolution(o.Ask,.002,25,0,o.CapturedAt)
  if c.Status!="COST_MODEL_ONLY_NOT_EXECUTABLE"{continue}
  i:=0;if o.Market=="15m"{i=1}
  x:=&out[i];x.Scored++;payout:=0.0;if o.Winner==o.Side{x.Wins++;payout=1}
  x.ModeledNetUSD+=payout-c.ModeledCostUSD
 }
 return out
}
