package research100
import("testing";"time")
func TestRound036SameBookRoundTripCanBreakStop(t *testing.T){
 now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Minute),"a");a.Bid=.49;a.Ask=.51
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 out:=AssessRoundTripFriction(d,now)
 if out.Observations!=1||out.CostOnlyLossAboveFivePct!=1||out.MeanModeledLossPct==nil||*out.MeanModeledLossPct<=5{t.Fatalf("no friction warning %+v",out)}
}
