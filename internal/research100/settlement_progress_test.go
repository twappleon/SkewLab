package research100
import("testing";"time")
func TestRound025ResolutionDenominatorIsMaturedOnly(t *testing.T){
 now:=time.Date(2026,11,20,10,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"closed");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-time.Hour),"pending");b.ExpiresAt=b.CapturedAt.Add(time.Minute)
 c:=qualityFixture(now.Add(-time.Minute),"unexpired")
 d,e:=BuildDataset([]Observation{a,b,c});if e!=nil{t.Fatal(e)}
 x:=AssessSettlementProgress(d,now)
 if x[0].MaturedMarkets!=2||x[0].Published!=1||x[0].Pending!=1||x[0].PublishedFraction==nil||*x[0].PublishedFraction!=.5{t.Fatalf("wrong resolution rate %+v",x[0])}
 if x[1].PublishedFraction!=nil{t.Fatal("invented empty denominator")}
}
