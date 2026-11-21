package research100
import("testing";"time";"github.com/twappleon/SkewLab/internal/experiment")
func pairedFixture(at time.Time,id string)experiment.Sample{
 return experiment.Sample{Slug:id,ConditionID:"condition-"+id,WindowSeconds:300,Policy:experiment.SamplingPolicy,CapturedAt:at,MarketEnd:at.Add(3*time.Minute),UpQuoteAt:at.Add(-time.Second),DownQuoteAt:at.Add(-time.Second),UpBid:.48,UpAsk:.50,UpAskDepth:100,DownBid:.48,DownAsk:.50,DownAskDepth:100,MomentumPct:.03,Skew:.3}
}
func TestRound081RejectUnsynchronizedBook(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 x:=pairedFixture(now,"m")
 if !ValidatePairedSnapshot(x,now).Valid{t.Fatal("valid pair refused")}
 x.DownQuoteAt=now.Add(-8*time.Second)
 if ValidatePairedSnapshot(x,now).Valid{t.Fatal("unsynchronized book passed")}
}
