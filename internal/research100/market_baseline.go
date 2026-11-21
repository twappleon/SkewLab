package research100
import("math";"time";"github.com/twappleon/SkewLab/internal/experiment")
type BaselineSide struct{
 Side string `json:"side"`
 Status string `json:"status"`
}
func MarketBaselineSide(s experiment.Sample,asOf time.Time)BaselineSide{
 out:=BaselineSide{Side:"WAIT",Status:"NO_VALID_SIMULTANEOUS_BOOK"}
 p:=AssessMarketProxy(s,asOf)
 if !p.Valid{return out}
 if math.Abs(p.UpFraction-p.DownFraction)<1e-9{out.Status="TIED_MARKET_PROXY_ABSTAIN";return out}
 if p.UpFraction>p.DownFraction{out.Side="UP"}else{out.Side="DOWN"}
 out.Status="MARKET_PRICE_PROXY_BASELINE_ONLY"
 return out
}
