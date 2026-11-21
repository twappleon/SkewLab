package research100
import("testing";"time")
func TestRound062SlippageGridShouldCostMore(t *testing.T){
 at:=time.Date(2026,11,19,0,0,0,0,time.UTC)
 x:=AssessSlippageSensitivity(.5,at)
 if len(x.Costs)!=3||!x.NonDecreasing||!(x.Costs[2].ModeledCostUSD>x.Costs[1].ModeledCostUSD){t.Fatalf("bad stress %+v",x)}
}
