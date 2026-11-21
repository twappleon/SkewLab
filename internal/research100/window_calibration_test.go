package research100
import("testing";"time")
func TestRound041WindowCalibrationUsesPublishedOnly(t *testing.T){
 now:=time.Date(2026,11,19,15,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-time.Hour),"b");b.Market="15m";b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.Winner="DOWN";b.ResolvedAt=now.Add(time.Hour)
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 x:=AssessWindowCalibration(d,now)
 if x[0].Samples!=1||x[0].CorrectCalls!=1||x[0].Brier==nil||x[1].Samples!=0||x[1].Brier!=nil{t.Fatalf("calibration leak %+v",x)}
}
