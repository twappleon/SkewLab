package research100
import("github.com/twappleon/SkewLab/internal/fees";"time")
type RoundTripStudy struct{
 Observations int `json:"observations"`
 CostOnlyLossAboveFivePct int `json:"cost_only_loss_above_five_pct"`
 MeanModeledLossPct *float64 `json:"mean_modeled_loss_pct,omitempty"`
 Status string `json:"status"`
}
func AssessRoundTripFriction(d Dataset,asOf time.Time)RoundTripStudy{
 out:=RoundTripStudy{Status:"HYPOTHETICAL_UNCHANGED_TOP_BOOK_NOT_REAL_FILLS"}
 total:=0.0
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  cost:=CostToResolution(o.Ask,.002,25,0,o.CapturedAt)
  if cost.Status!="COST_MODEL_ONLY_NOT_EXECUTABLE"{continue}
  exit:=o.Bid-.002;if exit<=0{continue}
  proceeds:=fees.BufferedProceeds(1,exit,25,o.CapturedAt)
  if !finite(proceeds)||proceeds<=0{continue}
  loss:=100*(cost.ModeledCostUSD-proceeds)/cost.ModeledCostUSD
  if !finite(loss){continue}
  out.Observations++;total+=loss
  if loss>5{out.CostOnlyLossAboveFivePct++}
 }
 if out.Observations>0{v:=total/float64(out.Observations);out.MeanModeledLossPct=&v}
 return out
}
