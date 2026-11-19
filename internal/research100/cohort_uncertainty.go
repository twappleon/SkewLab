package research100
import("math";"time")
type CohortUncertainty struct{
 Market string `json:"market"`
 Side string `json:"side"`
 Samples int `json:"samples"`
 Wins int `json:"wins"`
 WilsonLower95 *float64 `json:"wilson_lower_95,omitempty"`
 WilsonUpper95 *float64 `json:"wilson_upper_95,omitempty"`
 Status string `json:"status"`
}
func AssessCohortUncertainty(d Dataset,asOf time.Time)[]CohortUncertainty{
 out:=[]CohortUncertainty{}
 for _,market:=range []string{"5m","15m"}{for _,side:=range []string{"UP","DOWN"}{out=append(out,CohortUncertainty{Market:market,Side:side,Status:"FEWER_THAN_30_PUBLISHED_SAMPLES"})}}
 for _,o:=range d.Items{
  if o.Winner==""||o.ResolvedAt.After(asOf)||o.CapturedAt.After(asOf){continue}
  i:=0;if o.Market=="15m"{i=2};if o.Side=="DOWN"{i++}
  out[i].Samples++;if o.Winner==o.Side{out[i].Wins++}
 }
 for i:=range out{
  q:=&out[i];if q.Samples<30{continue}
  n:=float64(q.Samples);p:=float64(q.Wins)/n;z:=1.96
  den:=1+z*z/n;center:=p+z*z/(2*n)
  half:=z*math.Sqrt(p*(1-p)/n+z*z/(4*n*n))
  low:=math.Max(0,(center-half)/den);high:=math.Min(1,(center+half)/den)
  q.WilsonLower95=&low;q.WilsonUpper95=&high
  q.Status="RESEARCH_INTERVAL_NOT_INDEPENDENCE_PROOF"
 }
 return out
}
