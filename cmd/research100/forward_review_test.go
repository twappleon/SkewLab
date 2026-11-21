package main
import("encoding/json";"testing";"time";"github.com/twappleon/SkewLab/internal/research100")
func TestRound060CLIProspectiveResearchNoTrading(t *testing.T){
 now:=time.Date(2026,11,20,16,0,0,0,time.UTC)
 o:=research100.Observation{ID:"forward",Market:"5m",Side:"UP",CapturedAt:now.Add(-time.Hour),ExpiresAt:now.Add(-55*time.Minute),QuoteAt:now.Add(-time.Hour-time.Second),Bid:.49,Ask:.51,BidDepth:50,AskDepth:50,Momentum:.03,Skew:.3}
 b,e:=json.Marshal([]research100.Observation{o});if e!=nil{t.Fatal(e)}
 report,e:=Execute("audit",b,now);if e!=nil{t.Fatal(e)}
 if report.ForwardResearch==nil||report.ForwardResearch.TradingAuthorized||report.TradingAuthorized||report.ForwardResearch.Gate.ScoredResolved!=0{t.Fatalf("unverified promise %+v",report)}
}
