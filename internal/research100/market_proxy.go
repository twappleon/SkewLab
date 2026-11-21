package research100
import("time";"github.com/twappleon/SkewLab/internal/experiment")
type MarketProxy struct{
 UpFraction float64 `json:"up_fraction"`
 DownFraction float64 `json:"down_fraction"`
 Valid bool `json:"valid"`
 Status string `json:"status"`
}
func AssessMarketProxy(s experiment.Sample,asOf time.Time)MarketProxy{
 out:=MarketProxy{Status:"INVALID_PAIRED_QUOTES"}
 if !ValidatePairedSnapshot(s,asOf).Valid{return out}
 up:=(s.UpBid+s.UpAsk)/2;down:=(s.DownBid+s.DownAsk)/2;total:=up+down
 if total<=0||!finite(total){return out}
 out.UpFraction=up/total;out.DownFraction=down/total;out.Valid=true
 out.Status="NORMALIZED_MARKET_MIDPOINT_NOT_CALIBRATED_TRUTH"
 return out
}
