package main
import ("encoding/json";"testing";"time";"github.com/twappleon/SkewLab/internal/research100")
func TestRound020CLICaptureQualityIsReadOnly(t *testing.T){
 now:=time.Date(2026,11,19,14,0,0,0,time.UTC)
 sample:=research100.Observation{ID:"fixed-market",Market:"5m",Side:"UP",CapturedAt:now.Add(-5*time.Minute),ExpiresAt:now.Add(-time.Minute),QuoteAt:now.Add(-5*time.Minute-time.Second),Bid:.4,Ask:.42,BidDepth:20,AskDepth:20,Momentum:.03,Skew:.2}
 b,e:=json.Marshal([]research100.Observation{sample});if e!=nil{t.Fatal(e)}
 got,e:=Execute("audit",b,now);if e!=nil{t.Fatal(e)}
 if got.CaptureQuality==nil||got.CaptureQuality.Observations!=1||got.CaptureQuality.TradingAuthorized||got.QuoteAge==nil||got.QuoteAge.Samples!=1||got.TradingAuthorized{t.Fatalf("unsafe quality output %+v",got)}
 if len(got.LabelLag)!=2||len(got.DayDiversity)!=2{t.Fatalf("missing cohort diagnostics %+v",got)}
}
