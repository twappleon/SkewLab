package research100
import("testing";"time")
func TestRound048ChronologicalCalibrationSplit(t *testing.T){
 now:=time.Date(2026,11,20,15,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-4*time.Hour),"earlier");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-2*time.Hour),"later");b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.Winner="DOWN";b.ResolvedAt=b.ExpiresAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 x:=AssessCalibrationDrift(d,now.Add(-3*time.Hour),now)
 if x.Earlier!=1||x.Later!=1||x.EarlierBrier==nil||x.LaterBrier==nil||x.Status!="INSUFFICIENT_PREDECLARED_TIME_SPLIT"{t.Fatalf("leaky split %+v",x)}
}
