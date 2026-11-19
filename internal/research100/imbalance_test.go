package research100
import("testing";"time")
func TestRound033DepthRatioIsNotForecast(t *testing.T){
 now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Minute),"a");a.BidDepth=90;a.AskDepth=10
 b:=qualityFixture(now.Add(-time.Minute),"b");b.BidDepth=10;b.AskDepth=90
 c:=qualityFixture(now.Add(-time.Minute),"c")
 d,e:=BuildDataset([]Observation{a,b,c});if e!=nil{t.Fatal(e)}
 got:=AssessTopBookImbalance(d,now)
 if got.BidDominant!=1||got.AskDominant!=1||got.Neutral!=1||got.Observations!=3{t.Fatalf("bad ratio %+v",got)}
}
