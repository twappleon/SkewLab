package main

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/twappleon/SkewLab/internal/polymarket"
    "github.com/twappleon/SkewLab/internal/strategy"
)
func TestCoveragePublicAggregatePrivateEvidenceAndPause(t *testing.T){
    t.Setenv("DASHBOARD_PASSWORD","only-test-password")
    t.Setenv("PAPER_ENTRIES_ENABLED","")
    a:=NewApp(t.TempDir(),false,.02,.20,time.Second)
    now:=time.Now().UTC()
    m:=polymarket.Market{Slug:"private-5m-identity",ConditionID:"private-condition",
        Seconds:300,End:now.Add(205*time.Second),Active:true,
        Up:polymarket.Quote{Bid:.50,Ask:.52,AskDepth:100,UpdatedAt:now},
        Down:polymarket.Quote{Bid:.46,Ask:.48,AskDepth:100,UpdatedAt:now}}
    ok,err:=a.comparison.Capture(m,strategy.Result{Ready:true,Direction:"WAIT",MomentumPct:.025,Skew:-.3},
        strategy.Tick{Time:now,Price:98000,Source:"test"},now)
    if !ok||err!=nil{t.Fatalf("capture: %v %v",ok,err)}
    h:=a.handler()
    pub:=httptest.NewRecorder()
    h.ServeHTTP(pub,httptest.NewRequest(http.MethodGet,"/api/performance",nil))
    if pub.Code!=200{t.Fatalf("public status %d",pub.Code)}
    body:=pub.Body.String()
    if strings.Contains(body,"private-5m-identity")||strings.Contains(body,"private-condition"){t.Fatal("private data leak")}
    var out map[string]any
    if err:=json.Unmarshal(pub.Body.Bytes(),&out);err!=nil{t.Fatal(err)}
    q,ok:=out["comparison_quality"].(map[string]any)
    if !ok||out["version"]!="1.6.0"||q["trading_authorized"]!=false||out["paper_entries_enabled"]!=false{t.Fatalf("quality gate missing: %+v",out)}
    windows,ok:=q["windows"].([]any)
    if !ok||len(windows)!=2||windows[0].(map[string]any)["captured"]!=float64(1){t.Fatalf("invalid coverage: %+v",q)}
    private:=httptest.NewRecorder()
    h.ServeHTTP(private,httptest.NewRequest(http.MethodGet,"/api/strategy-comparison",nil))
    if private.Code!=401{t.Fatalf("private research exposed: %d",private.Code)}
    csv:=httptest.NewRecorder()
    h.ServeHTTP(csv,httptest.NewRequest(http.MethodGet,"/api/export.csv",nil))
    if csv.Code!=401{t.Fatalf("private CSV exposed: %d",csv.Code)}
}
