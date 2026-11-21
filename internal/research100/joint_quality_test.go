package research100
import("testing";"time")
func TestRound037JointProblemsNotDoubleCountedAsMarkets(t *testing.T){
 now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Minute),"a");a.QuoteAt=a.CapturedAt.Add(-7*time.Second);a.Bid=.4;a.Ask=.5
 b:=qualityFixture(now.Add(-time.Minute),"b");b.QuoteAt=b.CapturedAt.Add(-7*time.Second)
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 x:=AssessJointQuoteQuality(d,now)
 if x.Observations!=2||x.StaleOverFiveSeconds!=2||x.WideOverThreeCents!=1||x.BothProblems!=1{t.Fatalf("joint count %+v",x)}
}
