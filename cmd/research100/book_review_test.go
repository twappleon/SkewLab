package main
import("encoding/json";"testing";"time";"github.com/twappleon/SkewLab/internal/research100")
func TestRound040CLIIncludesBookEvidenceAndPauses(t *testing.T){
 now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
 o:=research100.Observation{ID:"book",Market:"5m",Side:"UP",CapturedAt:now.Add(-time.Minute),ExpiresAt:now.Add(time.Minute),QuoteAt:now.Add(-time.Minute-time.Second),Bid:.49,Ask:.51,BidDepth:50,AskDepth:50,Momentum:.03,Skew:.3}
 b,e:=json.Marshal([]research100.Observation{o});if e!=nil{t.Fatal(e)}
 x,e:=Execute("audit",b,now);if e!=nil{t.Fatal(e)}
 if x.BookQuality==nil||x.BookQuality.TradingAuthorized||x.TradingAuthorized||x.BookQuality.RoundTrip.CostOnlyLossAboveFivePct!=1{t.Fatalf("unsafe audit report %+v",x)}
}
