package research100
import("testing";"time";"github.com/twappleon/SkewLab/internal/experiment")
func TestRound087ThreeRulesSameDenominator(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 s:=pairedFixture(now.Add(-time.Hour),"m");s.Winner="UP";s.ResolvedAt=s.MarketEnd.Add(time.Minute);s.DownBid=.30;s.DownAsk=.32;s.Skew=-.3
 r:=ComparePairedRules([]experiment.Sample{s},now)
 if len(r.Rules)!=3||r.Rules[0].PublishedSameSnapshots!=1||r.Rules[1].PublishedSameSnapshots!=1||r.Rules[2].PublishedSameSnapshots!=1||r.Rules[2].Abstentions!=1||r.TradingAuthorized{t.Fatalf("unfair comparison %+v",r)}
}
