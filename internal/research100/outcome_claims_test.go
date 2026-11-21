package research100
import("testing";"time")
func TestRound027ClaimsCannotProveIndependentOracle(t *testing.T){
 now:=time.Date(2026,11,20,10,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"market");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 out:=CompareOutcomeClaims(d,[]ExternalOutcomeClaim{{ID:"market",Winner:"DOWN",ObservedAt:now,Source:"unverified claim"},{ID:"absent",Winner:"UP",ObservedAt:now}},now)
 if out.Disagreements!=1||out.UnknownClaims!=1||out.IndependentProofVerified{t.Fatalf("unsafe trust %+v",out)}
}
