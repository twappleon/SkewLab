package main

import (
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "testing"
 "time"
 "github.com/twappleon/SkewLab/internal/polymarketv2"
)

func TestPolymarketPublicResearchOnly(t *testing.T) {
 t.Setenv("DASHBOARD_USER","operator")
 t.Setenv("DASHBOARD_PASSWORD","secret")
 original:=polyResearchState.read()
 t.Cleanup(func(){polyResearchState.write(original)})
 polyResearchState.write(polySnapshot{
  Mode:"READ_ONLY_PAPER_RESEARCH",BankrollUSD:100,MaxPositionUSD:5,MaxExposureUSD:30,DailyLossUSD:2,
  AsOf:time.Now().UTC(),Error:"internal upstream secret must not be exposed",
  Candidates:[]polyCandidate{{Question:"Test market",URL:"https://polymarket.com/event/test",Decision:"NO_TRADE",Quote:polymarketv2.Quote{TokenID:"private-internal-token",BuyVWAP:0.5,Executable:true}}},
 })
 app:=NewApp(t.TempDir(),true,.02,.2,time.Second)
 srv:=httptest.NewServer(app.handler())
 defer srv.Close()
 for _,path:=range []string{"/polymarket","/api/polymarket"}{
  resp,err:=http.Get(srv.URL+path)
  if err!=nil{t.Fatal(err)}
  resp.Body.Close()
  if resp.StatusCode!=http.StatusOK{t.Fatalf("%s public research should be 200; got %d",path,resp.StatusCode)}
 }
 resp,err:=http.Get(srv.URL+"/api/polymarket")
 if err!=nil{t.Fatal(err)}
 defer resp.Body.Close()
 var got polySnapshot
 if err:=json.NewDecoder(resp.Body).Decode(&got);err!=nil{t.Fatal(err)}
 if got.Error!=""||len(got.Candidates)!=1||got.Candidates[0].Quote.TokenID!=""{t.Fatalf("public API leaked internal fields: %+v",got)}
 for _,path:=range []string{"/admin","/api/state","/api/export.csv","/api/observations.jsonl","/api/feeds","/api/diagnostics"}{
  resp,err:=http.Get(srv.URL+path)
  if err!=nil{t.Fatal(err)}
  resp.Body.Close()
  if resp.StatusCode!=http.StatusUnauthorized{t.Fatalf("%s must remain private: got %d",path,resp.StatusCode)}
 }
}
