package research100
import("testing";"time")
func TestRound030SettlementReviewIsNeverTradeApproval(t *testing.T){
 now:=time.Date(2026,11,20,10,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"m1");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-time.Hour),"m2");b.ExpiresAt=b.CapturedAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 r:=ReviewSettlementData(d,now)
 if r.TradingAuthorized||r.Frontier.Published!=1||r.PendingAge.Over30Minutes!=1||len(r.Blockers)<2{t.Fatalf("incorrect review %+v",r)}
}
