package research100

import("testing";"time")
func TestRound100AggregateReleaseWithoutTradingAuthorization(t *testing.T) {
 at:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 empty:=ReviewReleaseAssurance(Dataset{},at)
 if empty.TradingAuthorized||empty.Status!="RESEARCH_NOT_READY"||len(empty.Blockers)<3||len(empty.CanonicalSHA256)!=64 {
  t.Fatalf("empty research cannot pass %+v",empty)
 }
 o:=qualityFixture(at.Add(-time.Hour),"m")
 d,e:=BuildDataset([]Observation{o});if e!=nil{t.Fatal(e)}
 r:=ReviewReleaseAssurance(d,at)
 if r.TradingAuthorized||r.Observations!=1||r.PublishedLabels!=0||!r.StrictInputRequired {t.Fatalf("invalid release state %+v",r)}
}
