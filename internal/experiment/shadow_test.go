package experiment

import("testing";"time";"math")

func shadowFixture(at time.Time,i int,seconds int,winner string,known time.Time)Sample{
 return Sample{Slug:time.Unix(int64(i),0).Format("20060102T150405")+"-test",ConditionID:"c",
 WindowSeconds:seconds,Policy:SamplingPolicy,CapturedAt:at,
 MarketEnd:at.Add(time.Duration(seconds)*time.Second-100*time.Second),
 UpQuoteAt:at,DownQuoteAt:at,UpBid:.38,UpAsk:.40,UpAskDepth:100,
 DownBid:.58,DownAsk:.60,DownAskDepth:100,
 MomentumPct:.04,Skew:.40,BTCSpotSource:"Kraken",Winner:winner,ResolvedAt:known}
}
func TestShadowEmptyNoTrading(t *testing.T){
 x:=EvaluateShadow(nil,time.Now())
 if x.TradingAuthorized||x.Status!="NO_VALID_RESEARCH_SAMPLES"{t.Fatal(x)}
 for _,w:=range x.Windows{for _,r:=range w.Rules{if r.MeanExcessUSD!=nil||r.EligibleCandidates!=0{t.Fatal("fabricated result")}}}
}
func TestShadowChronologyAndMatchedBaseline(t *testing.T){
 base:=time.Date(2026,10,1,0,0,0,0,time.UTC);var ss []Sample
 for i:=0;i<110;i++{at:=base.Add(time.Duration(i)*10*time.Minute);ss=append(ss,shadowFixture(at,i,300,"UP",at.Add(6*time.Minute)))}
 x:=EvaluateShadow(ss,base.Add(1110*time.Minute));r:=x.Windows[0].Rules[0]
 if r.BlockedInsufficientHistory!=100||r.EligibleCandidates!=10||r.ResolvedCandidates!=10||r.Winners!=10||r.MeanExcessUSD==nil||*r.MeanExcessUSD<=0||x.TradingAuthorized{
  t.Fatalf("incorrect shadow controls: %+v",r)
 }
 if x.Windows[1].Rules[0].EligibleCandidates!=0{t.Fatal("window leakage")}
 for i:=range ss{ss[i].ResolvedAt=base.Add(3000*time.Minute)}
 x=EvaluateShadow(ss,base.Add(3100*time.Minute))
 if x.Windows[0].Rules[0].EligibleCandidates!=0{t.Fatal("used future resolution")}
}
func TestShadowBadSpreadAndNoEdge(t *testing.T){
 base:=time.Date(2026,10,1,0,0,0,0,time.UTC);var ss []Sample
 for i:=0;i<110;i++{at:=base.Add(time.Duration(i)*10*time.Minute);ss=append(ss,shadowFixture(at,i,300,"UP",at.Add(6*time.Minute)))}
 ss[105].UpBid=.29
 x:=EvaluateShadow(ss,base.Add(1110*time.Minute))
 if x.Windows[0].Rules[0].BlockedSpreadOrDepth!=1{t.Fatal("missed spread")}
 for i:=range ss{ss[i].Winner="DOWN"}
 x=EvaluateShadow(ss,base.Add(1110*time.Minute))
 r:=x.Windows[0].Rules[0]
 if r.BlockedNoConservativeEdge!=9||r.EligibleCandidates!=0||math.IsNaN(r.ShadowNetPerShareUSD){t.Fatal("missing cost gate")}
}
