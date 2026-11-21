package research100
import("math";"sort";"time")
type PendingAgeAudit struct{
 PendingExpired int `json:"pending_expired"`
 Over30Minutes int `json:"over_30_minutes"`
 OldestSeconds *float64 `json:"oldest_seconds,omitempty"`
 P95Seconds *float64 `json:"p95_seconds,omitempty"`
 Status string `json:"status"`
}
func AssessPendingAge(d Dataset,asOf time.Time)PendingAgeAudit{
 out:=PendingAgeAudit{Status:"NO_EXPIRED_PENDING_LABELS"}
 ages:=[]float64{}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf)||o.ExpiresAt.After(asOf)||(o.Winner!=""&&!o.ResolvedAt.After(asOf)){continue}
  age:=asOf.Sub(o.ExpiresAt).Seconds()
  if age<0||!finite(age){continue}
  ages=append(ages,age)
  if age>1800{out.Over30Minutes++}
 }
 sort.Float64s(ages);out.PendingExpired=len(ages)
 if len(ages)==0{return out}
 x:=ages[len(ages)-1];p:=ages[int(math.Ceil(float64(len(ages))*.95))-1]
 out.OldestSeconds=&x;out.P95Seconds=&p
 out.Status="ARCHIVED_PENDING_AGE_NOT_A_PROVIDER_OUTAGE_CLAIM"
 return out
}
