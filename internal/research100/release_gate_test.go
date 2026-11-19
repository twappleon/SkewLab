package research100

import (
 "testing"
 "time"
)
func TestRound099FailClosedReleaseGate(t *testing.T) {
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 empty:=ReviewReleaseGate(Dataset{},now)
 if empty.Status!="RESEARCH_NOT_READY"||len(empty.Blockers)<3||empty.TradingAuthorized {t.Fatalf("empty study unlocked %+v",empty)}
 o:=qualityFixture(now.Add(-time.Hour),"m")
 d,e:=BuildDataset([]Observation{o});if e!=nil{t.Fatal(e)}
 limited:=ReviewReleaseGate(d,now)
 if limited.TradingAuthorized||limited.PublishedLabels!=0||len(limited.Blockers)<3{t.Fatalf("unproven evidence unlocked %+v",limited)}
}
