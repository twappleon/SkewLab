package research100
import("testing";"time")
func TestRound021PublishedLagQuantiles(t *testing.T){
 now:=time.Date(2026,11,20,10,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-2*time.Hour),"b");b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.Winner="DOWN";b.ResolvedAt=b.ExpiresAt.Add(3*time.Minute)
 c:=qualityFixture(now.Add(-3*time.Hour),"c");c.ExpiresAt=c.CapturedAt.Add(time.Minute);c.Winner="UP";c.ResolvedAt=now.Add(time.Hour)
 d,e:=BuildDataset([]Observation{a,b,c});if e!=nil{t.Fatal(e)}
 x:=AssessLabelQuantiles(d,now)
 if x.Resolved!=2||x.MedianSeconds==nil||*x.MedianSeconds!=60||x.P95Seconds==nil||*x.P95Seconds!=180{t.Fatalf("lag boundary %+v",x)}
}
