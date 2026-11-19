package research100
import("time")
type SpreadAudit struct{
 Market string `json:"market"`
 Observed int `json:"observed"`
 WideOverThreeCents int `json:"wide_over_three_cents"`
 MeanSpread *float64 `json:"mean_spread,omitempty"`
 Note string `json:"note"`
}
func AssessObservedSpread(d Dataset,asOf time.Time)[]SpreadAudit{
 out:=[]SpreadAudit{
  {Market:"5m",Note:"Quoted spread only; full depth and executable VWAP unverified"},
  {Market:"15m",Note:"Quoted spread only; full depth and executable VWAP unverified"},
 }
 totals:=[2]float64{}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  i:=0;if o.Market=="15m"{i=1}
  sp:=o.Ask-o.Bid
  if !finite(sp)||sp<0{continue}
  out[i].Observed++;totals[i]+=sp
  if sp>.03+1e-9{out[i].WideOverThreeCents++}
 }
 for i:=range out{if out[i].Observed>0{v:=totals[i]/float64(out[i].Observed);out[i].MeanSpread=&v}}
 return out
}
