package research100
import("time";"github.com/twappleon/SkewLab/internal/experiment")
type PairedAsOfAudit struct{
 Submitted int `json:"submitted"`
 PreExpirySnapshots int `json:"pre_expiry_snapshots"`
 PublishedOutcomeRows int `json:"published_outcome_rows"`
 FutureCaptures int `json:"future_captures"`
 FutureLabels int `json:"future_labels"`
 InvalidSnapshots int `json:"invalid_snapshots"`
 Status string `json:"status"`
}
func AuditPairedAsOf(items []experiment.Sample,asOf time.Time)PairedAsOfAudit{
 out:=PairedAsOfAudit{Submitted:len(items),Status:"PRE_RESOLUTION_SNAPSHOT_AUDIT_ONLY"}
 for _,s:=range items{
  if s.CapturedAt.After(asOf){out.FutureCaptures++;continue}
  if !ValidatePairedSnapshot(s,asOf).Valid{out.InvalidSnapshots++;continue}
  out.PreExpirySnapshots++
  if s.Winner!=""&&s.ResolvedAt.After(asOf){out.FutureLabels++;continue}
  if s.Winner!=""{out.PublishedOutcomeRows++}
 }
 return out
}
