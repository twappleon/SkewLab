package analytics
import (
 "testing"
 "time"
 "math"
 "github.com/twappleon/SkewLab/internal/paper"
)
func closeAt(t time.Time,pnl float64)paper.Position{return paper.Position{Status:"CLOSED",ExitTime:&t,RealizedPnL:pnl}}
func TestPeriodsLocalDayAndNoInventedFills(t *testing.T){
 now:=time.Date(2026,11,18,12,0,0,0,time.UTC)
 d7:=time.Date(2026,11,12,1,0,0,0,time.UTC)
 d14:=time.Date(2026,11,5,18,0,0,0,time.UTC)
 old:=time.Date(2026,11,1,1,0,0,0,time.UTC)
 future:=now.Add(time.Hour)
 l:=paper.New(paper.DefaultConfig())
 l.Positions=[]paper.Position{closeAt(d7,2),closeAt(now.Add(-time.Minute),-1),closeAt(d14,5),closeAt(old,-7),closeAt(future,1000),{Status:"PENDING_RESOLUTION",RealizedPnL:800}}
 l.TotalClosed=5
 s:=Calculate(l,now)
 if !s.DataComplete||s.RecordedClosures!=5||s.Last7Days.Closed!=2||s.Last7Days.NetPnLUSD!=1||s.Last14Days.Closed!=3||s.Last14Days.NetPnLUSD!=6{t.Fatalf("bad summary: %+v",s)}
 if s.Last7Days.ProfitFactor==nil||*s.Last7Days.ProfitFactor!=2||s.Last7Days.AverageNetUSD==nil||*s.Last7Days.AverageNetUSD!=.5{t.Fatalf("bad PF/avg: %+v",s.Last7Days)}
 if len(s.Daily)!=14||s.Daily[7].Date!="2026-11-12"||s.Daily[7].NetPnLUSD!=2{t.Fatalf("bad days: %+v",s.Daily)}
}
func TestEmptyAndMissingHistory(t *testing.T){
 now:=time.Now()
 l:=paper.New(paper.DefaultConfig())
 s:=Calculate(l,now)
 if s.Last7Days.AverageNetUSD!=nil||s.Last7Days.ProfitFactor!=nil||s.Last7Days.WinRatePct!=nil{t.Fatal("empty samples must be undefined")}
 d:=now.Add(-time.Minute)
 l.Positions=[]paper.Position{closeAt(d,1),closeAt(d,math.NaN())}
 l.TotalClosed=2;s=Calculate(l,now)
 if s.DataComplete||s.SampleLevel!="incomplete_history"||s.RecordedClosures!=1{t.Fatalf("history not flagged: %+v",s)}
 l.TotalClosed=1;s=Calculate(l,now)
 if s.Last7Days.ProfitFactor!=nil{t.Fatalf("no losses must not create infinite PF: %+v",s.Last7Days)}
}
