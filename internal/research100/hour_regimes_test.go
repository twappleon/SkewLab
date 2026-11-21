package research100
import("testing";"time")
func TestRound054TimeOfDayRegimeDoesNotCrossBuckets(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=qualityFixture(time.Date(2026,11,20,1,0,0,0,time.UTC),"a");a.ExpiresAt=a.CapturedAt.Add(time.Minute);a.Winner="UP";a.ResolvedAt=a.ExpiresAt.Add(time.Minute)
 b:=qualityFixture(time.Date(2026,11,20,14,0,0,0,time.UTC),"b");b.ExpiresAt=b.CapturedAt.Add(time.Minute);b.Winner="DOWN";b.ResolvedAt=b.ExpiresAt.Add(time.Minute)
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 x:=AssessHourRegimeScore(d,now)
 if x[0].Samples!=1||x[1].Samples!=0||x[2].Samples!=1||x[3].Samples!=0{t.Fatalf("regime leak %+v",x)}
}
