package research100
import("testing";"time")
func TestRound013GapIsNotUptime(t *testing.T){
 now:=time.Date(2026,11,19,8,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-20*time.Minute),"a")
 b:=qualityFixture(now.Add(-2*time.Minute),"b")
 d,e:=BuildDataset([]Observation{b,a});if e!=nil{t.Fatal(e)}
 x:=GapAudit(d,now)
 if x[0].ConsecutivePairs!=1||x[0].GapsOverTwoWindows!=1||x[0].MaximumGapSeconds==nil||*x[0].MaximumGapSeconds!=1080{t.Fatalf("wrong gaps %+v",x)}
 if x[1].ConsecutivePairs!=0||x[1].MaximumGapSeconds!=nil{t.Fatal("invented market windows")}
}
