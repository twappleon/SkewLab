package research100
import("testing";"time")
func TestRound067ExpensiveQuotesRequireHigherWinRates(t *testing.T){
 now:=time.Date(2026,11,19,0,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Minute),"a");a.Bid=.79;a.Ask=.81
 b:=qualityFixture(now.Add(-time.Minute),"b");b.Bid=.29;b.Ask=.31
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 out:=AssessQuoteBreakEven(d,now)
 if out.Observations!=2||out.AboveSeventyFivePctRequired!=1{t.Fatalf("wrong win requirement %+v",out)}
}
