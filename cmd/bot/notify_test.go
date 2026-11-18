package main
import(
 "testing"
 "time"
 "github.com/twappleon/SkewLab/internal/notify"
 "github.com/twappleon/SkewLab/internal/paper"
)
func TestPaperActionNotificationAndDedup(t *testing.T){
 now:=time.Date(2026,11,18,5,0,0,0,time.UTC)
 a:=NewApp(t.TempDir(),false,0.02,0.2,3*time.Second)
 a.alerts=notify.New("fake","1234",t.TempDir())
 a.ledger.Positions=[]paper.Position{{ID:"entry-1",Market:"btc-updown-5m-100",Side:"UP",Status:"OPEN",EntryTime:now,EntryCost:20,EntryPrice:0.50}}
 a.alertPaperAction(now,"PAPER BUY UP btc-updown-5m-100")
 a.alertPaperAction(now,"PAPER BUY UP btc-updown-5m-100")
 if got:=a.alerts.Status().Queued;got!=1{t.Fatalf("expected one entry notification: %d",got)}
 a.ledger.Positions[0].Status="CLOSED";a.ledger.Positions[0].ExitTime=&now
 a.ledger.Positions[0].ExitPrice=0.55;a.ledger.Positions[0].RealizedPnL=1.24;a.ledger.Positions[0].Note="net take profit"
 a.alertPaperAction(now,"PAPER SELL UP btc-updown-5m-100")
 if got:=a.alerts.Status().Queued;got!=2{t.Fatalf("expected buy+take profit: %d",got)}
}
func TestDailyAndOutageNotificationRateLimit(t *testing.T) {
 now:=time.Date(2026,11,18,3,0,0,0,time.UTC)
 a:=NewApp(t.TempDir(),false,0.02,0.2,3*time.Second)
 a.alerts=notify.New("fake","1234",t.TempDir())
 a.alertDailySummary(now);a.alertDailySummary(now.Add(3*time.Second))
 if got:=a.alerts.Status().Queued;got!=1{t.Fatalf("daily duplicates: %d",got)}
 a.alertHealth(now,[]string{"BTC unavailable"})
 a.alertHealth(now.Add(65*time.Second),[]string{"BTC unavailable"})
 a.alertHealth(now.Add(70*time.Second),[]string{"BTC unavailable"})
 a.alertHealth(now.Add(75*time.Second),nil)
 if got:=a.alerts.Status().Queued;got!=3{t.Fatalf("expected daily+outage+recovery: %d",got)}
 a.ledger.Positions=[]paper.Position{{ID:"loss-1",Status:"CLOSED",RealizedPnL:-31,ExitTime:&now}}
 a.alertRisk(now);a.alertRisk(now)
 if got:=a.alerts.Status().Queued;got!=4{t.Fatalf("expected one loss-limit alert: %d",got)}
}
