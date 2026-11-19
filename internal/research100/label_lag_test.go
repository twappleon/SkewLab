package research100
import("testing";"time")
func TestRound015NoFutureLabelLeak(t *testing.T){
 now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-2*time.Hour),"pending");a.ExpiresAt=a.CapturedAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-3*time.Hour),"labeled");b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.ResolvedAt=now.Add(10*time.Minute);b.Winner="UP"
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 x:=AssessLabelLag(d,now)
 if x[0].Resolved!=0||x[0].Pending!=2||x[0].Overdue30Minutes!=2{t.Fatalf("future label leaked %+v",x[0])}
 x=AssessLabelLag(d,now.Add(time.Hour))
 if x[0].Resolved!=1||x[0].Overdue30Minutes!=1||x[0].MaximumPublishedLagSeconds==nil{t.Fatalf("lag failed %+v",x[0])}
}
