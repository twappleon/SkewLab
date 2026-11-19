package research100
import("testing";"time")
func TestRound040BookReviewCannotClaimExecutableEdge(t *testing.T){
 now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Minute),"a");a.Bid=.49;a.Ask=.51
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 got:=ReviewBookEvidence(d,now)
 if got.TradingAuthorized||got.Gate.Observations!=1||got.RoundTrip.CostOnlyLossAboveFivePct!=1||got.Status!="ROUND_TRIP_COST_CAN_EXCEED_5_PERCENT_STOP"{t.Fatalf("book model unsafe %+v",got)}
}
