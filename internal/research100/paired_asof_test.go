package research100
import("testing";"time";"github.com/twappleon/SkewLab/internal/experiment")
func TestRound088FutureOutcomeNotScored(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=pairedFixture(now.Add(-time.Hour),"a");a.Winner="UP";a.ResolvedAt=now.Add(time.Hour)
 b:=pairedFixture(now.Add(-2*time.Hour),"b");b.Winner="DOWN";b.ResolvedAt=b.MarketEnd.Add(time.Minute)
 c:=pairedFixture(now.Add(time.Minute),"c")
 r:=AuditPairedAsOf([]experiment.Sample{a,b,c},now)
 if r.FutureLabels!=1||r.PublishedOutcomeRows!=1||r.FutureCaptures!=1{t.Fatalf("leaked paired future %+v",r)}
}
