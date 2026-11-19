package research100

import("testing";"time")
func qualityFixture(at time.Time,id string) Observation{
 return Observation{ID:id,Market:"5m",Side:"UP",CapturedAt:at,ExpiresAt:at.Add(3*time.Minute),QuoteAt:at.Add(-time.Second),Bid:.48,Ask:.50,BidDepth:100,AskDepth:100,Momentum:.03,Skew:.3}
}
func TestRound011QuoteAgeHonorsAsOf(t *testing.T){
 now:=time.Date(2026,11,19,1,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Minute),"a");b:=qualityFixture(now.Add(-30*time.Second),"b");b.QuoteAt=b.CapturedAt.Add(-8*time.Second)
 future:=qualityFixture(now.Add(time.Minute),"future")
 d,e:=BuildDataset([]Observation{a,b,future});if e!=nil{t.Fatal(e)}
 x:=AssessQuoteAge(d,now)
 if x.Samples!=2||x.OverFiveSeconds!=1||x.P95Seconds==nil||*x.P95Seconds!=8||x.TradingAuthorized{t.Fatalf("bad evidence %+v",x)}
 empty:=AssessQuoteAge(d,now.Add(-2*time.Minute))
 if empty.Samples!=0||empty.P95Seconds!=nil{t.Fatalf("future leaked %+v",empty)}
}
