package research100
import("testing";"time";"github.com/twappleon/SkewLab/internal/experiment")
func TestRound089PositiveOneDayCannotUnlockTrade(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 s:=pairedFixture(now.Add(-time.Hour),"m");s.Winner="UP";s.ResolvedAt=s.MarketEnd.Add(time.Minute)
 x:=CheckPairedEvidenceGate([]experiment.Sample{s},now)
 if x.TradingAuthorized||x.PublishedSamples!=1||x.DistinctPublishedUTCDays!=1||x.Status!="INSUFFICIENT_PAIRED_FORWARD_EVIDENCE"{t.Fatalf("unsafe release %+v",x)}
}
