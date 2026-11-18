package main
import(
 "context"
 "encoding/json"
 "fmt"
 "log"
 "net/http"
 "sync"
 "time"
 "github.com/twappleon/SkewLab/internal/polymarketscan"
 "github.com/twappleon/SkewLab/internal/polymarketv2"
 "github.com/twappleon/SkewLab/internal/storage"
)
// No trade submission, wallet access or changes to existing paper ledger.
type polySnapshot struct {
 Mode string `json:"mode"`
 AsOf time.Time `json:"as_of"`
 Error string `json:"error,omitempty"`
 BankrollUSD float64 `json:"bankroll_usd"`
 MaxPositionUSD float64 `json:"max_position_usd"`
 MaxExposureUSD float64 `json:"max_exposure_usd"`
 DailyLossUSD float64 `json:"daily_loss_usd"`
 MarketsScanned int `json:"markets_scanned"`
 PaperTrades int `json:"paper_trades"`
 PaperPnLUSD float64 `json:"paper_pnl_usd"`
 Candidates []polyCandidate `json:"candidates"`
 Note string `json:"note"`
}
type polyCandidate struct {
 Question string `json:"question"`
 URL string `json:"url"`
 Quote polymarketv2.Quote `json:"clob_quote"`
 Decision string `json:"decision"`
 Reason string `json:"reason"`
 ConservativeEntryFeeUSD float64 `json:"conservative_entry_fee_usd"`
}
var polyResearchState = newPolyState()

type polyState struct{mu sync.RWMutex;latest polySnapshot}
func newPolyState()*polyState{return &polyState{latest:polySnapshot{Mode:"READ_ONLY_PAPER_RESEARCH",BankrollUSD:100,MaxPositionUSD:5,MaxExposureUSD:30,DailyLossUSD:2,Candidates:[]polyCandidate{},Note:"No independent calibrated probabilities, no fills, no realized returns. Real trading disabled."}}}
func(s *polyState)read()polySnapshot{s.mu.RLock();defer s.mu.RUnlock();return s.latest}
func(s *polyState)write(v polySnapshot){s.mu.Lock();s.latest=v;s.mu.Unlock()}
func polyTokens(raw json.RawMessage)[]string{
 var out []string
 if json.Unmarshal(raw,&out)==nil{return out}
 var wrapped string
 if json.Unmarshal(raw,&wrapped)==nil{_ = json.Unmarshal([]byte(wrapped),&out)}
 return out
}
func(s *polyState)scan(ctx context.Context)polySnapshot{
 snapshot:=newPolyState().read();snapshot.AsOf=time.Now().UTC()
 markets,err:=polymarketscan.Fetch(ctx,nil,30);if err!=nil{snapshot.Error=err.Error();return snapshot}
 report,err:=polymarketscan.Scan(snapshot.AsOf,markets,polymarketscan.DefaultPolicy());if err!=nil{snapshot.Error=err.Error();return snapshot}
 snapshot.MarketsScanned=len(markets)
 for i,c:=range report.Candidates{
  if i>=8{break}
  entry:=polyCandidate{Question:c.Question,URL:c.URL,Decision:"NO_TRADE",Reason:"independent calibrated probability unavailable"}
  for _,m:=range markets{if m.ID!=c.ID{continue}
   tokens:=m.TokenIDs()
   if len(tokens)!=2||tokens[0]==""{entry.Reason="missing or invalid CLOB token IDs";break}
   bookCtx,cancel:=context.WithTimeout(ctx,5*time.Second)
   book,e:=polymarketv2.FetchBook(bookCtx,nil,tokens[0]);cancel()
   if e!=nil{entry.Reason="CLOB unavailable (no synthetic fills)";break}
   entry.Quote=polymarketv2.Evaluate(book,time.Now().UTC(),5)
   if entry.Quote.Executable{entry.ConservativeEntryFeeUSD=polymarketv2.Fee(5,entry.Quote.BuyVWAP,0.07)}
   if !entry.Quote.Executable{entry.Reason=entry.Quote.Reason}else{entry.Reason="visible depth only, no independent probability model evidence"}
   break
  }
  snapshot.Candidates=append(snapshot.Candidates,entry)
 }
 return snapshot
}
func(a *App)startPolyResearch(ctx context.Context,s *polyState){
 if a.demo{return}
 go func(){
  ticker:=time.NewTicker(15*time.Minute);defer ticker.Stop()
  run:=func(){
   c,cancel:=context.WithTimeout(ctx,75*time.Second);defer cancel()
   result:=s.scan(c);s.write(result)
   if result.Error!=""{log.Printf("Polymarket research unavailable: %s",result.Error);return}
   if err:=storage.AppendJSONL(a.dataDir+"/polymarket-research.jsonl",result,25<<20);err!=nil{log.Printf("Polymarket audit persist error: %s",err);return}
   if a.alerts!=nil{
    local:=result.AsOf.In(time.FixedZone("Phnom Penh",7*3600))
    if local.Hour()>=9{
     msg:=fmt.Sprintf("🔎 Polymarket $100 日報（純研究 NO_TRADE）\n掃描市場：%d｜深度候選：%d\n已完成模擬交易：0｜實現模擬損益：$0（未驗證獲利）\n單筆 $5｜總曝險 $30｜日損 $2\n目前沒有獨立概率模型的驗證數據。\nDashboard: https://skewlab-paper-production.up.railway.app/polymarket",result.MarketsScanned,len(result.Candidates))
     a.alerts.Publish("polymarket:daily:"+local.Format("2006-01-02"),msg)
    }
   }
  }
  run()
  for{select{case <-ctx.Done():return;case <-ticker.C:run()}}
 }()
}
func registerPolyRoutes(mux *http.ServeMux,s *polyState){
 mux.HandleFunc("/api/polymarket",func(w http.ResponseWriter,r *http.Request){
  if r.Method!="GET"{http.Error(w,"method not allowed",405);return}
  w.Header().Set("Content-Type","application/json; charset=utf-8");w.Header().Set("Cache-Control","no-store")
  _=json.NewEncoder(w).Encode(s.read())
 })
 mux.HandleFunc("/polymarket",func(w http.ResponseWriter,r *http.Request){
  if r.Method!="GET"{http.Error(w,"method not allowed",405);return}
  w.Header().Set("Content-Type","text/html; charset=utf-8");w.Header().Set("Cache-Control","no-store")
  _,_=w.Write([]byte(polyHTML))
 })
}
const polyHTML=`<!doctype html><html lang="zh-Hant"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>SkewLab Polymarket $100</title><style>body{background:#111823;color:#ecf1fa;font:16px system-ui;max-width:850px;margin:40px auto;padding:0 18px}section{background:#202b39;border-radius:12px;padding:22px;margin:16px 0}a{color:#9bc8ff}small{color:#b4bfcd}.warn{color:#ffbd7f}</style><h1>Polymarket · $100 研究儀表板</h1><p class="warn">NO_TRADE · 模擬研究，不是真實交易或獲利保證</p><section id="summary">載入中…</section><section><h2>監控市場</h2><div id="candidates"></div></section><small>更新間隔 15 分鐘；訂單簿可見深度不等於保證成交。現有 SkewLab 帳本保持隔離。</small><script>
const escapeHTML=s=>String(s??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
async function refresh(){try{let r=await fetch('/api/polymarket',{cache:'no-store'});if(!r.ok)throw Error('HTTP '+r.status);let d=await r.json();document.querySelector('#summary').innerHTML='<b>虛擬本金 $'+d.bankroll_usd+'</b>｜單筆 $'+d.max_position_usd+'｜總曝險 $'+d.max_exposure_usd+'｜日損上限 $'+d.daily_loss_usd+'<p>完成模擬交易：'+d.paper_trades+'｜實現模擬盈虧：$'+d.paper_pnl_usd+'</p><p>掃描市場：'+d.markets_scanned+'｜最後更新：'+escapeHTML(d.as_of||'尚未掃描')+'</p><small>'+escapeHTML(d.note)+(d.error?'｜錯誤：'+escapeHTML(d.error):'')+'</small>';document.querySelector('#candidates').innerHTML=(d.candidates||[]).map(x=>'<p><a target="_blank" rel="noopener noreferrer" href="'+encodeURI(x.url)+'">'+escapeHTML(x.question)+'</a><br><small>'+escapeHTML(x.decision)+' · '+escapeHTML(x.reason)+(x.clob_quote?.executable?' · 可見買入 VWAP $'+Number(x.clob_quote.buy_vwap).toFixed(4):'')+'</small></p>').join('')||'<p>尚無合格的候選市場。</p>'}catch(e){document.querySelector('#summary').textContent='資料讀取失敗：'+e.message}}
refresh();setInterval(refresh,60000)</script></html>`
