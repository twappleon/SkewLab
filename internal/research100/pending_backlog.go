package research100
import("time")
type BacklogBucket struct{
 Market string `json:"market"`
 PendingExpired int `json:"pending_expired"`
 Over30Minutes int `json:"over_30_minutes"`
 Over2Hours int `json:"over_2_hours"`
 Over24Hours int `json:"over_24_hours"`
 Status string `json:"status"`
}
func AssessPendingBacklog(d Dataset,asOf time.Time)[]BacklogBucket{
 out:=[]BacklogBucket{{Market:"5m",Status:"NO_EXPIRED_PENDING_LABELS"},{Market:"15m",Status:"NO_EXPIRED_PENDING_LABELS"}}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf)||o.ExpiresAt.After(asOf)||(o.Winner!=""&&!o.ResolvedAt.After(asOf)){continue}
  i:=0;if o.Market=="15m"{i=1}
  lag:=asOf.Sub(o.ExpiresAt)
  x:=&out[i];x.PendingExpired++
  if lag>30*time.Minute{x.Over30Minutes++}
  if lag>2*time.Hour{x.Over2Hours++}
  if lag>24*time.Hour{x.Over24Hours++}
 }
 for i:=range out{if out[i].PendingExpired>0{out[i].Status="OUTCOME_PUBLICATION_OR_COLLECTION_DELAY_UNVERIFIED"}}
 return out
}
