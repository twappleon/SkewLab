package research100
import("testing";"time")
func TestRound057AlignmentIsNotCertainWin(t *testing.T){
 now:=time.Date(2026,11,20,15,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="DOWN";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-time.Hour),"b");b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.Winner="UP";b.ResolvedAt=b.ExpiresAt.Add(time.Minute);b.Skew=-.4
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 r:=AssessSignalCoherence(d,now)
 if r.Published!=2||r.MomentumAligned!=2||r.BothAligned!=1||r.BothAlignedWins!=0{t.Fatalf("signal alignment leaked wins %+v",r)}
}
