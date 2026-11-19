package research100
import("time")
type LabelLag struct{
 Market string `json:"market"`
 Resolved int `json:"resolved"`
 Pending int `json:"pending"`
 Overdue30Minutes int `json:"overdue_30_minutes"`
 MaximumPublishedLagSeconds *float64 `json:"max_published_lag_seconds,omitempty"`
 Status string `json:"status"`
}
// Publication lag uses the archived label timestamp, not an oracle assertion.
// Unresolved markets and labels published after asOf never count as wins.
func AssessLabelLag(d Dataset,asOf time.Time)[]LabelLag{
 out:=[]LabelLag{{Market:"5m",Status:"NO_PUBLISHED_LABELS"},{Market:"15m",Status:"NO_PUBLISHED_LABELS"}}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  i:=0;if o.Market=="15m"{i=1}
  x:=&out[i]
  if o.Winner==""||o.ResolvedAt.After(asOf){
   x.Pending++
   if asOf.After(o.ExpiresAt.Add(30*time.Minute)){x.Overdue30Minutes++}
   continue
  }
  x.Resolved++
  lag:=o.ResolvedAt.Sub(o.ExpiresAt).Seconds()
  if x.MaximumPublishedLagSeconds==nil||lag>*x.MaximumPublishedLagSeconds{v:=lag;x.MaximumPublishedLagSeconds=&v}
 }
 for i:=range out{
  switch{case out[i].Overdue30Minutes>0:out[i].Status="LABEL_BACKLOG_REVIEW"
  case out[i].Resolved>0:out[i].Status="PUBLISHED_LABELS_NOT_INDEPENDENT_ORACLE_PROOF"}
 }
 return out
}
