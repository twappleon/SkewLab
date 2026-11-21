package research100
import("testing";"time")
func TestRound012HorizonSeparatesWindows(t *testing.T){
 now:=time.Date(2026,11,19,5,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Minute),"5")
 b:=qualityFixture(now.Add(-time.Minute),"15");b.Market="15m";b.ExpiresAt=b.CapturedAt.Add(12*time.Minute)
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 got:=AssessCaptureHorizon(d,now)
 if len(got)!=2||got[0].Count!=1||got[1].Count!=1||*got[0].MinSeconds!=180||*got[1].MaxSeconds!=720{t.Fatalf("invalid horizon %+v",got)}
 prev:=AssessCaptureHorizon(d,now.Add(-2*time.Minute));if prev[0].MinSeconds!=nil||prev[1].Count!=0{t.Fatalf("future evidence %v",prev)}
}
