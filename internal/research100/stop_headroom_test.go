package research100
import("testing";"time")
func TestRound066StopThresholdCannotEraseRoundTripFees(t *testing.T){
 now:=time.Date(2026,11,19,0,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Minute),"a");a.Bid=.49;a.Ask=.51
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 out:=AssessStopHeadroom(d,now,5)
 if out.Observations!=1||out.AlreadyExceedsStop!=1||out.MinHeadroomPct==nil||*out.MinHeadroomPct>=0{t.Fatalf("stop gap %+v",out)}
}
