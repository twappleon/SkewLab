package research100
import("time";"github.com/twappleon/SkewLab/internal/experiment")
type PairedResearchReview struct{
 Status string `json:"status"`
 AsOfAudit PairedAsOfAudit `json:"as_of_audit"`
 Rules MatchedThreeRules `json:"matched_rules"`
 Gate PairedEvidenceGate `json:"gate"`
 TradingAuthorized bool `json:"trading_authorized"`
 Caveat string `json:"caveat"`
}
func ReviewPairedEvidence(samples []experiment.Sample,asOf time.Time)PairedResearchReview{
 out:=PairedResearchReview{TradingAuthorized:false,Caveat:"Same preselected first fixed-time two-sided quote; modeled one-share hold-to-settle, NOT real order fill or legacy early exit."}
 out.AsOfAudit=AuditPairedAsOf(samples,asOf)
 out.Rules=ComparePairedRules(samples,asOf)
 out.Gate=CheckPairedEvidenceGate(samples,asOf)
 out.Status=out.Gate.Status
 return out
}
