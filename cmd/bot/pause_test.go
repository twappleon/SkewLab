package main
import("encoding/json";"net/http";"net/http/httptest";"testing";"time")
func TestPaperEntryGateFailClosed(t *testing.T){
 t.Setenv("PAPER_ENTRIES_ENABLED","")
 if paperEntriesEnabled(){t.Fatal("enabled by default")}
 t.Setenv("PAPER_ENTRIES_ENABLED","true")
 if paperEntriesEnabled(){t.Fatal("must require exact 1")}
 t.Setenv("PAPER_ENTRIES_ENABLED","1")
 if !paperEntriesEnabled(){t.Fatal("explicit opt-in failed")}
}
func TestPublicPauseReadOnly(t *testing.T){
 t.Setenv("PAPER_ENTRIES_ENABLED","")
 t.Setenv("DASHBOARD_PASSWORD","test-only-secret")
 a:=NewApp(t.TempDir(),false,.02,.2,time.Second)
 w:=httptest.NewRecorder()
 a.handler().ServeHTTP(w,httptest.NewRequest(http.MethodGet,"/api/performance",nil))
 if w.Code!=200{t.Fatalf("public summary %d",w.Code)}
 var d map[string]any
 if err:=json.Unmarshal(w.Body.Bytes(),&d);err!=nil{t.Fatal(err)}
 if d["paper_entries_enabled"]!=false || d["entry_pause_reason"]==""{t.Fatalf("not clearly paused: %+v",d)}
 if _,ok:=d["positions"];ok{t.Fatal("private trade details leaked")}
}
