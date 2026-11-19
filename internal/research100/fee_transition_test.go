package research100
import("testing";"time")
func TestRound061FeeCoefficientDates(t *testing.T){
 change:=time.Date(2026,11,20,23,59,0,0,time.UTC)
 x:=CompareFeeTransition(.5,change.Add(-time.Second),change)
 if x.Status!="DATED_FEE_MODEL_ONLY_NOT_EXCHANGE_CONFIRMATION"||!x.ReducedAfterTransition||x.NewModeledCostUSD>=x.OldModeledCostUSD{t.Fatalf("fee switch %+v",x)}
}
