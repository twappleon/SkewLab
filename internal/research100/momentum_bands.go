package research100
import("math";"time")
type MomentumBand struct{
 Band string `json:"band"`
 Published int `json:"published"`
 Correct int `json:"correct"`
 Accuracy *float64 `json:"accuracy,omitempty"`
 Status string `json:"status"`
}
func AssessMomentumBands(d Dataset,asOf time.Time)[]MomentumBand{
 out:=[]MomentumBand{{Band:"abs_below_0.02_pct",Status:"NO_PUBLISHED_SAMPLES"},{Band:"abs_0.02_to_0.05_pct",Status:"NO_PUBLISHED_SAMPLES"},{Band:"abs_above_0.05_pct",Status:"NO_PUBLISHED_SAMPLES"}}
 for _,o:=range d.Items{
  if o.Winner==""||o.CapturedAt.After(asOf)||o.ResolvedAt.After(asOf){continue}
  magnitude:=math.Abs(o.Momentum);i:=0
  if magnitude>=.02{i=1}
  if magnitude>.05{i=2}
  out[i].Published++;if o.Winner==o.Side{out[i].Correct++}
 }
 for i:=range out{if out[i].Published>0{v:=float64(out[i].Correct)/float64(out[i].Published);out[i].Accuracy=&v;out[i].Status="DESCRIPTIVE_SELECTED_SIDE_ONLY"}}
 return out
}
