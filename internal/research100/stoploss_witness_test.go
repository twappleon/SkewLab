package research100
import("testing";"time")
func TestRound076CostOnlyStopIsNotGuaranteedLossLimit(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-2*time.Minute),"a");a.Bid=.49;a.Ask=.51
 w:=ExitWitness{ID:"a",QuoteAt:a.CapturedAt.Add(time.Minute),Bid:.49,BidDepth:20,Source:"CLOB"}
 x:=EvaluateStopLossWitness(ReplayPair{Entry:a,Exit:w},5)
 if !x.Evaluated||!x.Triggered||x.PnLPct>=-5{t.Fatalf("cost stop missed %+v",x)}
}
