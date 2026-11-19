package research100
import("time")
type BidAskScoreRange struct{
 Samples int `json:"samples"`
 BidBrierMean *float64 `json:"bid_brier_mean,omitempty"`
 AskBrierMean *float64 `json:"ask_brier_mean,omitempty"`
 Status string `json:"status"`
}
func AssessBidAskScoreRange(d Dataset,asOf time.Time)BidAskScoreRange{
 out:=BidAskScoreRange{Status:"NO_PUBLISHED_QUOTES"};a,b:=0.0,0.0
 for _,o:=range d.Items{
  if o.Winner==""||o.ResolvedAt.After(asOf)||o.CapturedAt.After(asOf){continue}
  y:=0.0;if o.Winner==o.Side{y=1}
  a+=(o.Bid-y)*(o.Bid-y);b+=(o.Ask-y)*(o.Ask-y);out.Samples++
 }
 if out.Samples>0{v:=a/float64(out.Samples);w:=b/float64(out.Samples);out.BidBrierMean=&v;out.AskBrierMean=&w;out.Status="QUOTE_INTERVAL_SENSITIVITY_ONLY"}
 return out
}
