package research100
import("testing";"time")
func TestRound042ReliabilityBinsEmptyUndefined(t *testing.T){
 now:=time.Date(2026,11,19,15,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="DOWN";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-time.Hour),"b");b.Bid=.79;b.Ask=.81;b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.Winner="UP";b.ResolvedAt=b.ExpiresAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 x:=AssessReliabilityBins(d,now)
 if x[2].Count!=1||x[2].Wins!=0||x[4].Count!=1||x[4].Wins!=1||x[0].ObservedWinFraction!=nil{t.Fatalf("bins %+v",x)}
}
