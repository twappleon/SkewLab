package research100
import("testing";"time")
func TestRound026ContradictoryArchiveDetected(t *testing.T){
 now:=time.Date(2026,11,20,10,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"same");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=a;b.Winner="DOWN"
 c:=a;c.ResolvedAt=now.Add(time.Hour)
 out:=AssessRawOutcomeConflicts([]Observation{a,b,c},now)
 if out.SeenLabeledIDs!=1||out.DuplicatePublishedLabels!=1||out.ConflictingWinners!=1{t.Fatalf("conflict not detected %+v",out)}
}
