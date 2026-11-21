package research100
import("time")
type FeeTransitionReview struct{
 OldModeledCostUSD float64 `json:"old_modeled_cost_usd"`
 NewModeledCostUSD float64 `json:"new_modeled_cost_usd"`
 ReducedAfterTransition bool `json:"reduced_after_transition"`
 Status string `json:"status"`
}
func CompareFeeTransition(ask float64,before,after time.Time)FeeTransitionReview{
 out:=FeeTransitionReview{Status:"INVALID_FEE_TRANSITION_REQUEST"}
 if !before.Before(after){return out}
 a:=CostToResolution(ask,.002,25,0,before)
 b:=CostToResolution(ask,.002,25,0,after)
 if a.Status!="COST_MODEL_ONLY_NOT_EXECUTABLE"||b.Status!="COST_MODEL_ONLY_NOT_EXECUTABLE"{return out}
 out.OldModeledCostUSD=a.ModeledCostUSD;out.NewModeledCostUSD=b.ModeledCostUSD
 out.ReducedAfterTransition=b.ModeledCostUSD<a.ModeledCostUSD
 out.Status="DATED_FEE_MODEL_ONLY_NOT_EXCHANGE_CONFIRMATION"
 return out
}
