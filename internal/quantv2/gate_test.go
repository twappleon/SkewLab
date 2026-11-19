package quantv2
import("testing";"time"
"github.com/twappleon/SkewLab/internal/polymarket"
"github.com/twappleon/SkewLab/internal/strategy"
"github.com/twappleon/SkewLab/internal/paper")
func TestBDDRiskProfileAndWorstCaseDailyReservation(t *testing.T){
 c:=Config()
 if c.StartingCash!=100||c.OrderUSD!=2||c.MaxExposureUSD!=2||c.MaxDailyLossUSD!=2{t.Fatalf("wrong capital limits: %+v",c)}
 l:=paper.New(c);now:=time.Date(2026,11,19,3,0,0,0,time.UTC)
 if !l.CanReserveNewStake(now){t.Fatal("first stake should be reservable")}
 l.Positions=append(l.Positions,paper.Position{EntryCost:2,Status:"OPEN",EntryTime:now})
 if l.CanReserveNewStake(now){t.Fatal("outstanding 2 USDC worst-case exposure must prevent any additional entry")}
}
func TestBDDStaleBookAndWideSpreadNeverTrade(t *testing.T){
 now:=time.Date(2026,11,19,3,0,0,0,time.UTC)
 c:=Config()
 s:=strategy.Result{Direction:"UP",Ready:true,MomentumPct:.03,Skew:.5}
 m:=polymarket.Market{Active:true,End:now.Add(3*time.Minute),Up:polymarket.Quote{Bid:.55,Ask:.56,BidDepth:100,AskDepth:100,UpdatedAt:now}}
 if r:=Gate(s,m,c,now);r.Direction!="WAIT"{t.Fatal("wide spread/fees should not pass")}
 m.Up.UpdatedAt=now.Add(-6*time.Second)
 if r:=Gate(s,m,c,now);r.Direction!="WAIT"{t.Fatal("stale quote should not pass")}
}
func TestBDDNoSignalAndInvalidRiskConfiguration(t *testing.T){
 now:=time.Date(2026,11,19,3,0,0,0,time.UTC)
 m:=polymarket.Market{Active:true,End:now.Add(3*time.Minute),Up:polymarket.Quote{Bid:.5,Ask:.5,BidDepth:100,AskDepth:100,UpdatedAt:now}}
 s:=strategy.Result{Direction:"WAIT",Ready:true}
 if Gate(s,m,Config(),now).Direction!="WAIT"{t.Fatal("WAIT must remain WAIT")}
 s.Direction="UP";c:=Config();c.OrderUSD=20
 if Gate(s,m,c,now).Direction!="WAIT"{t.Fatal("wrong stake must fail closed")}
}
