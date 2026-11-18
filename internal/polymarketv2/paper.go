package polymarketv2

import (
 "errors"
 "fmt"
 "math"
 "time"
)
// PaperPosition is a hypothetical hold-to-resolution position. It is never
// submitted to Polymarket and must not be presented as an actual fill.
type PaperPosition struct {
 ID string `json:"id"`
 EventID string `json:"event_id"`
 OpenedAt time.Time `json:"opened_at"`
 EntryPrice float64 `json:"entry_price"`
 Shares float64 `json:"shares"`
 EntryCostUSD float64 `json:"entry_cost_usd"`
 Status string `json:"status"`
 ResolvedAt time.Time `json:"resolved_at,omitempty"`
 Won bool `json:"won"`
 RealizedPnLUSD float64 `json:"realized_pnl_usd"`
}
type PaperAccount struct {
 StartingUSD float64 `json:"starting_usd"`
 CashUSD float64 `json:"cash_usd"`
 Positions []PaperPosition `json:"positions"`
}
func NewPaperAccount()PaperAccount{return PaperAccount{StartingUSD:100,CashUSD:100,Positions:[]PaperPosition{}}}
func(a *PaperAccount) ExposureUSD()float64{v:=0.0;for _,p:=range a.Positions{if p.Status=="OPEN"{v+=p.EntryCostUSD}};return v}
func(a *PaperAccount) RealizedPnLUSD()float64{v:=0.0;for _,p:=range a.Positions{if p.Status=="RESOLVED"{v+=p.RealizedPnLUSD}};return v}
func(a *PaperAccount) Open(now time.Time,id,event string,q Quote,assessment Assessment)error{
 if a.StartingUSD!=100||a.CashUSD<0||a.CashUSD>10000 {return errors.New("invalid account state")}
 if id==""||event==""||!q.Executable||assessment.Decision!="PAPER_CANDIDATE_ONLY"{return errors.New("no valid paper candidate")}
 if assessment.Shares<=0||!finite(assessment.Shares)||assessment.ExpectedCostsUSD<0||!finite(assessment.ExpectedCostsUSD){return errors.New("invalid stake")}
 cost:=assessment.Shares*q.BuyVWAP+assessment.ExpectedCostsUSD
 if cost>5.000001||cost<=0||cost>a.CashUSD||cost+a.ExposureUSD()>30 {return errors.New("position, cash or exposure limit")}
 for _,p:=range a.Positions {if p.ID==id{return errors.New("duplicate id")};if p.Status=="OPEN"&&p.EventID==event{return errors.New("correlated event exposure")}}
 // Realized losses from this UTC day plus conservative open exposure loss.
 todayLoss:=0.0
 for _,p:=range a.Positions {if p.Status=="RESOLVED"&&p.ResolvedAt.UTC().Format("2006-01-02")==now.UTC().Format("2006-01-02")&&p.RealizedPnLUSD<0{todayLoss-=p.RealizedPnLUSD}}
 if todayLoss>=2{return errors.New("daily loss cap reached")}
 a.CashUSD-=cost
 a.Positions=append(a.Positions,PaperPosition{ID:id,EventID:event,OpenedAt:now.UTC(),EntryPrice:q.BuyVWAP,Shares:assessment.Shares,EntryCostUSD:cost,Status:"OPEN"})
 return nil
}
// Resolve requires an externally verified binary winner, and is idempotent.
// Missing, disputed, or pending outcomes must not call this function.
func(a *PaperAccount) Resolve(id string,won bool,resolvedAt time.Time)error{
 if resolvedAt.IsZero(){return errors.New("unverified settlement time")}
 for i:=range a.Positions {p:=&a.Positions[i]
  if p.ID!=id{continue}
  if p.Status=="RESOLVED"{return nil}
  if p.Status!="OPEN"||resolvedAt.Before(p.OpenedAt){return errors.New("invalid settlement order")}
  payout:=0.0;if won{payout=p.Shares}
  p.RealizedPnLUSD=payout-p.EntryCostUSD;p.Status="RESOLVED";p.Won=won;p.ResolvedAt=resolvedAt.UTC()
  a.CashUSD+=payout
  if !finite(a.CashUSD)||math.IsNaN(p.RealizedPnLUSD){return errors.New("invalid accounting")}
  return nil
 }
 return fmt.Errorf("unknown paper position %s",id)
}
