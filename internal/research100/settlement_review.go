package research100
import("time")
type SettlementReview struct{
 Status string `json:"status"`
 Frontier LabelFrontier `json:"frontier"`
 PublicationLag LabelQuantiles `json:"publication_lag"`
 PendingAge PendingAgeAudit `json:"pending_age"`
 Progress []SettlementProgress `json:"progress"`
 Blockers []string `json:"blockers"`
 TradingAuthorized bool `json:"trading_authorized"`
}
// ReviewSettlementData only assesses timestamped archive evidence.
// It neither queries Gamma nor authenticates independent settlement oracles.
func ReviewSettlementData(d Dataset,asOf time.Time)SettlementReview{
 r:=SettlementReview{Status:"INSUFFICIENT_PUBLISHED_OUTCOMES",TradingAuthorized:false,Blockers:[]string{}}
 r.Frontier=InspectLabelFrontier(d,asOf)
 r.PublicationLag=AssessLabelQuantiles(d,asOf)
 r.PendingAge=AssessPendingAge(d,asOf)
 r.Progress=AssessSettlementProgress(d,asOf)
 if r.Frontier.Published<100{r.Blockers=append(r.Blockers,"FEWER_THAN_100_PUBLISHED_OUTCOMES")}
 if r.PendingAge.Over30Minutes>0{r.Blockers=append(r.Blockers,"EXPIRED_OUTCOME_PUBLICATION_BACKLOG")}
 if r.Frontier.FutureCaptures>0{r.Blockers=append(r.Blockers,"POST_ASOF_CAPTURE_EXCLUDED")}
 if r.Frontier.NotYetPublished>0{r.Blockers=append(r.Blockers,"UNPUBLISHED_OR_FUTURE_OUTCOME_EXCLUDED")}
 if len(r.Blockers)==0{r.Status="PUBLICATION_ARCHIVE_REVIEW_ONLY_NOT_ORACLE_PROOF"}
 return r
}
