package research100
import("testing";"time")
func TestRound019GateNeverAuthorizesTrade(t *testing.T){
 now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
 empty:=CheckCaptureReadiness(Dataset{},now)
 if len(empty.Blockers)<4||empty.TradingAuthorized{t.Fatalf("no-data gate unsafe %+v",empty)}
 a:=qualityFixture(now.Add(-2*time.Hour),"a")
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 r:=CheckCaptureReadiness(d,now)
 if r.Observations!=1||r.DistinctDays!=1||r.TradingAuthorized||r.Status!="INSUFFICIENT_CAPTURE_EVIDENCE"{t.Fatalf("sample gate unsafe %+v",r)}
}
