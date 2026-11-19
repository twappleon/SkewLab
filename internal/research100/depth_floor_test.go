package research100
import("testing";"time")
func TestRound031SeparateVisibleDepthSides(t *testing.T){
 now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Minute),"a");a.BidDepth=15;a.AskDepth=35
 b:=qualityFixture(now.Add(-time.Minute),"b");b.BidDepth=35;b.AskDepth=10
 c:=qualityFixture(now.Add(-time.Minute),"c");c.BidDepth=35;c.AskDepth=35
 d,e:=BuildDataset([]Observation{a,b,c});if e!=nil{t.Fatal(e)}
 out:=AssessDepthFloors(d,now,20,20)
 if out[0].Observations!=3||out[0].BelowBidFloor!=1||out[0].BelowAskFloor!=1||out[0].MeetsBoth!=1{t.Fatalf("wrong depth %+v",out[0])}
 if x:=AssessDepthFloors(d,now,-1,20);x[0].Status!="INVALID_DEPTH_THRESHOLD"||x[0].Observations!=0{t.Fatal("invalid threshold allowed")}
}
