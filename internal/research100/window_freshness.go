package research100
import("time")
type FreshnessWindow struct{
 Market string `json:"market"`
 Observed int `json:"observed"`
 OverFiveSeconds int `json:"over_five_seconds"`
 LastCaptureAt *time.Time `json:"last_capture_at,omitempty"`
 StaleMonitoring bool `json:"stale_monitoring"`
 Status string `json:"status"`
}
func AssessWindowFreshness(d Dataset,asOf time.Time)[]FreshnessWindow{
 out:=[]FreshnessWindow{{Market:"5m",Status:"NO_OBSERVED_CAPTURES"},{Market:"15m",Status:"NO_OBSERVED_CAPTURES"}}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  i:=0;if o.Market=="15m"{i=1}
  x:=&out[i];x.Observed++
  if o.CapturedAt.Sub(o.QuoteAt)>5*time.Second{x.OverFiveSeconds++}
  if x.LastCaptureAt==nil||o.CapturedAt.After(*x.LastCaptureAt){v:=o.CapturedAt;x.LastCaptureAt=&v}
 }
 for i:=range out{
  x:=&out[i];if x.LastCaptureAt==nil{continue}
  x.StaleMonitoring=asOf.Sub(*x.LastCaptureAt)>24*time.Hour
  x.Status="OBSERVED_ONLY_NOT_FEED_UPTIME"
  if x.StaleMonitoring{x.Status="NO_RECENT_QUALIFYING_CAPTURE"}
 }
 return out
}
