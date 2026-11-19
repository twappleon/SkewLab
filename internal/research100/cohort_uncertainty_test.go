package research100
import("testing";"time")
func TestRound045WilsonByCohortNeedsThirty(t *testing.T){
 now:=time.Date(2026,11,19,15,0,0,0,time.UTC)
 d:=Dataset{Items:[]Observation{}}
 for i:=0;i<30;i++{
  at:=now.Add(-time.Duration(i+40)*time.Minute)
  o:=qualityFixture(at,string(rune('a'+i)));o.ExpiresAt=at.Add(time.Minute);o.Winner="UP";o.ResolvedAt=o.ExpiresAt.Add(time.Minute)
  d.Items=append(d.Items,o)
 }
 r:=AssessCohortUncertainty(d,now)
 if r[0].Samples!=30||r[0].WilsonLower95==nil||*r[0].WilsonLower95<=0||*r[0].WilsonUpper95!=1||r[1].WilsonLower95!=nil{t.Fatalf("interval %+v",r)}
}
