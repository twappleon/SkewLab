package research100
import("time")
type ExitResearchReview struct{
 Coverage WitnessCoverage `json:"coverage"`
 StopPolicies StopPolicyComparison `json:"stop_policies"`
 Status string `json:"status"`
 TradingAuthorized bool `json:"trading_authorized"`
 Limitations string `json:"limitations"`
}
func ReviewExitEvidence(d Dataset,witnesses []ExitWitness,asOf time.Time)ExitResearchReview{
 out:=ExitResearchReview{TradingAuthorized:false,
  Limitations:"Only independently supplied post-entry quoted bid witnesses are compared. No realized fills, first stop trigger, exchange latency or settlement proof can be inferred."}
 out.Coverage=AssessWitnessCoverage(d,witnesses,asOf)
 out.StopPolicies=CompareStopPolicies(d,witnesses,asOf)
 out.Status="NO_VALID_POST_ENTRY_BOOK_EVIDENCE"
 if out.Coverage.Matched>0{out.Status="OBSERVED_SINGLE_WITNESSES_NOT_FULL_TRADE_REPLAY"}
 return out
}
