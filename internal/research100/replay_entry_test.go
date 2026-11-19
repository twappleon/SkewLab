package research100
import("testing";"time")
func TestRound071ValidEntryIsNotRealExecution(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Minute),"a")
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 x:=AssessReplayEntryCandidates(d,now)
 if x.Observations!=1||x.PricedEntryCandidates!=1||x.TradingAuthorized{t.Fatalf("claimed trade %+v",x)}
}
