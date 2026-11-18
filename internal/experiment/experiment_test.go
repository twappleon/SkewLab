package experiment
import("encoding/json";"os";"path/filepath";"testing";"time"
"github.com/twappleon/SkewLab/internal/polymarket"
"github.com/twappleon/SkewLab/internal/strategy")
func sample(at time.Time,secs,n int)Sample{
 start:=at.Add(-95*time.Second)
 return Sample{Slug:time.Unix(int64(n),0).Format("btc-market-150405.000000000"),ConditionID:"condition-x",WindowSeconds:secs,
 Policy:SamplingPolicy,CapturedAt:at,MarketEnd:start.Add(time.Duration(secs)*time.Second),UpQuoteAt:at,DownQuoteAt:at,
 UpBid:.52,UpAsk:.54,UpAskDepth:100,DownBid:.44,DownAsk:.46,DownAskDepth:100,MomentumPct:.03,Skew:-.4,BTCSpotSource:"Kraken"}
}
func TestFixedTimeIndependentOfSkewAndRestart(t *testing.T){
 now:=time.Date(2026,11,18,10,0,0,0,time.UTC);dir:=t.TempDir();r:=New(dir)
 s:=sample(now,300,1)
 m:=polymarket.Market{Slug:s.Slug,ConditionID:s.ConditionID,Seconds:s.WindowSeconds,End:s.MarketEnd,Active:true,
 Up:polymarket.Quote{Bid:s.UpBid,Ask:s.UpAsk,AskDepth:100,UpdatedAt:now},
 Down:polymarket.Quote{Bid:s.DownBid,Ask:s.DownAsk,AskDepth:100,UpdatedAt:now}}
 sig:=strategy.Result{Ready:true,Direction:"WAIT",MomentumPct:.03,Skew:-.4}
 btc:=strategy.Tick{Price:95000,Source:"Kraken",Time:now}
 yes,e:=r.Capture(m,sig,btc,now);if !yes||e!=nil{t.Fatalf("unaligned candidate rejected %v %v",yes,e)}
 if yes,_=r.Capture(m,sig,btc,now);yes{t.Fatal("duplicate market")}
 restored:=New(dir);if len(restored.Snapshot())!=1{t.Fatal("restart lost archive")}
 if len(restored.Due(s.MarketEnd,8))!=0{t.Fatal("outcome before expiry")}
 if len(restored.Due(s.MarketEnd.Add(31*time.Second),8))!=1{t.Fatal("due result missing")}
 if e=restored.Resolve(s.Slug,"wrong","UP",s.MarketEnd.Add(time.Minute));e==nil{t.Fatal("wrong condition accepted")}
 if e=restored.Resolve(s.Slug,s.ConditionID,"UP",s.MarketEnd.Add(time.Minute));e!=nil{t.Fatal(e)}
 if e=restored.Resolve(s.Slug,s.ConditionID,"UP",s.MarketEnd.Add(time.Minute));e!=nil{t.Fatal("duplicate resolution")}
 if e=restored.Resolve(s.Slug,s.ConditionID,"DOWN",s.MarketEnd.Add(time.Minute));e==nil{t.Fatal("contradiction accepted")}
 if New(dir).Snapshot()[0].Winner!="UP"{t.Fatal("resolution not persistent")}
}
func TestInvalidAndStaleQuotes(t *testing.T){
 now:=time.Now().UTC();s:=sample(now,300,2);r:=New(t.TempDir())
 m:=polymarket.Market{Slug:s.Slug,ConditionID:s.ConditionID,Seconds:300,End:s.MarketEnd,Active:true,
 Up:polymarket.Quote{Bid:.52,Ask:.54,AskDepth:100,UpdatedAt:now.Add(-15*time.Second)},
 Down:polymarket.Quote{Bid:.44,Ask:.46,AskDepth:100,UpdatedAt:now}}
 sig:=strategy.Result{Ready:true,MomentumPct:.03,Skew:-.3};spot:=strategy.Tick{Time:now,Price:99000,Source:"Kraken"}
 if y,_:=r.Capture(m,sig,spot,now);y{t.Fatal("stale quote accepted")}
 m.Up.UpdatedAt=now;m.Down.UpdatedAt=now.Add(-7*time.Second)
 if y,_:=r.Capture(m,sig,spot,now);y{t.Fatal("out-of-sync books accepted")}
 m.Down.UpdatedAt=now;m.Up.Bid=.56
 if y,_:=r.Capture(m,sig,spot,now);y{t.Fatal("crossed book accepted")}
}
func TestSameUniverseAndForwardNoFuture(t *testing.T){
 base:=time.Date(2026,11,1,0,0,0,0,time.UTC);list:=[]Sample{}
 for i:=0;i<103;i++{
  at:=base.Add(time.Duration(i)*10*time.Minute)
  s:=sample(at,300,i+10);s.Winner="UP";s.ResolvedAt=s.MarketEnd.Add(time.Minute)
  list=append(list,s)
 }
 summ:=Compare(list,base.Add(1200*time.Minute))
 if summ.Resolved!=103||summ.AllWindows[0].ForwardResolved!=3||summ.AllWindows[1].ForwardResolved!=0||summ.TradingAuthorized{t.Fatalf("future leak or trading auto enabled %+v",summ)}
 r:=summ.AllWindows[0].AllRules
 if r[0].Decisions!=103||r[0].Wins!=103||r[1].Decisions!=0||r[2].Decisions!=103||r[2].MarketPriceBrier==nil{t.Fatalf("bad equal universe %+v",r)}
 if r[0].NetPnLPerShare>=46{t.Fatal("cost was omitted")}
 for i:=range list{list[i].ResolvedAt=base.Add(5000*time.Minute)}
 summ=Compare(list,base.Add(5100*time.Minute))
 if summ.AllWindows[0].ForwardResolved!=0{t.Fatal("later outcomes leaked into past observations")}
 summ=Compare(list,base.Add(1200*time.Minute))
 if summ.Resolved!=0||summ.TradingAuthorized{t.Fatal("future labels counted")}
}
func TestBadArchiveAndZeroSample(t *testing.T){
 dir:=t.TempDir();path:=filepath.Join(dir,"strategy-comparison-v1.json")
 if e:=os.WriteFile(path,[]byte("{bad"),0600);e!=nil{t.Fatal(e)}
 if r:=New(dir);!r.Error()||r.Summary(time.Now()).Status!="RESEARCH_ARCHIVE_ERROR"{t.Fatal("corrupt file hidden")}
 x:=sample(time.Now().UTC(),300,1);b,_:=json.Marshal([]Sample{x,x})
 if e:=os.WriteFile(path,b,0600);e!=nil{t.Fatal(e)}
 if !New(dir).Error(){t.Fatal("duplicate archive accepted")}
 sum:=Compare(nil,time.Now())
 if sum.Total!=0||sum.TradingAuthorized||sum.AllWindows[0].AllRules[0].AvgNetPerDecision!=nil{t.Fatal("fabricated empty sample performance")}
}
