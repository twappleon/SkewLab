package main
import("encoding/json";"testing";"time";"github.com/twappleon/SkewLab/internal/research100")
func TestRound050CLIProbabilityCalibrationReadOnly(t *testing.T){
 now:=time.Date(2026,11,19,15,0,0,0,time.UTC)
 o:=research100.Observation{ID:"cal",Market:"5m",Side:"UP",CapturedAt:now.Add(-time.Hour),ExpiresAt:now.Add(-55*time.Minute),QuoteAt:now.Add(-time.Hour-time.Second),Bid:.49,Ask:.51,BidDepth:20,AskDepth:20,Momentum:.03,Skew:.3,Winner:"UP",ResolvedAt:now.Add(-54*time.Minute)}
 b,e:=json.Marshal([]research100.Observation{o});if e!=nil{t.Fatal(e)}
 x,e:=Execute("audit",b,now);if e!=nil{t.Fatal(e)}
 if x.ProbabilityReview==nil||x.ProbabilityReview.TradingAuthorized||x.ProbabilityReview.FlatBaseline.PublishedSamples!=1||x.TradingAuthorized{t.Fatalf("in-sample research gate unsafe %+v",x)}
}
