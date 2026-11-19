package main
import("encoding/json";"testing";"time";"github.com/twappleon/SkewLab/internal/experiment")
func TestRound090CLIComparisonOnlyReleasesAggregates(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 s:=experiment.Sample{Slug:"private-m90",ConditionID:"private-condition",WindowSeconds:300,Policy:experiment.SamplingPolicy,CapturedAt:now.Add(-time.Hour),MarketEnd:now.Add(-57*time.Minute),UpQuoteAt:now.Add(-time.Hour-time.Second),DownQuoteAt:now.Add(-time.Hour-time.Second),UpBid:.48,UpAsk:.5,UpAskDepth:100,DownBid:.48,DownAsk:.5,DownAskDepth:100,MomentumPct:.03,Skew:.3,Winner:"UP",ResolvedAt:now.Add(-56*time.Minute)}
 b,e:=json.Marshal([]experiment.Sample{s});if e!=nil{t.Fatal(e)}
 x,e:=Execute("comparison",b,now);if e!=nil{t.Fatal(e)}
 if x.PairedResearch==nil||x.PairedResearch.TradingAuthorized||x.TradingAuthorized||x.PairedResearch.AsOfAudit.PublishedOutcomeRows!=1{t.Fatalf("bad comparison gate %+v",x)}
}
