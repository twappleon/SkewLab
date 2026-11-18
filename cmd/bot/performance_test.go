package main

import (
 "encoding/json"
 "io"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "time"
 "github.com/twappleon/SkewLab/internal/paper"
)

func TestPublicPaperSummaryWithoutPasswordAndPrivateArchivesProtected(t *testing.T){
 t.Setenv("DASHBOARD_USER","admin")
 t.Setenv("DASHBOARD_PASSWORD","secret-password")
 a:=NewApp(t.TempDir(),false,.02,.2,time.Second)
 a.ledger.RealizedPnL=12.5
 a.ledger.TotalClosed=3
 a.ledger.Wins=2
 a.ledger.Positions=[]paper.Position{{ID:"sensitive-id",Market:"private-market",Status:"CLOSED"}}
 s:=httptest.NewServer(a.handler())
 defer s.Close()
 for _,path:=range []string{"/","/performance","/api/performance","/health"}{
  r,e:=http.Get(s.URL+path);if e!=nil{t.Fatal(e)};r.Body.Close()
  if r.StatusCode!=200{t.Fatalf("%s: expected 200, got %d",path,r.StatusCode)}
 }
 r,e:=http.Get(s.URL+"/api/performance");if e!=nil{t.Fatal(e)}
 body,_:=io.ReadAll(r.Body);r.Body.Close()
 if strings.Contains(string(body),"sensitive-id")||strings.Contains(string(body),"private-market")||strings.Contains(string(body),"\"positions\":["){t.Fatalf("public summary leaks private data: %s",body)}
 var data struct{ClosedTrades int `json:"closed_trades"`;Wins int `json:"wins"`;RealizedPnL float64 `json:"realized_pnl_usd"`}
 if e:=json.Unmarshal(body,&data);e!=nil{t.Fatal(e)}
 if data.ClosedTrades!=3||data.Wins!=2||data.RealizedPnL!=12.5{t.Fatalf("bad summary: %+v",data)}
 for _,path:=range []string{"/admin","/api/state","/api/export.csv","/api/observations.jsonl","/api/feeds","/api/chainlink-twap.jsonl"}{
  r,e:=http.Get(s.URL+path);if e!=nil{t.Fatal(e)};r.Body.Close()
  if r.StatusCode!=401{t.Fatalf("%s unprotected: %d",path,r.StatusCode)}
 }
 req,_:=http.NewRequest("GET",s.URL+"/admin",nil);req.SetBasicAuth("admin","secret-password")
 r,e=http.DefaultClient.Do(req);if e!=nil{t.Fatal(e)};r.Body.Close()
 if r.StatusCode!=200{t.Fatalf("admin auth failed: %d",r.StatusCode)}
}
