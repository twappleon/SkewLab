package research100
import("testing";"time")
func TestRound038SameBTCWindowNotIndependentSample(t *testing.T){
 now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-10*time.Minute),"a")
 b:=qualityFixture(now.Add(-10*time.Minute+30*time.Second),"b");b.Market="15m"
 c:=qualityFixture(now.Add(-time.Minute),"c");c.Market="15m"
 d,e:=BuildDataset([]Observation{c,a,b});if e!=nil{t.Fatal(e)}
 x:=AssessCrossWindowOverlap(d,now)
 if x.FiveMinuteCandidates!=1||x.FifteenMinuteCandidates!=2||x.FiveMinuteNearFifteenMinute!=1{t.Fatalf("overlap %+v",x)}
}
