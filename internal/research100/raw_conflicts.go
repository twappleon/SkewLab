package research100
import("time")
type RawOutcomeConflicts struct{
 SeenLabeledIDs int `json:"seen_labeled_ids"`
 DuplicatePublishedLabels int `json:"duplicate_published_labels"`
 ConflictingWinners int `json:"conflicting_winners"`
 Status string `json:"status"`
}
func AssessRawOutcomeConflicts(raw []Observation,asOf time.Time)RawOutcomeConflicts{
 out:=RawOutcomeConflicts{Status:"NO_CONFLICT_DETECTED_IN_SUPPLIED_ARCHIVE"}
 seen:=map[string]string{}
 for _,o:=range raw{
  if o.ID==""||o.Winner==""||o.CapturedAt.After(asOf)||o.ResolvedAt.IsZero()||o.ResolvedAt.After(asOf){continue}
  prior,exists:=seen[o.ID]
  if exists{
   out.DuplicatePublishedLabels++
   if prior!=o.Winner{out.ConflictingWinners++}
  }else{
   seen[o.ID]=o.Winner;out.SeenLabeledIDs++
  }
 }
 if out.ConflictingWinners>0{out.Status="CONTRADICTORY_OUTCOMES_REQUIRE_HUMAN_REVIEW"}
 return out
}
