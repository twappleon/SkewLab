package research100
import("github.com/twappleon/SkewLab/internal/fees";"math";"time")
type RoundTripStressCase struct{
 Name string `json:"name"`
 Observations int `json:"observations"`
 CostOnlyLossAbove5Pct int `json:"cost_only_loss_above_5_pct"`
 MeanModeledLossPct *float64 `json:"mean_modeled_loss_pct,omitempty"`
}
type RoundTripStressGrid struct{
 Cases []RoundTripStressCase `json:"cases"`
 Status string `json:"status"`
}
func AssessRoundTripStress(d Dataset,asOf time.Time)RoundTripStressGrid{
 out:=RoundTripStressGrid{Status:"SAME_BOOK_MODEL_ONLY_NOT_EXECUTABLE"}
 for _,sc:=range []struct{name string;slip,bps float64}{{"base",.002,25},{"double_slippage",.004,25},{"quadruple_reserve",.002,100}}{
  c:=RoundTripStressCase{Name:sc.name};sum:=0.0
  for _,o:=range d.Items{
   if o.CapturedAt.After(asOf){continue}
   px:=o.Ask+sc.slip;exit:=math.Max(.001,o.Bid-sc.slip)
   if px>=1||exit<=0{continue}
   cost:=fees.BufferedCost(1,px,sc.bps,o.CapturedAt)
   proceeds:=fees.BufferedProceeds(1,exit,sc.bps,o.CapturedAt)
   if cost<=0||proceeds<=0||!finite(cost)||!finite(proceeds){continue}
   loss:=100*(cost-proceeds)/cost
   c.Observations++;sum+=loss
   if loss>5{c.CostOnlyLossAbove5Pct++}
  }
  if c.Observations>0{v:=sum/float64(c.Observations);c.MeanModeledLossPct=&v}
  out.Cases=append(out.Cases,c)
 }
 return out
}
