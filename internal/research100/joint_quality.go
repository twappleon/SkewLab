package research100
import("time")
type JointQuoteQuality struct{
 Observations int `json:"observations"`
 StaleOverFiveSeconds int `json:"stale_over_five_seconds"`
 WideOverThreeCents int `json:"wide_over_three_cents"`
 BothProblems int `json:"both_problems"`
 Status string `json:"status"`
}
func AssessJointQuoteQuality(d Dataset,asOf time.Time)JointQuoteQuality{
 out:=JointQuoteQuality{Status:"JOINT_OBSERVATION_ONLY"}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  out.Observations++
  stale:=o.CapturedAt.Sub(o.QuoteAt)>5*time.Second
  wide:=o.Ask-o.Bid>.03+1e-9
  if stale{out.StaleOverFiveSeconds++}
  if wide{out.WideOverThreeCents++}
  if stale&&wide{out.BothProblems++}
 }
 return out
}
