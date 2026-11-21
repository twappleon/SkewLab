package research100
import("testing";"time")
func TestRound028PreviouslyPublishedLabelCannotDisappearSilently(t *testing.T){
 now:=time.Date(2026,11,20,10,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-time.Hour),"b");b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.Winner="DOWN";b.ResolvedAt=b.ExpiresAt.Add(time.Minute)
 before,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 changed:=a;changed.Winner="DOWN"
 after,e:=BuildDataset([]Observation{changed});if e!=nil{t.Fatal(e)}
 got:=AuditLabelRevisions(before,after,now)
 if got.ComparablePreviouslyPublished!=2||got.RemovedMarketIDs!=1||got.ChangedWinners!=1{t.Fatalf("missing revisions %+v",got)}
}
