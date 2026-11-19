package research100
import("testing";"time")
func TestRound074PairedExitIncludesBothSideCosts(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-2*time.Minute),"a")
 w:=ExitWitness{ID:"a",QuoteAt:a.CapturedAt.Add(time.Minute),Bid:.5,BidDepth:10,Source:"CLOB"}
 x:=CalculatePairedExitCost(ReplayPair{Entry:a,Exit:w},.002,25)
 if x.Status!="MODELED_SINGLE_QUOTE_NOT_PROVEN_EXIT_FILL"||x.ModeledOneShareNetUSD>=0{t.Fatalf("missing fees %+v",x)}
}
