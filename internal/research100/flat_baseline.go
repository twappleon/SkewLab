package research100
import("time")
type FlatBrierBaseline struct{
 PublishedSamples int `json:"published_samples"`
 MeanBrier *float64 `json:"mean_brier,omitempty"`
 Status string `json:"status"`
}
func AssessFlatBrierBaseline(d Dataset,asOf time.Time)FlatBrierBaseline{
 out:=FlatBrierBaseline{Status:"NO_PUBLISHED_OUTCOMES"}
 for _,o:=range d.Items{
  if o.Winner==""||o.ResolvedAt.After(asOf)||o.CapturedAt.After(asOf){continue}
  out.PublishedSamples++
 }
 if out.PublishedSamples>0{v:=.25;out.MeanBrier=&v;out.Status="PREDECLARED_COIN_FLIP_BASELINE_NOT_MARKET_PRICE"}
 return out
}
