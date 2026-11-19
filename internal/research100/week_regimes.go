package research100
import("time")
type WeekRegimeScore struct{
 Kind string `json:"kind"`
 Samples int `json:"samples"`
 Wins int `json:"wins"`
 MeanBrier *float64 `json:"mean_brier,omitempty"`
 Status string `json:"status"`
}
func AssessWeekRegimeScore(d Dataset,asOf time.Time)[]WeekRegimeScore{
 out:=[]WeekRegimeScore{{Kind:"weekday",Status:"NO_SAMPLES"},{Kind:"weekend",Status:"NO_SAMPLES"}}
 totals:=[2]float64{}
 for _,o:=range d.Items{
  if o.Winner==""||o.CapturedAt.After(asOf)||o.ResolvedAt.After(asOf){continue}
  i:=0;wd:=o.CapturedAt.UTC().Weekday();if wd==time.Saturday||wd==time.Sunday{i=1}
  p:=(o.Bid+o.Ask)/2;y:=0.0;if o.Winner==o.Side{y=1;out[i].Wins++}
  out[i].Samples++;totals[i]+=(p-y)*(p-y)
 }
 for i:=range out{if out[i].Samples>0{v:=totals[i]/float64(out[i].Samples);out[i].MeanBrier=&v;out[i].Status="DESCRIPTIVE_REGIME_ONLY"}}
 return out
}
