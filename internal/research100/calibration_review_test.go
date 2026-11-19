package research100
import("testing";"time")
func TestRound050CalibrationNeverBecomesTradingPermission(t *testing.T){
 now:=time.Date(2026,11,19,15,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 out:=ReviewCalibrationEvidence(d,now)
 if out.TradingAuthorized||out.Gate.TradingAuthorized||out.Status!="INSUFFICIENT_CALIBRATION_EVIDENCE"||len(out.WindowScores)!=2||len(out.Reliability)!=5{t.Fatalf("bad review %+v",out)}
}
