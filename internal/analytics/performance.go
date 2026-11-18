// Package analytics calculates read-only paper performance from completed
// positions; it never changes the underlying paper ledger.
package analytics

import (
 "math"
 "time"
 "github.com/twappleon/SkewLab/internal/paper"
)
type Day struct {
 Date string `json:"date"`
 Closed int `json:"closed"`
 Wins int `json:"wins"`
 NetPnLUSD float64 `json:"net_pnl_usd"`
}
type Period struct {
 Days int `json:"days"`
 Closed int `json:"closed"`
 Wins int `json:"wins"`
 NetPnLUSD float64 `json:"net_pnl_usd"`
 GrossProfitUSD float64 `json:"gross_profit_usd"`
 GrossLossUSD float64 `json:"gross_loss_usd"`
 AverageNetUSD *float64 `json:"average_net_usd"`
 ProfitFactor *float64 `json:"profit_factor"`
 WinRatePct *float64 `json:"win_rate_pct"`
}
type Summary struct {
 Last7Days Period `json:"last_7_days"`
 Last14Days Period `json:"last_14_days"`
 Daily []Day `json:"daily"`
 HistoricalClosures int `json:"historical_closures"`
 RecordedClosures int `json:"recorded_closures"`
 DataComplete bool `json:"data_complete"`
 SampleLevel string `json:"sample_level"`
}
var phnomPenh=time.FixedZone("Asia/Phnom_Penh",7*60*60)
func round(n float64)float64{return math.Round(n*10000)/10000}
func ratio(n,d float64)*float64{if d<=0{return nil};v:=round(n/d);return &v}
func add(p *Period,n float64) {
 p.Closed++;p.NetPnLUSD+=n
 if n>0{p.Wins++;p.GrossProfitUSD+=n}else if n<0{p.GrossLossUSD-=n}
}
func finish(p *Period) {
 p.NetPnLUSD=round(p.NetPnLUSD);p.GrossProfitUSD=round(p.GrossProfitUSD);p.GrossLossUSD=round(p.GrossLossUSD)
 p.AverageNetUSD=ratio(p.NetPnLUSD,float64(p.Closed))
 p.WinRatePct=ratio(float64(p.Wins)*100,float64(p.Closed))
 p.ProfitFactor=ratio(p.GrossProfitUSD,p.GrossLossUSD)
}
// Calculate uses Asia/Phnom_Penh calendar days; open/unresolved entries do
// not count as realized trades. Future exit times are not counted in periods.
func Calculate(l *paper.Ledger,now time.Time) Summary {
 out:=Summary{Last7Days:Period{Days:7},Last14Days:Period{Days:14}}
 if l==nil{out.SampleLevel="no_data";return out}
 localNow:=now.In(phnomPenh)
 today:=time.Date(localNow.Year(),localNow.Month(),localNow.Day(),0,0,0,0,phnomPenh)
 days:=make([]Day,14);index:=make(map[string]int,14)
 for i:=range days{date:=today.AddDate(0,0,i-13).Format("2006-01-02");days[i]=Day{Date:date};index[date]=i}
 for _,pos:=range l.Positions {
  if (pos.Status!="CLOSED"&&pos.Status!="SETTLED")||pos.ExitTime==nil{continue}
  if math.IsNaN(pos.RealizedPnL)||math.IsInf(pos.RealizedPnL,0){continue}
  out.RecordedClosures++
  local:=pos.ExitTime.In(phnomPenh)
  if local.After(localNow){continue}
  idx,ok:=index[local.Format("2006-01-02")];if !ok{continue}
  d:=&days[idx];d.Closed++;if pos.RealizedPnL>0{d.Wins++};d.NetPnLUSD+=pos.RealizedPnL
  add(&out.Last14Days,pos.RealizedPnL)
  if idx>=7{add(&out.Last7Days,pos.RealizedPnL)}
 }
 out.HistoricalClosures=l.TotalClosed
 out.DataComplete=out.HistoricalClosures==out.RecordedClosures
 for i:=range days{days[i].NetPnLUSD=round(days[i].NetPnLUSD)}
 out.Daily=days;finish(&out.Last7Days);finish(&out.Last14Days)
 switch {
 case !out.DataComplete:out.SampleLevel="incomplete_history"
 case out.Last14Days.Closed<30:out.SampleLevel="insufficient"
 case out.Last14Days.Closed<100:out.SampleLevel="preliminary"
 default:out.SampleLevel="research_sample"
 }
 return out
}
