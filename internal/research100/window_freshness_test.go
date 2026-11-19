package research100
import("testing";"time")
func TestRound017FreshnessIsolatedByMarket(t *testing.T){
 now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-25*time.Hour),"old5")
 b:=qualityFixture(now.Add(-time.Minute),"new15");b.Market="15m";b.QuoteAt=b.CapturedAt.Add(-7*time.Second)
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 x:=AssessWindowFreshness(d,now)
 if !x[0].StaleMonitoring||x[1].StaleMonitoring||x[1].OverFiveSeconds!=1||x[0].Observed!=1{t.Fatalf("freshness not separated %+v",x)}
}
