package research100
import("testing";"time")
func TestRound039LayeredBookGateNeverOrders(t *testing.T){
 now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Minute),"old");a.QuoteAt=a.CapturedAt.Add(-8*time.Second)
 b:=qualityFixture(now.Add(-time.Minute),"wide");b.Bid=.4;b.Ask=.5
 c:=qualityFixture(now.Add(-time.Minute),"shallow");c.BidDepth=5
 d:=qualityFixture(now.Add(-time.Minute),"good")
 records,e:=BuildDataset([]Observation{a,b,c,d});if e!=nil{t.Fatal(e)}
 x:=AssessBookGate(records,now)
 if x.Observations!=4||x.RejectedOldQuote!=1||x.RejectedWideSpread!=1||x.RejectedShallowBook!=1||x.PassedVisibleFilters!=1||x.TradingAuthorized{t.Fatalf("unexpected book permission %+v",x)}
}
