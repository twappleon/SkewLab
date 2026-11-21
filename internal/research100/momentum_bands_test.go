package research100
import("testing";"time")
func TestRound056StrengthBandsAreDistinct(t *testing.T){
 now:=time.Date(2026,11,20,15,0,0,0,time.UTC)
 d:=Dataset{Items:[]Observation{}}
 for i,m:=range []float64{.01,.03,.06}{
  at:=now.Add(-time.Duration(i+2)*time.Hour)
  o:=qualityFixture(at,string(rune('a'+i)));o.ExpiresAt=at.Add(time.Minute);o.Winner="UP";o.ResolvedAt=o.ExpiresAt.Add(time.Minute);o.Momentum=m
  d.Items=append(d.Items,o)
 }
 b:=AssessMomentumBands(d,now)
 if len(b)!=3||b[0].Published!=1||b[1].Published!=1||b[2].Published!=1{t.Fatalf("bands %+v",b)}
}
