package research100
import("time";"github.com/twappleon/SkewLab/internal/experiment")
type PairedPayoff struct{
 Scored bool `json:"scored"`
 Correct bool `json:"correct"`
 NetOneShareUSD float64 `json:"net_one_share_usd"`
 Status string `json:"status"`
}
func ScorePairedPayoff(s experiment.Sample,side string,asOf time.Time)PairedPayoff{
 out:=PairedPayoff{Status:"NO_PUBLISHED_PAIRED_OUTCOME_OR_ABSTAIN"}
 if !ValidatePairedSnapshot(s,asOf).Valid||s.Winner==""||s.ResolvedAt.IsZero()||s.ResolvedAt.After(asOf)||(side!="UP"&&side!="DOWN"){return out}
 ask:=s.UpAsk;if side=="DOWN"{ask=s.DownAsk}
 c:=CostToResolution(ask,.002,25,0,s.CapturedAt)
 if c.Status!="COST_MODEL_ONLY_NOT_EXECUTABLE"{out.Status="INFEASIBLE_HYPOTHETICAL_ENTRY";return out}
 out.Scored=true;out.Correct=side==s.Winner
 if out.Correct{out.NetOneShareUSD=1-c.ModeledCostUSD}else{out.NetOneShareUSD=-c.ModeledCostUSD}
 out.Status="MODELED_HOLD_TO_SETTLEMENT_NOT_FILLED_OR_EARLY_EXIT"
 return out
}
