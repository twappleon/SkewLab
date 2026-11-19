package research100
import("testing";"time")
func TestRound018AsOfTimelineNoFutureLeak(t *testing.T){
 now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-3*time.Hour),"known");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-2*time.Hour),"later");b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.Winner="DOWN";b.ResolvedAt=now.Add(time.Hour)
 c:=qualityFixture(now.Add(-time.Minute),"pending")
 future:=qualityFixture(now.Add(time.Minute),"future")
 d,e:=BuildDataset([]Observation{a,b,c,future});if e!=nil{t.Fatal(e)}
 x:=AssessAsOfIntegrity(d,now)
 if x.PublishedLabels!=1||x.FutureOutcomeLabels!=1||x.PendingWithoutOutcome!=1||x.CapturedInFuture!=1||x.TradingAuthorized{t.Fatalf("lookahead audit broken %+v",x)}
}
