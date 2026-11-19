package research100
import("testing";"time")
func TestRound053OneDayNotTenIndependentDays(t *testing.T){
 now:=time.Date(2026,11,20,16,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-time.Hour),"b");b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.Winner="DOWN";b.ResolvedAt=b.ExpiresAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 x:=AssessHeldOutDays(d,now.Add(-2*time.Hour),now)
 if x.CapturedAfterCutoff!=2||x.ScorableAtAsOf!=2||x.DistinctUTCScoreDays!=1||x.Status!="INSUFFICIENT_DISTINCT_HELDOUT_DAYS"{t.Fatalf("day clustering %+v",x)}
}
