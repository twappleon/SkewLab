package research100
import("math";"sort";"time")
type LabelQuantiles struct{
 Resolved int `json:"resolved"`
 MedianSeconds *float64 `json:"median_seconds,omitempty"`
 P95Seconds *float64 `json:"p95_seconds,omitempty"`
 Status string `json:"status"`
}
func AssessLabelQuantiles(d Dataset,asOf time.Time)LabelQuantiles{
 out:=LabelQuantiles{Status:"NO_PUBLISHED_LABELS"}
 ages:=[]float64{}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf)||o.Winner==""||o.ResolvedAt.After(asOf){continue}
  lag:=o.ResolvedAt.Sub(o.ExpiresAt).Seconds()
  if lag<0||!finite(lag){continue}
  ages=append(ages,lag)
 }
 sort.Float64s(ages);out.Resolved=len(ages)
 if len(ages)==0{return out}
 med:=ages[(len(ages)-1)/2];p95:=ages[int(math.Ceil(float64(len(ages))*.95))-1]
 out.MedianSeconds=&med;out.P95Seconds=&p95
 out.Status="ARCHIVE_PUBLICATION_TIMES_ONLY_NOT_ORACLE_PROOF"
 return out
}
