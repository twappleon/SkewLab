package main
import("encoding/json";"io";"net/http";"net/http/httptest";"strings";"testing";"time"
"github.com/twappleon/SkewLab/internal/polymarket"
"github.com/twappleon/SkewLab/internal/strategy")
func TestComparisonPrivateAndPublicAggregate(t *testing.T){
 t.Setenv("DASHBOARD_PASSWORD","test-only-private-password");t.Setenv("PAPER_ENTRIES_ENABLED","")
 a:=NewApp(t.TempDir(),false,.02,.2,time.Second)
 now:=time.Now().UTC()
 m:=polymarket.Market{Slug:"private-btc-market",ConditionID:"private-condition",Seconds:300,End:now.Add(205*time.Second),Active:true,
 Up:polymarket.Quote{Bid:.52,Ask:.54,AskDepth:100,UpdatedAt:now},
 Down:polymarket.Quote{Bid:.44,Ask:.46,AskDepth:100,UpdatedAt:now}}
 yes,e:=a.comparison.Capture(m,strategy.Result{Ready:true,Direction:"WAIT",MomentumPct:.03,Skew:-.4},strategy.Tick{Time:now,Price:98000,Source:"Kraken"},now)
 if e!=nil||!yes{t.Fatalf("non-aligned unbiased sample missing: %v %v",yes,e)}
 if a.edge.Status().Total!=0{t.Fatal("mixed legacy selected-signal research")}
 server:=httptest.NewServer(a.handler());defer server.Close()
 rsp,e:=http.Get(server.URL+"/api/performance");if e!=nil{t.Fatal(e)}
 b,e:=io.ReadAll(rsp.Body);rsp.Body.Close();if e!=nil{t.Fatal(e)}
 if rsp.StatusCode!=200||strings.Contains(string(b),"private-condition")||strings.Contains(string(b),"private-btc-market"){t.Fatalf("public private fields exposed: %d %s",rsp.StatusCode,b)}
 var out struct{Version string `json:"version"`;Comparison struct{Total int `json:"total"`;TradingAuthorized bool `json:"trading_authorized"`} `json:"strategy_comparison"`}
 if e=json.Unmarshal(b,&out);e!=nil||out.Version!="1.3.0"||out.Comparison.Total!=1||out.Comparison.TradingAuthorized{t.Fatalf("bad public study summary: %+v %v",out,e)}
 other,e:=http.Get(server.URL+"/api/strategy-comparison");if e!=nil{t.Fatal(e)};other.Body.Close();if other.StatusCode!=401{t.Fatalf("private research unprotected: %d",other.StatusCode)}
 csv,e:=http.Get(server.URL+"/api/export.csv");if e!=nil{t.Fatal(e)};csv.Body.Close();if csv.StatusCode!=401{t.Fatalf("private CSV unprotected: %d",csv.StatusCode)}
}
