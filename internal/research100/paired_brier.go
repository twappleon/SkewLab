package research100
import("time")
type PairedBrier struct{
 Samples int `json:"samples"`
 MeanQuoteMinusCoinFlip *float64 `json:"mean_quote_minus_coin_flip,omitempty"`
 Status string `json:"status"`
}
func AssessPairedBrier(d Dataset,asOf time.Time)PairedBrier{
 out:=PairedBrier{Status:"NO_PUBLISHED_PAIRED_SAMPLES"}
 total:=0.0
 for _,o:=range d.Items{
  if o.Winner==""||o.CapturedAt.After(asOf)||o.ResolvedAt.After(asOf){continue}
  p:=(o.Bid+o.Ask)/2;y:=0.0
  if o.Winner==o.Side{y=1}
  total+=(p-y)*(p-y)-.25
  out.Samples++
 }
 if out.Samples>0{v:=total/float64(out.Samples);out.MeanQuoteMinusCoinFlip=&v;out.Status="IN_SAMPLE_COMPARISON_NOT_PROSPECTIVE_TRADING_EDGE"}
 return out
}
