package research100
import("testing";"time")
func TestRound080ReadOnlyReplayRequiresWitnesses(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-2*time.Minute),"a")
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 x:=ReviewExitEvidence(d,nil,now)
 if x.TradingAuthorized||x.Coverage.Missing!=1||x.Status!="NO_VALID_POST_ENTRY_BOOK_EVIDENCE"{t.Fatalf("missing evidence treated as fills %+v",x)}
}
