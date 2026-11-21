package main
import("encoding/json";"testing";"time";"github.com/twappleon/SkewLab/internal/research100")
func TestRound030CLISettlementEvidenceVisibleAndReadOnly(t *testing.T){
 now:=time.Date(2026,11,20,11,0,0,0,time.UTC)
 rec:=research100.Observation{ID:"m30",Market:"5m",Side:"UP",CapturedAt:now.Add(-time.Hour),ExpiresAt:now.Add(-55*time.Minute),QuoteAt:now.Add(-time.Hour-time.Second),Bid:.48,Ask:.50,BidDepth:20,AskDepth:20,Momentum:.03,Skew:.3}
 b,e:=json.Marshal([]research100.Observation{rec});if e!=nil{t.Fatal(e)}
 report,e:=Execute("audit",b,now);if e!=nil{t.Fatal(e)}
 if report.TradingAuthorized||report.SettlementQuality==nil||report.SettlementQuality.TradingAuthorized||report.SettlementQuality.PendingAge.Over30Minutes!=1{t.Fatalf("settlement gate bypassed %+v",report)}
}
