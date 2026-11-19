package research100
import("testing";"time")
func TestRound085AlignmentAbstainsWhenSkewContradicts(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 s:=pairedFixture(now,"a")
 s.Skew=-.3
 if x:=MomentumSkewSide(s,now);x.Side!="WAIT"{t.Fatal("skew disagreement ignored")}
 s.Skew=.4
 if x:=MomentumSkewSide(s,now);x.Side!="UP"{t.Fatal("aligned rule did not qualify")}
}
