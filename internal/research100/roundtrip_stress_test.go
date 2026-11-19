package research100
import("testing";"time")
func TestRound064StressMustNeverImproveWhenCostsIncrease(t *testing.T){
 now:=time.Date(2026,11,19,0,0,0,0,time.UTC)
 o:=qualityFixture(now.Add(-time.Minute),"o");o.Bid=.49;o.Ask=.51
 d,e:=BuildDataset([]Observation{o});if e!=nil{t.Fatal(e)}
 r:=AssessRoundTripStress(d,now)
 if len(r.Cases)!=3||r.Cases[0].MeanModeledLossPct==nil||r.Cases[1].MeanModeledLossPct==nil||*r.Cases[1].MeanModeledLossPct<=*r.Cases[0].MeanModeledLossPct{t.Fatalf("wrong stress %+v",r)}
}
