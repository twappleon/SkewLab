package polymarketv2
import("testing";"time")
func TestCalibrationFailClosed(t *testing.T){
 now:=time.Date(2026,11,18,0,0,0,0,time.UTC)
 samples:=make([]LabeledOutcome,130)
 for i:=range samples{
  start:=now.Add(-time.Duration(150-i)*48*time.Hour)
  samples[i]=LabeledOutcome{Cohort:"binary-weather-independent",OccurredAt:start,ResolvedAt:start.Add(time.Hour),Outcome:i%5!=0,Source:"official resolution"}
 }
 c,err:=Calibrate("binary-weather-independent",now,samples)
 if err!=nil{t.Fatal(err)}
 if c.TrainingCount!=100||c.ForwardCount!=30||!c.Forecast.Valid(now){t.Fatalf("bad forecast %+v",c)}
 if c.ForwardBrier>=c.ForwardBaselineBrier{t.Fatal("forward model must beat baseline for fixture")}
 c,err=Calibrate("binary-weather-independent",now,samples[:120])
 if err!=nil||c.Status!="INSUFFICIENT_EVIDENCE"{t.Fatalf("incorrect sample gate %+v %v",c,err)}
 bad:=append([]LabeledOutcome{},samples...)
 bad[110].OccurredAt=bad[0].OccurredAt
 if _,err=Calibrate("binary-weather-independent",now,bad);err==nil{t.Fatal("leakage chronology accepted")}
}
