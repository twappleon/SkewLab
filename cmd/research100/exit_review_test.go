package main
import("encoding/json";"strings";"testing";"time";"github.com/twappleon/SkewLab/internal/research100")
func TestRound080CLIOnlyPairsExplicitWitnesses(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 o:=research100.Observation{ID:"private-candidate",Market:"5m",Side:"UP",CapturedAt:now.Add(-2*time.Minute),ExpiresAt:now.Add(time.Minute),QuoteAt:now.Add(-2*time.Minute-time.Second),Bid:.49,Ask:.51,BidDepth:50,AskDepth:50,Momentum:.03,Skew:.3}
 b,e:=json.Marshal([]research100.Observation{o});if e!=nil{t.Fatal(e)}
 none,e:=ExecuteWithWitnesses("audit",b,nil,now);if e!=nil{t.Fatal(e)}
 if none.ExitResearch==nil||none.ExitResearch.Coverage.Missing!=1||none.TradingAuthorized{t.Fatal("missing witnesses accepted")}
 w:=research100.ExitWitness{ID:"private-candidate",QuoteAt:o.CapturedAt.Add(time.Minute),Bid:.50,BidDepth:10,Source:"ARCHIVE"}
 raw,e:=json.Marshal([]research100.ExitWitness{w});if e!=nil{t.Fatal(e)}
 got,e:=ExecuteWithWitnesses("audit",b,raw,now);if e!=nil{t.Fatal(e)}
 public,e:=json.Marshal(got.ExitResearch);if e!=nil{t.Fatal(e)}
 if got.ExitResearch.Coverage.Matched!=1||got.TradingAuthorized||strings.Contains(string(public),"private-candidate"){t.Fatalf("private replay exposed %+v",got)}
}
