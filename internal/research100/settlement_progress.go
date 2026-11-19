package research100
import("time")
type SettlementProgress struct{
 Market string `json:"market"`
 MaturedMarkets int `json:"matured_markets"`
 Published int `json:"published"`
 Pending int `json:"pending"`
 PublishedFraction *float64 `json:"published_fraction,omitempty"`
 Status string `json:"status"`
}
func AssessSettlementProgress(d Dataset,asOf time.Time)[]SettlementProgress{
 out:=[]SettlementProgress{{Market:"5m",Status:"NO_MATURED_MARKET_SAMPLES"},{Market:"15m",Status:"NO_MATURED_MARKET_SAMPLES"}}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf)||o.ExpiresAt.After(asOf){continue}
  i:=0;if o.Market=="15m"{i=1}
  x:=&out[i];x.MaturedMarkets++
  if o.Winner!=""&&!o.ResolvedAt.After(asOf){x.Published++}else{x.Pending++}
 }
 for i:=range out{
  x:=&out[i];if x.MaturedMarkets==0{continue}
  v:=float64(x.Published)/float64(x.MaturedMarkets)
  x.PublishedFraction=&v;x.Status="OBSERVED_MATURED_MARKETS_ONLY_NOT_EXCHANGE_UPTIME"
 }
 return out
}
