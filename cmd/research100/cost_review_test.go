package main
import("encoding/json";"testing";"time";"github.com/twappleon/SkewLab/internal/research100")
func TestRound070CLICostResearchReadOnly(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 o:=research100.Observation{ID:"risk",Market:"5m",Side:"UP",CapturedAt:now.Add(-time.Hour),ExpiresAt:now.Add(-55*time.Minute),QuoteAt:now.Add(-time.Hour-time.Second),Bid:.49,Ask:.51,BidDepth:50,AskDepth:50,Momentum:.03,Skew:.3}
 b,e:=json.Marshal([]research100.Observation{o});if e!=nil{t.Fatal(e)}
 x,e:=Execute("audit",b,now);if e!=nil{t.Fatal(e)}
 if x.CostResearch==nil||x.CostResearch.TradingAuthorized||x.CostResearch.RiskEnvelope.WithinDailyTwoDollarLoss||x.TradingAuthorized{t.Fatalf("risk review authority leak %+v",x)}
}
