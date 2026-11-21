package research100
import("testing";"time";"github.com/twappleon/SkewLab/internal/experiment")
func TestRound090NeverTradeOnPairedComparison(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 s:=pairedFixture(now.Add(-time.Hour),"m");s.Winner="UP";s.ResolvedAt=s.MarketEnd.Add(time.Minute)
 x:=ReviewPairedEvidence([]experiment.Sample{s},now)
 if x.TradingAuthorized||x.Gate.TradingAuthorized||x.AsOfAudit.PublishedOutcomeRows!=1||x.Rules.Rules[0].PublishedSameSnapshots!=1{t.Fatalf("paired research unsafe %+v",x)}
}
