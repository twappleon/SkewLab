package research100
import("testing";"time")
func TestRound049CalibrationMustNotAuthorizeTrading(t *testing.T){
 now:=time.Date(2026,11,19,15,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 out:=CheckCalibrationGate(d,now)
 if out.Published!=1||out.UTCObservationDays!=1||out.TradingAuthorized||out.Status!="INSUFFICIENT_CALIBRATION_EVIDENCE"{t.Fatalf("optimistic gate %+v",out)}
}
