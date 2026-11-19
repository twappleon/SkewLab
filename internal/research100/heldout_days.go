package research100
import("time")
type HeldOutDays struct{
 CapturedAfterCutoff int `json:"captured_after_cutoff"`
 ScorableAtAsOf int `json:"scorable_at_as_of"`
 DistinctUTCScoreDays int `json:"distinct_utc_score_days"`
 Status string `json:"status"`
}
func AssessHeldOutDays(d Dataset,cutoff,asOf time.Time)HeldOutDays{
 out:=HeldOutDays{Status:"INVALID_HELDOUT_INTERVAL"}
 if cutoff.IsZero()||asOf.IsZero()||!cutoff.Before(asOf){return out}
 days:=map[string]bool{}
 for _,o:=range d.Items{
  if o.CapturedAt.Before(cutoff)||o.CapturedAt.After(asOf){continue}
  out.CapturedAfterCutoff++
  if o.Winner!=""&&!o.ResolvedAt.After(asOf){
   out.ScorableAtAsOf++;days[o.CapturedAt.UTC().Format("2006-01-02")]=true
  }
 }
 out.DistinctUTCScoreDays=len(days)
 out.Status="INSUFFICIENT_DISTINCT_HELDOUT_DAYS"
 if out.ScorableAtAsOf>=30&&out.DistinctUTCScoreDays>=10{out.Status="REVIEW_ONLY_DISTINCT_HELDOUT_DAYS"}
 return out
}
