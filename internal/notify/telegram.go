// Package notify sends low-volume, deduplicated, paper-only operational alerts.
// It never executes trades or exposes the Telegram bot token in logs.
package notify

import (
 "bytes"
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "log"
 "net/http"
 "os"
 "path/filepath"
 "strings"
 "sync"
 "time"
 "github.com/twappleon/SkewLab/internal/storage"
)

const maxQueue = 128

type Event struct { ID string `json:"id"`; Text string `json:"text"` }
type Delivery struct { ID string `json:"id"`; Time time.Time `json:"time"` }
type Status struct {
 Enabled bool `json:"enabled"`
 Sent int `json:"sent"`
 Queued int `json:"queued"`
 LastSent time.Time `json:"last_sent,omitempty"`
 LastError string `json:"last_error,omitempty"`
}
type Client struct {
 botToken string
 chatID string
 apiBase string
 dataDir string
 client *http.Client
 queue chan Event
 mu sync.Mutex
 pending map[string]bool
 delivered map[string]bool
 status Status
}
// FromEnv requires the existing bot's token and an authorized recipient.
// @crytotradebot is the sender's username, never the destination.
func FromEnv(dataDir string) *Client {
 token := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
 chat := strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID"))
 if chat == "" { chat = strings.TrimSpace(os.Getenv("TELEGRAM_OWNER_ID")) }
 if chat == "" { chat = strings.TrimSpace(os.Getenv("TELEGRAM_TARGET_USERNAME")) }
 if token == "" || chat == "" || strings.EqualFold(chat,"@crytotradebot") {
  log.Print("Telegram notifications disabled: configure TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID (numeric owner or group ID)")
  return nil
 }
 return New(token,chat,dataDir)
}
func New(token,chat,dataDir string) *Client {
 c := &Client{botToken:token,chatID:chat,apiBase:"https://api.telegram.org",dataDir:dataDir,
 client:&http.Client{Timeout:8*time.Second},queue:make(chan Event,maxQueue),
 pending:make(map[string]bool),delivered:make(map[string]bool),status:Status{Enabled:true}}
 records, err := storage.ReadRecentJSONL[Delivery](filepath.Join(dataDir,"telegram-delivered.jsonl"),5000)
 if err != nil {log.Print("Telegram delivery history unreadable; check filesystem permissions")}
 for _,d := range records { if d.ID!="" {c.delivered[d.ID]=true} }
 return c
}
func (c *Client) Status() Status {
 if c==nil{return Status{}}
 c.mu.Lock();defer c.mu.Unlock()
 s:=c.status;s.Queued=len(c.queue)
 return s
}
// Publish never blocks trading under a slow/unavailable Telegram API.
func (c *Client) Publish(id, body string) bool {
 if c==nil || id=="" || body=="" {return false}
 if len(body)>4000 {body=body[:4000]}
 c.mu.Lock();defer c.mu.Unlock()
 if c.delivered[id] || c.pending[id] {return false}
 ev:=Event{ID:id,Text:body}
 select{
 case c.queue<-ev:
  c.pending[id]=true
  return true
 default:
  c.status.LastError="notification queue full"
  log.Print("Telegram queue full; notification dropped (paper strategy unaffected)")
  return false
 }
}
// Run is isolated from market-data and paper-execution locks.
func (c *Client) Run(ctx context.Context){
 if c==nil{return}
 for{
  select{
  case <-ctx.Done():return
  case ev:=<-c.queue:
   err:=c.retry(ctx,ev.Text)
   // Journal before showing success; a failed journal write is visible in logs.
   if err==nil{
    if e:=storage.AppendJSONL(filepath.Join(c.dataDir,"telegram-delivered.jsonl"),Delivery{ID:ev.ID,Time:time.Now().UTC()},4<<20);e!=nil{
     log.Print("Telegram delivery journal write failed; de-dup across restart not guaranteed")
    }
   }
   c.mu.Lock()
   delete(c.pending,ev.ID)
   if err==nil{
    c.delivered[ev.ID]=true;c.status.Sent++;c.status.LastSent=time.Now().UTC();c.status.LastError=""
   }else{
    c.status.LastError=err.Error()
    log.Printf("Telegram notification delivery failed: %s",err)
   }
   c.mu.Unlock()
  }
 }
}
func (c *Client) retry(ctx context.Context, text string) error {
 var err error
 for i:=0;i<3;i++{
  if ctx.Err()!=nil{return ctx.Err()}
  err=c.send(ctx,text)
  if err==nil{return nil}
  if i<2 {select{case <-ctx.Done():return ctx.Err();case <-time.After(time.Duration(i+1)*2*time.Second):}}
 }
 return err
}
func (c *Client) send(ctx context.Context, message string) error {
 b,e:=json.Marshal(map[string]any{"chat_id":c.chatID,"text":message,"disable_web_page_preview":true})
 if e!=nil{return errors.New("telegram request encode failed")}
 // Do not print request URL: the URL embeds the private bot token.
 req,e:=http.NewRequestWithContext(ctx,http.MethodPost,c.apiBase+"/bot"+c.botToken+"/sendMessage",bytes.NewReader(b))
 if e!=nil{return errors.New("telegram request setup failed")}
 req.Header.Set("Content-Type","application/json")
 resp,e:=c.client.Do(req)
 if e!=nil{return errors.New("telegram transport unavailable")}
 defer resp.Body.Close()
 if resp.StatusCode!=http.StatusOK{return fmt.Errorf("telegram HTTP %d",resp.StatusCode)}
 var result struct{OK bool `json:"ok"`}
 if e=json.NewDecoder(resp.Body).Decode(&result);e!=nil||!result.OK{return errors.New("telegram API did not acknowledge message")}
 return nil
}
// SendTest sends only when explicitly requested by the user; CI never calls it.
func (c *Client) SendTest(ctx context.Context) error {
 if c==nil{return errors.New("telegram configuration missing")}
 return c.send(ctx,"🧪 SkewLab Telegram 連線測試成功。此訊息由手動命令發出，沒有進行任何真實交易。")
}
