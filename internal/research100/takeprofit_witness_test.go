package research100
import("testing";"time")
func TestRound075TakeProfitCannotUseInstantQuote(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-2*time.Minute),"a")
 w:=ExitWitness{ID:"a",QuoteAt:a.CapturedAt.Add(2*time.Second),Bid:.9,BidDepth:10,Source:"CLOB"}
 x:=EvaluateTakeProfitWitness(ReplayPair{Entry:a,Exit:w},1)
 if x.Evaluated||x.Triggered{t.Fatalf("instant hindsight take profit %+v",x)}
 w.QuoteAt=a.CapturedAt.Add(time.Minute)
 x=EvaluateTakeProfitWitness(ReplayPair{Entry:a,Exit:w},1)
 if !x.Evaluated||!x.Triggered{t.Fatalf("missing high bid potential %+v",x)}
}
