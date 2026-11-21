package research100
import("testing")
func TestRound065FiveDollarStakeCanExceedTwoDollarDailyStop(t *testing.T){
 a:=CheckStakeEnvelope([]float64{5})
 if a.WithinDailyTwoDollarLoss||!a.WithinThirtyDollarExposure||a.TradingAuthorized{t.Fatalf("unsafe risk cap %+v",a)}
 b:=CheckStakeEnvelope([]float64{1,.5})
 if !b.WithinDailyTwoDollarLoss||b.TradingAuthorized{t.Fatalf("bad small stake %+v",b)}
}
