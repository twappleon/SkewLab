package research100
import("testing";"time")
func TestRound051TrainingCannotUseLatePublishedLabel(t *testing.T){
 now:=time.Date(2026,11,20,16,0,0,0,time.UTC);cutoff:=now.Add(-time.Hour)
 a:=qualityFixture(now.Add(-3*time.Hour),"train");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-2*time.Hour),"late");b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.Winner="UP";b.ResolvedAt=cutoff.Add(time.Minute)
 c:=qualityFixture(now.Add(-30*time.Minute),"hold");c.ExpiresAt=c.CapturedAt.Add(time.Minute);c.Winner="DOWN";c.ResolvedAt=c.ExpiresAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a,b,c});if e!=nil{t.Fatal(e)}
 x:=AssessChronologicalFold(d,cutoff,now)
 if x.TrainPublishedBeforeCutoff!=1||x.UnavailableTrainLabels!=1||x.HeldOutPublishedAtAsOf!=1{t.Fatalf("leaky fold %+v",x)}
}
