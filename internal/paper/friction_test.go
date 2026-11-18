package paper
import (
 "testing"
 "time"
 "github.com/twappleon/SkewLab/internal/fees"
 "github.com/twappleon/SkewLab/internal/polymarket"
 "github.com/twappleon/SkewLab/internal/strategy"
)
func TestFlatMarketStillLosesBeyondStopDueToFees(t *testing.T){
 now:=time.Date(2026,11,18,7,0,0,0,time.UTC)
 c:=DefaultConfig()
 px:=.5+c.SlippagePrice
 qty:=c.OrderUSD/fees.BufferedCost(1,px,c.FeeBps,now)
 proceeds:=fees.BufferedProceeds(qty,.5-c.SlippagePrice,c.FeeBps,now)
 costPct:=100*(c.OrderUSD-proceeds)/c.OrderUSD
 if costPct<=5{t.Fatalf("need to demonstrate stop-cost defect, got %.2f%%",costPct)}
 l:=New(c)
 m:=polymarket.Market{Slug:"btc-updown-5m-test",End:now.Add(4*time.Minute),Active:true,UpToken:"up",Up:polymarket.Quote{Bid:.5,Ask:.5,BidDepth:200,AskDepth:200}}
 l.Evaluate(m,strategy.Result{Direction:"UP",Ready:true},now)
 if len(l.Positions)!=0{t.Fatal("cost-driven stop loss churning must not open new positions")}
}
func TestMissingExitDepthRejectsNewEntry(t *testing.T){
 now:=time.Date(2026,11,18,7,0,0,0,time.UTC)
 c:=DefaultConfig();c.StopLossPct=20
 l:=New(c)
 m:=polymarket.Market{Slug:"test",End:now.Add(4*time.Minute),Active:true,UpToken:"up",Up:polymarket.Quote{Bid:.74,Ask:.745,AskDepth:500,BidDepth:0}}
 l.Evaluate(m,strategy.Result{Direction:"UP",Ready:true},now)
 if len(l.Positions)!=0{t.Fatal("missing exit liquidity should reject entry")}
}
