package research100

import (
 "time"
)

type ReleaseGate struct {
 Status string `json:"status"`
 Observations int `json:"observations"`
 PublishedLabels int `json:"published_labels"`
 Blockers []string `json:"blockers"`
 TradingAuthorized bool `json:"trading_authorized"`
}

// ReviewReleaseGate is strictly an audit-ready signal for manual review;
// no branch can return true trading authorization.
func ReviewReleaseGate(d Dataset,asOf time.Time) ReleaseGate {
 out:=ReleaseGate{Status:"RESEARCH_NOT_READY",Blockers:[]string{},TradingAuthorized:false}
 integrity:=AssessAsOfIntegrity(d,asOf)
 capture:=CheckCaptureReadiness(d,asOf)
 settlement:=ReviewSettlementData(d,asOf)
 forward:=CheckForwardGate(d,asOf)
 out.Observations=capture.Observations
 out.PublishedLabels=integrity.PublishedLabels
 if len(capture.Blockers)>0{out.Blockers=append(out.Blockers,"CAPTURE_QUALITY_GATE_BLOCKED")}
 if len(settlement.Blockers)>0{out.Blockers=append(out.Blockers,"SETTLEMENT_EVIDENCE_GATE_BLOCKED")}
 if forward.Status!="MANUAL_REVIEW_ONLY_NOT_TRADING_PERMISSION" {out.Blockers=append(out.Blockers,"FORWARD_PROSPECTIVE_GATE_BLOCKED")}
 if integrity.FutureOutcomeLabels>0||integrity.CapturedInFuture>0{out.Blockers=append(out.Blockers,"FUTURE_DATA_EXCLUDED")}
 if len(out.Blockers)==0 {out.Status="MANUAL_RESEARCH_REVIEW_ONLY_NO_TRADING_AUTHORIZATION"}
 return out
}
