package research100
import("math";"time")
type WindowCalibration struct{
 Market string `json:"market"`
 Samples int `json:"samples"`
 CorrectCalls int `json:"correct_calls"`
 Brier *float64 `json:"brier,omitempty"`
 LogLoss *float64 `json:"log_loss,omitempty"`
 Status string `json:"status"`
}
func AssessWindowCalibration(d Dataset,asOf time.Time)[]WindowCalibration{
 out:=[]WindowCalibration{{Market:"5m",Status:"NO_PUBLISHED_SAMPLES"},{Market:"15m",Status:"NO_PUBLISHED_SAMPLES"}}
 sq:=[2]float64{};ll:=[2]float64{}
 for _,o:=range d.Items{
  if o.Winner==""||o.ResolvedAt.After(asOf)||o.CapturedAt.After(asOf){continue}
  i:=0;if o.Market=="15m"{i=1}
  p:=(o.Bid+o.Ask)/2
  if p<=0||p>=1||!finite(p){continue}
  y:=0.0;if o.Winner==o.Side{y=1;out[i].CorrectCalls++}
  sq[i]+=(p-y)*(p-y)
  ll[i]-=y*math.Log(p)+(1-y)*math.Log(1-p)
  out[i].Samples++
 }
 for i:=range out{
  n:=out[i].Samples;if n==0{continue}
  b:=sq[i]/float64(n);l:=ll[i]/float64(n)
  out[i].Brier=&b;out[i].LogLoss=&l;out[i].Status="IN_SAMPLE_QUOTE_PROXY_ONLY"
 }
 return out
}
