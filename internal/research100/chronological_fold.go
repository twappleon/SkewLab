package research100
import("time")
type ChronologicalFold struct{
 TrainPublishedBeforeCutoff int `json:"train_published_before_cutoff"`
 HeldOutPublishedAtAsOf int `json:"held_out_published_at_as_of"`
 UnavailableTrainLabels int `json:"unavailable_train_labels"`
 Status string `json:"status"`
}
func AssessChronologicalFold(d Dataset,cutoff,asOf time.Time)ChronologicalFold{
 out:=ChronologicalFold{Status:"INVALID_TIME_SPLIT"}
 if cutoff.IsZero()||asOf.IsZero()||!cutoff.Before(asOf){return out}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  if o.CapturedAt.Before(cutoff){
   if o.Winner!=""&&o.ResolvedAt.Before(cutoff){out.TrainPublishedBeforeCutoff++}else{out.UnavailableTrainLabels++}
   continue
  }
  if o.Winner!=""&&!o.ResolvedAt.After(asOf){out.HeldOutPublishedAtAsOf++}
 }
 out.Status="TEMPORAL_FOLD_ONLY_NOT_MODEL_VALIDATION"
 return out
}
