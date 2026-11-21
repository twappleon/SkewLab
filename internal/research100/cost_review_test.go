package research100
import("testing";"time")
func TestRound070NoCostModelCanAuthorizeTrades(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.Bid=.49;a.Ask=.51
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 x:=ReviewCostResearch(d,now)
 if x.TradingAuthorized||x.RiskEnvelope.TradingAuthorized||x.RiskEnvelope.WithinDailyTwoDollarLoss||x.StopHeadroom.AlreadyExceedsStop!=1||len(x.RoundTrips.Cases)!=3{t.Fatalf("unsafe cost review %+v",x)}
}
