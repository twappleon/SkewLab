package research100
import("testing";"time")
func TestRound029PendingAgeExcludesResolvedAndFuture(t *testing.T){
 now:=time.Date(2026,11,20,10,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-10*time.Minute),"b");b.ExpiresAt=b.CapturedAt.Add(time.Minute)
 c:=qualityFixture(now.Add(-2*time.Hour),"c");c.ExpiresAt=c.CapturedAt.Add(time.Minute);c.Winner="UP";c.ResolvedAt=c.ExpiresAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a,b,c});if e!=nil{t.Fatal(e)}
 out:=AssessPendingAge(d,now)
 if out.PendingExpired!=2||out.Over30Minutes!=1||out.OldestSeconds==nil||*out.OldestSeconds!=3540{t.Fatalf("pending classification %+v",out)}
}
