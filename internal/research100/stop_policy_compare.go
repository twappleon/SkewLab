package research100
import("time")
type StopPolicyComparison struct{
 MatchedWitnesses int `json:"matched_witnesses"`
 FivePercentTriggers int `json:"five_percent_triggers"`
 TenPercentTriggers int `json:"ten_percent_triggers"`
 Status string `json:"status"`
 TradingAuthorized bool `json:"trading_authorized"`
}
func CompareStopPolicies(d Dataset,witnesses []ExitWitness,asOf time.Time)StopPolicyComparison{
 pairs:=PairExitEvidence(d,witnesses,asOf)
 out:=StopPolicyComparison{Status:"SINGLE_EXIT_WITNESS_NOT_FULL_PATH_REPLAY",TradingAuthorized:false}
 for _,p:=range pairs.Pairs{
  a:=EvaluateStopLossWitness(p,5);b:=EvaluateStopLossWitness(p,10)
  if !a.Evaluated||!b.Evaluated{continue}
  out.MatchedWitnesses++
  if a.Triggered{out.FivePercentTriggers++}
  if b.Triggered{out.TenPercentTriggers++}
 }
 if out.MatchedWitnesses==0{out.Status="NO_VALID_POST_ENTRY_WITNESSES"}
 return out
}
