package research100
import("testing";"time")
func TestRound083BaselineTieMustAbstain(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=pairedFixture(now,"a")
 x:=MarketBaselineSide(a,now)
 if x.Side!="WAIT"||x.Status!="TIED_MARKET_PROXY_ABSTAIN"{t.Fatalf("forced direction %+v",x)}
 a.DownBid=.30;a.DownAsk=.32
 x=MarketBaselineSide(a,now)
 if x.Side!="UP"{t.Fatalf("wrong baseline %+v",x)}
}
