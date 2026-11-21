package research100
import("testing";"time")
func TestRound073NoBestQuoteCherryPicking(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-2*time.Minute),"a")
 b:=qualityFixture(now.Add(-2*time.Minute),"b")
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 w:=ExitWitness{ID:"a",QuoteAt:a.CapturedAt.Add(time.Minute),Bid:.5,BidDepth:10,Source:"CLOB"}
 x:=PairExitEvidence(d,[]ExitWitness{w,w},now)
 if x.ValidPairs!=0||x.DuplicatedWitnesses!=1||x.MissingWitnesses!=1{t.Fatalf("chosen best duplicate %+v",x)}
}
