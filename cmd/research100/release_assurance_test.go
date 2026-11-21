package main

import("encoding/json";"testing";"time";"github.com/twappleon/SkewLab/internal/research100")
func TestRound100StrictCLIAuditAndReleaseStatus(t *testing.T) {
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 o:=research100.Observation{ID:"archive-market",Market:"5m",Side:"UP",CapturedAt:now.Add(-time.Hour),ExpiresAt:now.Add(-55*time.Minute),QuoteAt:now.Add(-time.Hour-time.Second),Bid:.49,Ask:.51,BidDepth:50,AskDepth:50,Momentum:.03,Skew:.3}
 data,e:=json.Marshal([]research100.Observation{o});if e!=nil{t.Fatal(e)}
 out,e:=Execute("audit",data,now);if e!=nil{t.Fatal(e)}
 if out.ReleaseAssurance==nil||out.ReleaseAssurance.TradingAuthorized||out.TradingAuthorized||out.ReleaseAssurance.Observations!=1{t.Fatalf("unsafe report %+v",out)}
 public,e:=json.Marshal(out);if e!=nil{t.Fatal(e)}
 if e=research100.CheckPublicAggregateFields(public);e!=nil{t.Fatal(e)}
 invalid:=append(append([]byte{},data...),[]byte(" {}")...)
 if _,e=Execute("audit",invalid,now);e==nil{t.Fatal("trailing payload accepted")}
 if _,e=Execute("audit",[]byte(`[{"id":"one","secret_field":"private"}]`),now);e==nil{t.Fatal("unknown private field accepted")}
}
