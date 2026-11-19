package research100
import("testing";"time")
func TestRound078IdenticalWitnessPolicyComparison(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-2*time.Minute),"a");a.Bid=.49;a.Ask=.51
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 w:=ExitWitness{ID:"a",QuoteAt:a.CapturedAt.Add(time.Minute),Bid:.49,BidDepth:20,Source:"CLOB"}
 x:=CompareStopPolicies(d,[]ExitWitness{w},now)
 if x.TradingAuthorized||x.MatchedWitnesses!=1||x.FivePercentTriggers!=1||x.TenPercentTriggers>1{t.Fatalf("comparison %+v",x)}
}
