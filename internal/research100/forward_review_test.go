package research100
import("testing";"time")
func TestRound060ForwardReportStillResearchOnly(t *testing.T){
 now:=time.Date(2026,11,20,16,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 x:=ReviewForwardResearch(d,now)
 if x.TradingAuthorized||x.Gate.TradingAuthorized||x.Status!="INSUFFICIENT_FORWARD_SAMPLES"||x.Abstentions.PublishedSameSnapshots!=1||len(x.UTCRegimes)!=4{t.Fatalf("unsafe forward report %+v",x)}
}
