package research100
import("testing";"time")
func TestRound059ForwardGateNotSatisfiedByTinyArchive(t *testing.T){
 now:=time.Date(2026,11,20,15,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 x:=CheckForwardGate(d,now)
 if x.ScoredResolved!=0||x.ObservedPublishedDays!=1||x.TradingAuthorized||x.Status!="INSUFFICIENT_FORWARD_SAMPLES"{t.Fatalf("premature release %+v",x)}
}
