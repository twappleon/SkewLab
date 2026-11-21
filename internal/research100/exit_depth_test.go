package research100
import("testing";"time")
func TestRound032TopBidIsNotSellFill(t *testing.T){
 now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Minute),"a");a.BidDepth=3
 b:=qualityFixture(now.Add(-time.Minute),"b");b.BidDepth=10
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 got:=AssessPotentialExitDepth(d,now,5)
 if got.VisibleCandidates!=2||got.InsufficientTopBidShares!=1||got.EnoughTopBidShares!=1||got.TradingAuthorized{t.Fatalf("unsafe exit depth %+v",got)}
}
