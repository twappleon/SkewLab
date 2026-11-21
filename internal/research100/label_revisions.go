package research100
import("time")
type LabelRevisionReview struct{
 ComparablePreviouslyPublished int `json:"comparable_previously_published"`
 RemovedMarketIDs int `json:"removed_market_ids"`
 RemovedPublishedLabels int `json:"removed_published_labels"`
 ChangedWinners int `json:"changed_winners"`
 Status string `json:"status"`
}
func AuditLabelRevisions(before,after Dataset,asOf time.Time)LabelRevisionReview{
 out:=LabelRevisionReview{Status:"NO_REVISIONS_DETECTED_IN_PROVIDED_SNAPSHOTS"}
 next:=map[string]Observation{}
 for _,o:=range after.Items{next[o.ID]=o}
 for _,o:=range before.Items{
  if o.Winner==""||o.ResolvedAt.After(asOf){continue}
  out.ComparablePreviouslyPublished++
  n,ok:=next[o.ID];if !ok{out.RemovedMarketIDs++;continue}
  if n.Winner==""||n.ResolvedAt.After(asOf){out.RemovedPublishedLabels++;continue}
  if n.Winner!=o.Winner{out.ChangedWinners++}
 }
 if out.RemovedMarketIDs+out.RemovedPublishedLabels+out.ChangedWinners>0{out.Status="PUBLISHED_OUTCOME_REVISION_REQUIRES_MANUAL_REVIEW"}
 return out
}
