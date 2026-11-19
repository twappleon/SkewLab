package research100
import("testing";"time")
func TestRound077ThirtyTwoSecondBoundary(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Minute),"a")
 w:=ExitWitness{ID:"a",QuoteAt:a.ExpiresAt.Add(-32*time.Second),Bid:.49,BidDepth:10,Source:"CLOB"}
 x:=EvaluateNearExpiryWitness(ReplayPair{Entry:a,Exit:w})
 if !x.Evaluated||!x.WithinExitBuffer||x.SecondsToExpiry!=32{t.Fatalf("time boundary %+v",x)}
 w.QuoteAt=a.ExpiresAt.Add(-33*time.Second)
 if y:=EvaluateNearExpiryWitness(ReplayPair{Entry:a,Exit:w});!y.Evaluated||y.WithinExitBuffer{t.Fatalf("premature expiry exit %+v",y)}
}
