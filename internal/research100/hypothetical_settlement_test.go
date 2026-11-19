package research100
import("testing";"time")
func TestRound068HypotheticalPayoutCannotOverwriteLedger(t *testing.T){
 now:=time.Date(2026,11,19,0,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="DOWN";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 x:=AssessHypotheticalSettlement(d,now)
 if x[0].Scored!=1||x[0].Wins!=0||x[0].ModeledNetUSD>=0||x[1].Scored!=0{t.Fatalf("counterfactual mislabeled %+v",x)}
}
