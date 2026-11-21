package research100
import("testing";"time")
func TestRound052SameTimestampCannotBePriorOutcome(t *testing.T){
 now:=time.Date(2026,11,20,16,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Hour),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(now.Add(-time.Hour),"b");b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.Winner="UP";b.ResolvedAt=now
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 x:=AssessPriorLabels(d,"5m","UP",now)
 if x.PublishedBefore!=1||x.CorrectBefore!=1||x.DistinctUTCDays!=1||x.TradingAuthorized{t.Fatalf("strict prior labels %+v",x)}
}
