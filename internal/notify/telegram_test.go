package notify
import(
 "context"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "strings"
 "sync/atomic"
 "testing"
 "time"
)
func TestSendAndPersistDedup(t *testing.T) {
 var count atomic.Int32
 ts:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.Method!="POST"||r.URL.Path!="/botprivate-test-token/sendMessage"{t.Error("unexpected request");http.Error(w,"bad",400);return}
  var b struct{ ChatID string `json:"chat_id"`; Text string `json:"text"`}
  if e:=json.NewDecoder(r.Body).Decode(&b);e!=nil{t.Error(e)}
  if b.ChatID!="1234"||!strings.Contains(b.Text,"模擬"){t.Errorf("bad body: %#v",b)}
  count.Add(1);w.Header().Set("Content-Type","application/json")
  w.Write([]byte(`{"ok":true,"result":{}}`))
 }));defer ts.Close()
 dir:=t.TempDir();c:=New("private-test-token","1234",dir);c.apiBase=ts.URL
 ctx,cancel:=context.WithCancel(context.Background());defer cancel();go c.Run(ctx)
 if !c.Publish("trade-123","🧪 模擬交易"){t.Fatal("enqueue failed")}
 if c.Publish("trade-123","duplicate"){t.Fatal("duplicate queued")}
 deadline:=time.Now().Add(2*time.Second)
 for c.Status().Sent!=1&&time.Now().Before(deadline){time.Sleep(10*time.Millisecond)}
 if c.Status().Sent!=1||count.Load()!=1{t.Fatalf("not sent: %+v",c.Status())}
 if c.Publish("trade-123","duplicate"){t.Fatal("sent duplicate")}
 c2:=New("private-test-token","1234",dir)
 if c2.Publish("trade-123","duplicate after restart"){t.Fatal("persisted duplicate")}
}
func TestDisableWithoutSecretsAndNeverSendToBotName(t *testing.T){
 t.Setenv("TELEGRAM_BOT_TOKEN","");t.Setenv("TELEGRAM_CHAT_ID","");t.Setenv("TELEGRAM_OWNER_ID","");t.Setenv("TELEGRAM_TARGET_USERNAME","")
 if FromEnv(t.TempDir())!=nil{t.Fatal("enabled without configuration")}
 t.Setenv("TELEGRAM_BOT_TOKEN","dummy");t.Setenv("TELEGRAM_CHAT_ID","@crytotradebot")
 if FromEnv(t.TempDir())!=nil{t.Fatal("bot should be sender, not recipient")}
}
func TestSecretNeverLeakedInErrors(t *testing.T){
 ts:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){http.Error(w,"private-token-error",401)}));defer ts.Close()
 c:=New("super-secret-token","1234",t.TempDir());c.apiBase=ts.URL
 err:=c.send(context.Background(),"test")
 if err==nil||strings.Contains(err.Error(),"super-secret-token")||strings.Contains(err.Error(),"private-token-error"){t.Fatalf("unsafe error: %v",err)}
}
func TestEnvOwnerFallback(t *testing.T){
 t.Setenv("TELEGRAM_BOT_TOKEN","some-token");t.Setenv("TELEGRAM_CHAT_ID","");t.Setenv("TELEGRAM_OWNER_ID","4567")
 c:=FromEnv(t.TempDir());if c==nil||c.chatID!="4567"{t.Fatal("owner fallback failed")}
}
func TestQueueBounded(t *testing.T){
 c:=New("t","1",t.TempDir())
 for i:=0;i<maxQueue;i++ {if !c.Publish(string(rune(0x100+i)),"msg"){t.Fatal("early rejection")}}
 if c.Publish("overflow","msg"){t.Fatal("overflow should not block strategy")}
}
