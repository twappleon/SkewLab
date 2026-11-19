package research100
import("testing";"time")
func TestRound069WorstDayHypotheticalOnly(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-3*time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="DOWN";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-2*time.Hour),"b");b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.Winner="DOWN";b.ResolvedAt=b.ExpiresAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 x:=AssessHypotheticalDailyLoss(d,now)
 if x.LabeledSamples!=2||x.UTCScoredDays!=1||x.WorstDailyNetUSD==nil||*x.WorstDailyNetUSD>=-1{t.Fatalf("worst loss day %+v",x)}
}
