package research100
import("testing";"time")
func TestRound023FrontierCannotSeeFuture(t *testing.T){
 now:=time.Date(2026,11,20,10,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"seen");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-2*time.Hour),"futurelabel");b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.Winner="DOWN";b.ResolvedAt=now.Add(time.Minute)
 c:=qualityFixture(now.Add(time.Minute),"futurecapture")
 d,e:=BuildDataset([]Observation{a,b,c});if e!=nil{t.Fatal(e)}
 x:=InspectLabelFrontier(d,now)
 if x.Published!=1||x.NotYetPublished!=1||x.FutureCaptures!=1||x.LatestKnownPublication==nil{t.Fatalf("future leak %+v",x)}
}
