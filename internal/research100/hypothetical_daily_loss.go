package research100
import("time")
type HypotheticalDayLoss struct{
 LabeledSamples int `json:"labeled_samples"`
 UTCScoredDays int `json:"utc_scored_days"`
 WorstUTCDate string `json:"worst_utc_date,omitempty"`
 WorstDailyNetUSD *float64 `json:"worst_daily_net_usd,omitempty"`
 Status string `json:"status"`
}
func AssessHypotheticalDailyLoss(d Dataset,asOf time.Time)HypotheticalDayLoss{
 out:=HypotheticalDayLoss{Status:"NO_PUBLISHED_COUNTERFACTUAL_SAMPLES"}
 days:=map[string]float64{}
 for _,o:=range d.Items{
  if o.Winner==""||o.ResolvedAt.After(asOf)||o.CapturedAt.After(asOf){continue}
  c:=CostToResolution(o.Ask,.002,25,0,o.CapturedAt)
  if c.Status!="COST_MODEL_ONLY_NOT_EXECUTABLE"{continue}
  date:=o.CapturedAt.UTC().Format("2006-01-02")
  payoff:=0.0;if o.Winner==o.Side{payoff=1}
  days[date]+=payoff-c.ModeledCostUSD;out.LabeledSamples++
 }
 out.UTCScoredDays=len(days)
 for date,pnl:=range days{
  if out.WorstDailyNetUSD==nil||pnl<*out.WorstDailyNetUSD||(pnl==*out.WorstDailyNetUSD&&date<out.WorstUTCDate){
   v:=pnl;out.WorstDailyNetUSD=&v;out.WorstUTCDate=date
  }
 }
 if out.LabeledSamples>0{out.Status="SELECTED_SIDE_DAILY_COUNTERFACTUAL_NOT_REAL_EQUITY"}
 return out
}
