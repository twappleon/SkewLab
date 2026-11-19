package research100
import("testing";"time")
func TestRound058DecisionCountKeepsAbstentions(t *testing.T){
 now:=time.Date(2026,11,20,15,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-time.Hour),"b");b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.Winner="DOWN";b.ResolvedAt=b.ExpiresAt.Add(time.Minute);b.Momentum=.01
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 x:=AssessAbstentions(d,now)
 if x.PublishedSameSnapshots!=2||x.MomentumDecisions!=1||x.MomentumSkewDecisions!=1||x.MomentumAbstentions!=1||x.MomentumSkewAbstentions!=1{t.Fatalf("unfair denominator %+v",x)}
}
