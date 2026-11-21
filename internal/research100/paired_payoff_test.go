package research100
import("testing";"time")
func TestRound086PayoffRequiresPublishedResult(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 s:=pairedFixture(now.Add(-time.Hour),"a");s.Winner="UP";s.ResolvedAt=s.MarketEnd.Add(time.Minute)
 x:=ScorePairedPayoff(s,"UP",now)
 if !x.Scored||!x.Correct||x.NetOneShareUSD<=0{t.Fatalf("winner reward %+v",x)}
 y:=ScorePairedPayoff(s,"DOWN",now)
 if !y.Scored||y.Correct||y.NetOneShareUSD>=0{t.Fatalf("loss mislabeled %+v",y)}
 z:=ScorePairedPayoff(s,"UP",s.CapturedAt)
 if z.Scored{t.Fatal("future outcome leak")}
}
