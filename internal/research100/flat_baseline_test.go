package research100
import("testing";"time")
func TestRound044CoinFlipBaselineNoLookahead(t *testing.T){
 now:=time.Date(2026,11,19,15,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-time.Hour),"b");b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.Winner="DOWN";b.ResolvedAt=now.Add(time.Hour)
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 x:=AssessFlatBrierBaseline(d,now)
 if x.PublishedSamples!=1||x.MeanBrier==nil||*x.MeanBrier!=.25{t.Fatalf("flat baseline %+v",x)}
}
