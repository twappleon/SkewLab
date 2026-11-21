package research100

import (
 "math"
 "sort"
 "time"
)

// QuoteAgeAudit measures OBSERVED quote age, not end-to-end exchange latency.
// It deliberately excludes candidate snapshots captured after asOf.
type QuoteAgeAudit struct {
 Samples int `json:"samples"`
 OverFiveSeconds int `json:"over_five_seconds"`
 P95Seconds *float64 `json:"p95_seconds,omitempty"`
 Status string `json:"status"`
 TradingAuthorized bool `json:"trading_authorized"`
}
func AssessQuoteAge(d Dataset, asOf time.Time) QuoteAgeAudit {
 out:=QuoteAgeAudit{Status:"NO_OBSERVED_QUOTES",TradingAuthorized:false}
 ages:=make([]float64,0,len(d.Items))
 for _,o:=range d.Items {
  if o.CapturedAt.After(asOf) || o.QuoteAt.IsZero(){continue}
  age:=o.CapturedAt.Sub(o.QuoteAt).Seconds()
  if age<0 || !finite(age){continue}
  out.Samples++
  if age>5{out.OverFiveSeconds++}
  ages=append(ages,age)
 }
 if len(ages)>0 {
  sort.Float64s(ages)
  v:=math.Round(ages[int(math.Ceil(.95*float64(len(ages))))-1]*1000)/1000
  out.P95Seconds=&v
  out.Status="OBSERVED_QUOTE_AGE_ONLY"
 }
 return out
}
