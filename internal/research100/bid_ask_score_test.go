package research100
import("testing";"time")
func TestRound047BidAskScoreRange(t *testing.T){
 now:=time.Date(2026,11,19,15,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 x:=AssessBidAskScoreRange(d,now)
 if x.Samples!=1||x.AskBrierMean==nil||x.BidBrierMean==nil||*x.AskBrierMean>=*x.BidBrierMean{t.Fatalf("wrong ask/bid scores %+v",x)}
}
