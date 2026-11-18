// Fixed-time, read-only strategy comparison. No paper or real orders.
package experiment
import("encoding/json";"errors";"fmt";"math";"os";"path/filepath";"sort";"sync";"time"
"github.com/twappleon/SkewLab/internal/fees"
"github.com/twappleon/SkewLab/internal/polymarket"
"github.com/twappleon/SkewLab/internal/storage"
"github.com/twappleon/SkewLab/internal/strategy")
const SamplingPolicy="fixed-90-to-120s-after-market-open-v1"
const MaxSamples=50000
const PriorOutcomesRequired=100
type Sample struct {
 Slug string `json:"slug"`
 ConditionID string `json:"condition_id"`
 WindowSeconds int `json:"window_seconds"`
 Policy string `json:"policy"`
 CapturedAt time.Time `json:"captured_at"`
 MarketEnd time.Time `json:"market_end"`
 UpQuoteAt time.Time `json:"up_quote_at"`
 DownQuoteAt time.Time `json:"down_quote_at"`
 UpBid float64 `json:"up_bid"`
 UpAsk float64 `json:"up_ask"`
 UpAskDepth float64 `json:"up_ask_depth"`
 DownBid float64 `json:"down_bid"`
 DownAsk float64 `json:"down_ask"`
 DownAskDepth float64 `json:"down_ask_depth"`
 MomentumPct float64 `json:"momentum_pct"`
 Skew float64 `json:"skew"`
 BTCSpotSource string `json:"btc_spot_source"`
 Winner string `json:"winner,omitempty"`
 ResolvedAt time.Time `json:"resolved_at,omitempty"`
 ResolutionSource string `json:"resolution_source,omitempty"`
}
type Recorder struct{mu sync.RWMutex;path string;items []Sample;index map[string]int;err error}
func New(dir string)*Recorder{
 r:=&Recorder{path:filepath.Join(dir,"strategy-comparison-v1.json"),index:map[string]int{}}
 b,e:=os.ReadFile(r.path)
 if os.IsNotExist(e){return r};if e!=nil{r.err=fmt.Errorf("research file unreadable: %w",e);return r}
 if e=json.Unmarshal(b,&r.items);e!=nil{r.err=fmt.Errorf("research file invalid: %w",e);r.items=nil;return r}
 if len(r.items)>MaxSamples{r.err=errors.New("research file exceeds capacity");return r}
 for i,s:=range r.items{
  if s.Policy!=SamplingPolicy||s.Slug==""||s.ConditionID==""||!validSnapshot(s){r.err=errors.New("invalid comparison sample");return r}
  if _,ok:=r.index[s.Slug];ok{r.err=errors.New("duplicate market in research file");return r}
  r.index[s.Slug]=i
 }
 return r
}
func finite(v float64)bool{return !math.IsNaN(v)&&!math.IsInf(v,0)}
func quoteOK(q polymarket.Quote,now time.Time)bool{
 if !finite(q.Ask)||!finite(q.Bid)||!finite(q.AskDepth)||q.Ask<=0||q.Ask>=1||q.Bid<=0||q.Bid>q.Ask||q.AskDepth<=0{return false}
 if q.UpdatedAt.IsZero()||q.UpdatedAt.After(now.Add(3*time.Second))||now.Sub(q.UpdatedAt)>10*time.Second{return false}
 return true
}
func validSnapshot(s Sample)bool{
 if s.WindowSeconds!=300&&s.WindowSeconds!=900||s.CapturedAt.IsZero()||s.MarketEnd.IsZero(){return false}
 elapsed:=s.CapturedAt.Sub(s.MarketEnd.Add(-time.Duration(s.WindowSeconds)*time.Second))
 if elapsed<90*time.Second||elapsed>120*time.Second||!s.CapturedAt.Before(s.MarketEnd){return false}
 if s.UpQuoteAt.IsZero()||s.DownQuoteAt.IsZero()||s.CapturedAt.Sub(s.UpQuoteAt)>10*time.Second||s.CapturedAt.Sub(s.DownQuoteAt)>10*time.Second||s.UpQuoteAt.After(s.CapturedAt.Add(3*time.Second))||s.DownQuoteAt.After(s.CapturedAt.Add(3*time.Second)){return false}
 if !finite(s.MomentumPct)||!finite(s.Skew)||s.Skew < -1.000001||s.Skew>1.000001{return false}
 if !finite(s.UpAsk)||!finite(s.DownAsk)||!finite(s.UpBid)||!finite(s.DownBid)||!finite(s.UpAskDepth)||!finite(s.DownAskDepth){return false}
 if s.UpAsk<=0||s.UpAsk>=1||s.DownAsk<=0||s.DownAsk>=1||s.UpBid<=0||s.DownBid<=0||s.UpBid>s.UpAsk||s.DownBid>s.DownAsk||s.UpAskDepth<=0||s.DownAskDepth<=0{return false}
 if s.Winner!=""&&(s.Winner!="UP"&&s.Winner!="DOWN"||s.ResolvedAt.Before(s.MarketEnd)){return false}
 if s.Winner==""&&!s.ResolvedAt.IsZero(){return false}
 return true
}
func(r *Recorder) save()error{b,e:=json.Marshal(r.items);if e!=nil{return e};return storage.Atomic(r.path,b)}
// Capture chooses the first valid market observation within 90-120s after open.
// Both prices and the raw BTC features are saved, even when signals disagree.
// No backfill, trades or inferred outcomes.
func(r *Recorder) Capture(m polymarket.Market,s strategy.Result,spot strategy.Tick,now time.Time)(bool,error){
 if r==nil{return false,nil}
 r.mu.Lock();defer r.mu.Unlock()
 if r.err!=nil{return false,r.err}
 if (m.Seconds!=300&&m.Seconds!=900)||m.Slug==""||m.ConditionID==""||!m.Active||!s.Ready||!finite(s.MomentumPct)||!finite(s.Skew)||!finite(spot.Price)||spot.Price<=0||spot.Source==""||spot.Time.IsZero(){return false,nil}
 if _,ok:=r.index[m.Slug];ok{return false,nil}
 elapsed:=now.Sub(m.End.Add(-time.Duration(m.Seconds)*time.Second))
 if elapsed<90*time.Second||elapsed>120*time.Second||!now.Before(m.End){return false,nil}
 if now.Sub(spot.Time)>5*time.Second||spot.Time.After(now.Add(time.Second))||!quoteOK(m.Up,now)||!quoteOK(m.Down,now){return false,nil}
 delta:=m.Up.UpdatedAt.Sub(m.Down.UpdatedAt);if delta>5*time.Second||delta< -5*time.Second{return false,nil}
 if len(r.items)>=MaxSamples{r.err=errors.New("comparison archive full; archival required");return false,r.err}
 x:=Sample{Slug:m.Slug,ConditionID:m.ConditionID,WindowSeconds:m.Seconds,Policy:SamplingPolicy,
 CapturedAt:now,MarketEnd:m.End,UpQuoteAt:m.Up.UpdatedAt,DownQuoteAt:m.Down.UpdatedAt,
 UpBid:m.Up.Bid,UpAsk:m.Up.Ask,UpAskDepth:m.Up.AskDepth,
 DownBid:m.Down.Bid,DownAsk:m.Down.Ask,DownAskDepth:m.Down.AskDepth,
 MomentumPct:s.MomentumPct,Skew:s.Skew,BTCSpotSource:spot.Source}
 if !validSnapshot(x){return false,nil}
 r.items=append(r.items,x)
 if e:=r.save();e!=nil{r.items=r.items[:len(r.items)-1];r.err=fmt.Errorf("comparison journal write failed: %w",e);return false,r.err}
 r.index[x.Slug]=len(r.items)-1
 return true,nil
}
func(r *Recorder) Due(now time.Time,limit int)[]Sample{
 if r==nil{return nil}
 r.mu.RLock();defer r.mu.RUnlock()
 if r.err!=nil{return nil}
 if limit<=0||limit>32{limit=8}
 out:=make([]Sample,0,limit)
 for _,x:=range r.items{if x.Winner==""&&!now.Before(x.MarketEnd.Add(30*time.Second)){out=append(out,x);if len(out)>=limit{break}}}
 return out
}
func(r *Recorder) Resolve(slug,id,winner string,now time.Time)error{
 if r==nil{return nil}
 r.mu.Lock();defer r.mu.Unlock()
 if r.err!=nil{return r.err}
 i,ok:=r.index[slug];if !ok{return errors.New("sample not found")}
 x:=&r.items[i]
 if id!=x.ConditionID||(winner!="UP"&&winner!="DOWN")||now.Before(x.MarketEnd){return errors.New("inconsistent market resolution")}
 if x.Winner!=""{if x.Winner!=winner{return errors.New("contradictory published result")};return nil}
 x.Winner=winner;x.ResolvedAt=now;x.ResolutionSource="Gamma published outcome; not independently verified Chainlink TWAP"
 if e:=r.save();e!=nil{x.Winner="";x.ResolvedAt=time.Time{};x.ResolutionSource="";r.err=fmt.Errorf("resolution write failed: %w",e);return r.err}
 return nil
}
func(r *Recorder) Snapshot()[]Sample{if r==nil{return nil};r.mu.RLock();defer r.mu.RUnlock();return append([]Sample{},r.items...)}
func(r *Recorder) Error()bool{if r==nil{return true};r.mu.RLock();defer r.mu.RUnlock();return r.err!=nil}
type RuleStats struct{
 Name string `json:"name"`
 Observed int `json:"observed"`
 Decisions int `json:"decisions"`
 Wins int `json:"wins"`
 DirectionAccuracyPct *float64 `json:"direction_accuracy_pct"`
 NetPnLPerShare float64 `json:"net_pnl_per_share_usd"`
 AvgNetPerDecision *float64 `json:"avg_net_per_decision_usd"`
 MarketPriceBrier *float64 `json:"market_price_brier,omitempty"`
}
type WindowStats struct{
 Window string `json:"window"`
 Observed int `json:"observed"`
 Resolved int `json:"resolved"`
 Pending int `json:"pending"`
 ForwardResolved int `json:"forward_resolved"`
 AllRules []RuleStats `json:"all_rules"`
 ForwardRules []RuleStats `json:"forward_rules"`
 MissingReason string `json:"missing_reason"`
}
type Summary struct{
 Policy string `json:"sampling_policy"`
 Total int `json:"total"`
 Resolved int `json:"resolved"`
 Pending int `json:"pending"`
 LatestCapture *time.Time `json:"latest_capture,omitempty"`
 PriorOutcomesRequired int `json:"prior_outcomes_required"`
 AllWindows []WindowStats `json:"windows"`
 TradingAuthorized bool `json:"trading_authorized"`
 Status string `json:"status"`
 Limitations string `json:"limitations"`
}
func emptyRules()[]RuleStats{return []RuleStats{{Name:"momentum_only"},{Name:"momentum_semivariance"},{Name:"market_price_baseline"}}}
func ratio(n float64,d int)*float64{if d<=0{return nil};v:=math.Round(n/float64(d)*10000)/10000;return &v}
func direction(s Sample,which int)string{
 if which==2{
  up:=(s.UpBid+s.UpAsk)/2;down:=(s.DownBid+s.DownAsk)/2
  if math.Abs(up-down)<1e-9{return ""}
  if up>down{return "UP"};return "DOWN"
 }
 side:=""
 if s.MomentumPct>=.02{side="UP"}else if s.MomentumPct<=-.02{side="DOWN"}
 if which==1&&((side=="UP"&&s.Skew<.20)||(side=="DOWN"&&s.Skew>-.20)){return ""}
 return side
}
func addScore(r *RuleStats,s Sample,which int){
 r.Observed++
 if which==2{
  midUp:=(s.UpBid+s.UpAsk)/2;midDown:=(s.DownBid+s.DownAsk)/2
  prob:=midUp/(midUp+midDown);target:=0.0;if s.Winner=="UP"{target=1}
  brier:=(prob-target)*(prob-target)
  if r.MarketPriceBrier==nil{zero:=0.0;r.MarketPriceBrier=&zero}
  *r.MarketPriceBrier+=brier
 }
 side:=direction(s,which);if side==""{return}
 ask:=s.UpAsk;if side=="DOWN"{ask=s.DownAsk}
 px:=ask+.002;if px<=0||px>=1{return}
 cost:=fees.BufferedCost(1,px,25,s.CapturedAt)
 if !finite(cost)||cost<=0{return}
 r.Decisions++
 payoff:=0.0
 if side==s.Winner{r.Wins++;payoff=1}
 r.NetPnLPerShare+=payoff-cost
}
func finishRules(rs []RuleStats){
 for i:=range rs{
  r:=&rs[i]
  r.NetPnLPerShare=math.Round(r.NetPnLPerShare*10000)/10000
  r.DirectionAccuracyPct=ratio(float64(r.Wins)*100,r.Decisions)
  r.AvgNetPerDecision=ratio(r.NetPnLPerShare,r.Decisions)
  if r.MarketPriceBrier!=nil{v:=math.Round(*r.MarketPriceBrier/float64(r.Observed)*10000)/10000;r.MarketPriceBrier=&v}
 }
}
// Compare scores the SAME prospective observation pool, with abstentions.
// Forward eligibility requires >=100 same-window published results STRICTLY
// before the target observation, never afterward. No trading authorization.
func Compare(samples []Sample,now time.Time)Summary{
 out:=Summary{Policy:SamplingPolicy,PriorOutcomesRequired:PriorOutcomesRequired,Status:"INSUFFICIENT_FORWARD_EVIDENCE",TradingAuthorized:false,
 Limitations:"Research only; missing fixed-window observations are not backfilled. Official Gamma labels are not independent Chainlink TWAP proof. Each rule hypothetically buys one share at ask plus modeled costs and holds to settlement; this is NOT the old early-exit strategy or real execution. Market observations are correlated and not a full market census."}
 out.AllWindows=[]WindowStats{{Window:"5m",AllRules:emptyRules(),ForwardRules:emptyRules()},{Window:"15m",AllRules:emptyRules(),ForwardRules:emptyRules()}}
 items:=append([]Sample{},samples...)
 sort.Slice(items,func(i,j int)bool{return items[i].CapturedAt.Before(items[j].CapturedAt)})
 resolved:=[]Sample{}
 for _,s:=range items{if s.Policy==SamplingPolicy&&validSnapshot(s)&&s.Winner!=""&&!s.ResolvedAt.IsZero()&&!s.ResolvedAt.After(now)&&!s.CapturedAt.After(now){resolved=append(resolved,s)}}
 sort.Slice(resolved,func(i,j int)bool{return resolved[i].ResolvedAt.Before(resolved[j].ResolvedAt)})
 past:=[2]int{};ix:=0
 for _,s:=range items{
  if s.Policy!=SamplingPolicy||!validSnapshot(s)||s.CapturedAt.After(now){continue}
  w:=-1;if s.WindowSeconds==300{w=0}else if s.WindowSeconds==900{w=1};if w<0{continue}
  for ix<len(resolved)&&resolved[ix].ResolvedAt.Before(s.CapturedAt){if resolved[ix].WindowSeconds==300{past[0]++}else if resolved[ix].WindowSeconds==900{past[1]++};ix++}
  out.Total++;out.AllWindows[w].Observed++
  if out.LatestCapture==nil||s.CapturedAt.After(*out.LatestCapture){tm:=s.CapturedAt;out.LatestCapture=&tm}
  if s.Winner==""||s.ResolvedAt.After(now){out.Pending++;out.AllWindows[w].Pending++;continue}
  out.Resolved++;out.AllWindows[w].Resolved++
  for k:=0;k<3;k++{addScore(&out.AllWindows[w].AllRules[k],s,k)}
  if past[w]>=PriorOutcomesRequired{
   out.AllWindows[w].ForwardResolved++
   for k:=0;k<3;k++{addScore(&out.AllWindows[w].ForwardRules[k],s,k)}
  }
 }
 for i:=range out.AllWindows{
  w:=&out.AllWindows[i];finishRules(w.AllRules);finishRules(w.ForwardRules)
  if w.ForwardResolved<30{w.MissingReason="Requires 100 already resolved same-window markets plus at least 30 subsequent settled observations."}
 }
 if out.AllWindows[0].ForwardResolved>=30&&out.AllWindows[1].ForwardResolved>=30{out.Status="RESEARCH_FORWARD_SAMPLE_AVAILABLE_NOT_TRADING_APPROVAL"}
 if out.Total==0{out.Status="NO_FIXED_TIME_SAMPLES_YET"}
 return out
}
func(r *Recorder) Summary(now time.Time)Summary{
 out:=Compare(r.Snapshot(),now)
 if r.Error(){out.Status="RESEARCH_ARCHIVE_ERROR";out.Limitations+=" Research archive unreadable; cannot overwrite damaged data."}
 return out
}
