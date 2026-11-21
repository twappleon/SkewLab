package research100
import("testing";"time")
func TestRound055WeekdayWeekendSeparated(t *testing.T){
 now:=time.Date(2026,11,23,0,0,0,0,time.UTC)
 a:=qualityFixture(time.Date(2026,11,20,10,0,0,0,time.UTC),"fri");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(time.Date(2026,11,21,10,0,0,0,time.UTC),"sat");b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.Winner="DOWN";b.ResolvedAt=b.ExpiresAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 x:=AssessWeekRegimeScore(d,now)
 if x[0].Samples!=1||x[0].Wins!=1||x[1].Samples!=1||x[1].Wins!=0{t.Fatalf("calendar bucket %+v",x)}
}
