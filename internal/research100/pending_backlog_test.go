package research100
import("testing";"time")
func TestRound022BacklogSeparatesMatured(t *testing.T){
 now:=time.Date(2026,11,20,10,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-4*time.Hour),"4h");a.ExpiresAt=a.CapturedAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-20*time.Minute),"fresh");b.ExpiresAt=b.CapturedAt.Add(time.Minute)
 c:=qualityFixture(now.Add(-time.Minute),"unexpired")
 d,e:=BuildDataset([]Observation{a,b,c});if e!=nil{t.Fatal(e)}
 x:=AssessPendingBacklog(d,now)
 if x[0].PendingExpired!=2||x[0].Over30Minutes!=1||x[0].Over2Hours!=1||x[0].Over24Hours!=0{t.Fatalf("backlog %+v",x)}
}
