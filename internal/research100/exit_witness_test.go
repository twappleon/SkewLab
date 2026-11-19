package research100
import("testing";"time")
func TestRound072ExitMustBeAfterEntryAndBeforeExpiry(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 entry:=qualityFixture(now.Add(-2*time.Minute),"m")
 witness:=ExitWitness{ID:"m",QuoteAt:entry.CapturedAt.Add(time.Minute),Bid:.49,BidDepth:10,Source:"CLOB_ARCHIVED"}
 if x:=ValidateExitWitness(entry,witness,now);!x.Valid{t.Fatalf("valid post entry evidence: %+v",x)}
 witness.QuoteAt=entry.CapturedAt
 if x:=ValidateExitWitness(entry,witness,now);x.Valid{t.Fatal("entry-time quote treated as post-entry exit")}
}
