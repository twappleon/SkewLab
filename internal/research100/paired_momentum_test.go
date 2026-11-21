package research100
import("testing";"time")
func TestRound084MomentumFilterCanAbstain(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=pairedFixture(now,"a");a.MomentumPct=.01
 if x:=MomentumSide(a,now);x.Side!="WAIT"{t.Fatal("weak momentum forced")}
 a.MomentumPct=-.03
 if x:=MomentumSide(a,now);x.Side!="DOWN"{t.Fatal("negative momentum mismatch")}
}
