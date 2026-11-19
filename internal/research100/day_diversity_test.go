package research100
import("testing";"time")
func TestRound014DayConcentrationWarning(t *testing.T){
 now:=time.Date(2026,11,20,0,0,0,0,time.UTC)
 d:=Dataset{Items:[]Observation{}}
 for i:=0;i<8;i++{
  o:=qualityFixture(now.Add(-time.Duration(i+10)*time.Minute),string(rune('a'+i)))
  d.Items=append(d.Items,o)
 }
 r:=DayConcentration(d,now)
 if r[0].UTCDateCount!=1||!r[0].Concentrated||r[0].MaximumSingleDayCount!=8{t.Fatalf("missing warning %+v",r[0])}
 if r[1].Concentrated||r[1].UTCDateCount!=0{t.Fatalf("invented day %+v",r[1])}
}
