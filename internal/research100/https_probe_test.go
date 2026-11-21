package research100

import (
 "context"
 "net/http"
 "net/http/httptest"
 "testing"
)

func TestRound097TLSOnlyGETAndNoCredentials(t *testing.T) {
 service:=httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.Method!=http.MethodGet||r.Header.Get("Authorization")!=""||len(r.Cookies())>0 {w.WriteHeader(500);return}
  if r.URL.Path=="/health"{w.WriteHeader(200)}else{w.WriteHeader(401)}
 }))
 defer service.Close()
 client:=service.Client()
 health:=ProbeHTTPSReadOnly(context.Background(),client,service.URL+"/health",200)
 private:=ProbeHTTPSReadOnly(context.Background(),client,service.URL+"/api/export.csv",401)
 if !health.Passed||!private.Passed {t.Fatalf("unexpected checks %+v %+v",health,private)}
 invalid:=ProbeHTTPSReadOnly(context.Background(),client,"http://example.com/health",200)
 if invalid.Passed||invalid.Status!="INVALID_HTTPS_PROBE" {t.Fatalf("insecure url accepted %+v",invalid)}
}
