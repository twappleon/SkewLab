package research100
import("testing";"time")
func TestRound016WideSpreadIsObservedNotFill(t *testing.T){
 now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Minute),"a")
 b:=qualityFixture(now.Add(-2*time.Minute),"b");b.Bid=.40;b.Ask=.50
 c:=qualityFixture(now.Add(time.Minute),"future");c.Bid=.20;c.Ask=.90
 d,e:=BuildDataset([]Observation{a,b,c});if e!=nil{t.Fatal(e)}
 x:=AssessObservedSpread(d,now)
 if x[0].Observed!=2||x[0].WideOverThreeCents!=1||x[0].MeanSpread==nil||*x[0].MeanSpread<=.05{t.Fatalf("wrong spread %+v",x[0])}
 if x[1].Observed!=0||x[1].MeanSpread!=nil{t.Fatal("phantom spread")}
}
