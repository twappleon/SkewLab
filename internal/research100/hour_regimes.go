package research100
import("time")
type HourRegimeScore struct{
 UTCStartHour int `json:"utc_start_hour"`
 Samples int `json:"samples"`
 Brier *float64 `json:"brier,omitempty"`
 Status string `json:"status"`
}
func AssessHourRegimeScore(d Dataset,asOf time.Time)[]HourRegimeScore{
 out:=[]HourRegimeScore{{UTCStartHour:0,Status:"NO_SAMPLES"},{UTCStartHour:6,Status:"NO_SAMPLES"},{UTCStartHour:12,Status:"NO_SAMPLES"},{UTCStartHour:18,Status:"NO_SAMPLES"}}
 totals:=[4]float64{}
 for _,o:=range d.Items{
  if o.Winner==""||o.CapturedAt.After(asOf)||o.ResolvedAt.After(asOf){continue}
  i:=o.CapturedAt.UTC().Hour()/6
  p:=(o.Bid+o.Ask)/2;y:=0.0;if o.Winner==o.Side{y=1}
  totals[i]+=(p-y)*(p-y);out[i].Samples++
 }
 for i:=range out{if out[i].Samples>0{v:=totals[i]/float64(out[i].Samples);out[i].Brier=&v;out[i].Status="SELECTED_SIDE_REGIME_RESEARCH_ONLY"}}
 return out
}
