package research100
import("testing";"time")
func TestRound046SameSnapshotBrierDelta(t *testing.T){
 now:=time.Date(2026,11,19,15,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a")
 a.Bid=.69;a.Ask=.71
 a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 got:=AssessPairedBrier(d,now)
 if got.Samples!=1||got.MeanQuoteMinusCoinFlip==nil||*got.MeanQuoteMinusCoinFlip>=0{t.Fatalf("paired Brier is not better than fixed baseline %+v",got)}
 empty:=AssessPairedBrier(d,now.Add(-2*time.Hour));if empty.MeanQuoteMinusCoinFlip!=nil{t.Fatal("future label leaked")}
}
