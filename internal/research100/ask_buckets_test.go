package research100
import("testing";"time")
func TestRound034PriceBucketStability(t *testing.T){
 now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Minute),"a");a.Bid=.19;a.Ask=.20
 b:=qualityFixture(now.Add(-time.Minute),"b");b.Bid=.99;b.Ask=.999
 c:=qualityFixture(now.Add(-time.Minute),"c");c.Market="15m"
 d,e:=BuildDataset([]Observation{a,b,c});if e!=nil{t.Fatal(e)}
 x:=AssessAskBuckets(d,now)
 if x[0].BucketCounts[1]!=1||x[0].BucketCounts[4]!=1||x[1].BucketCounts[2]!=1{t.Fatalf("bucket error %+v",x)}
}
