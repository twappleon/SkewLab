package research100
import("time")
type QuoteBreakevenReview struct{
 Observations int `json:"observations"`
 InfeasibleAfterCosts int `json:"infeasible_after_costs"`
 AboveSeventyFivePctRequired int `json:"above_75_pct_required"`
 Status string `json:"status"`
}
func AssessQuoteBreakEven(d Dataset,asOf time.Time)QuoteBreakevenReview{
 out:=QuoteBreakevenReview{Status:"MODEL_WIN_RATE_REQUIREMENT_NOT_OBSERVED_TRUE_PROBABILITY"}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  c:=CostToResolution(o.Ask,.002,25,2,o.CapturedAt)
  out.Observations++
  if c.Status!="COST_MODEL_ONLY_NOT_EXECUTABLE"{out.InfeasibleAfterCosts++;continue}
  if c.MinimumWinRateWithBufferPct>75{out.AboveSeventyFivePctRequired++}
 }
 return out
}
