package main
import (
 "encoding/json"
 "io"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "time"
 "github.com/twappleon/SkewLab/internal/analytics"
 "github.com/twappleon/SkewLab/internal/paper"
)
func TestPublicAggregatesAndPrivateTradeDetails(t *testing.T){
 t.Setenv("DASHBOARD_USER","admin");t.Setenv("DASHBOARD_PASSWORD","test-secret")
 app:=NewApp(t.TempDir(),false,.02,.2,time.Second)
 now:=time.Now().UTC()
 app.ledger.Positions=[]paper.Position{
  {ID:"private-order-id",Market:"private-market",Token:"secret-token",Status:"CLOSED",ExitTime:&now,RealizedPnL:2},
  {ID:"private-2",Status:"SETTLED",ExitTime:&now,RealizedPnL:-1},
  {ID:"private-pending",Status:"PENDING_RESOLUTION",RealizedPnL:1000},
 }
 app.ledger.TotalClosed=2;app.ledger.Wins=1;app.ledger.RealizedPnL=1
 server:=httptest.NewServer(app.handler());defer server.Close()
 resp,err:=http.Get(server.URL+"/api/performance");if err!=nil{t.Fatal(err)}
 body,err:=io.ReadAll(resp.Body);resp.Body.Close();if err!=nil{t.Fatal(err)}
 if resp.StatusCode!=200{t.Fatalf("public code %d",resp.StatusCode)}
 for _,secret:=range []string{"private-order-id","private-market","secret-token","private-pending","\"positions\":["}{if strings.Contains(string(body),secret){t.Fatalf("public data leak: %s",secret)}}
 var answer struct{Analytics analytics.Summary `json:"analytics"`}
 if err:=json.Unmarshal(body,&answer);err!=nil{t.Fatal(err)}
 if !answer.Analytics.DataComplete||answer.Analytics.Last7Days.Closed!=2||answer.Analytics.Last7Days.NetPnLUSD!=1||answer.Analytics.Last7Days.ProfitFactor==nil||*answer.Analytics.Last7Days.ProfitFactor!=2{t.Fatalf("bad aggregate: %+v",answer.Analytics)}
 private,err:=http.Get(server.URL+"/api/export.csv");if err!=nil{t.Fatal(err)};private.Body.Close()
 if private.StatusCode!=401{t.Fatalf("private CSV exposed: %d",private.StatusCode)}
}
