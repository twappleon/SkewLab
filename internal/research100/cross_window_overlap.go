package research100
import("time")
type CrossWindowOverlap struct{
 FiveMinuteCandidates int `json:"five_minute_candidates"`
 FifteenMinuteCandidates int `json:"fifteen_minute_candidates"`
 FiveMinuteNearFifteenMinute int `json:"five_minute_near_fifteen_minute"`
 WindowSeconds int `json:"window_seconds"`
 Status string `json:"status"`
}
func AssessCrossWindowOverlap(d Dataset,asOf time.Time)CrossWindowOverlap{
 out:=CrossWindowOverlap{WindowSeconds:60,Status:"CORRELATED_BTC_EVENT_POSSIBILITY_NOT_INDEPENDENCE_PROOF"}
 var fives,fifteens []time.Time
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  if o.Market=="5m"{fives=append(fives,o.CapturedAt)}else if o.Market=="15m"{fifteens=append(fifteens,o.CapturedAt)}
 }
 out.FiveMinuteCandidates=len(fives);out.FifteenMinuteCandidates=len(fifteens)
 j:=0
 for _,t:=range fives{
  for j<len(fifteens)&&fifteens[j].Before(t.Add(-time.Minute)){j++}
  if j<len(fifteens)&&!fifteens[j].After(t.Add(time.Minute)){out.FiveMinuteNearFifteenMinute++}
 }
 return out
}
