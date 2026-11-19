// Package quantv2 implements an opt-in, paper-only, cost-aware signal gate.
// It deliberately contains no exchange execution, signer or credential handling.
package quantv2

import (
 "math"
 "time"
 "github.com/twappleon/SkewLab/internal/fees"
 "github.com/twappleon/SkewLab/internal/paper"
 "github.com/twappleon/SkewLab/internal/polymarket"
 "github.com/twappleon/SkewLab/internal/strategy"
)

const Version = "QuantSkew V2 paper-preview"
func Config() paper.Config {
 c:=paper.DefaultConfig()
 c.StartingCash=100
 c.OrderUSD=2
 c.MaxExposureUSD=2
 c.MaxDailyLossUSD=2
 c.MaxDailyEntries=100
 return c
}
func valid(x float64) bool{return !math.IsNaN(x)&&!math.IsInf(x,0)}
// Gate only selects hypothetical paper opportunities when the currently
// executable book can absorb a complete round-trip with modeled friction.
// This does not establish a profitable predictive edge or guaranteed fills.
func Gate(s strategy.Result,m polymarket.Market,c paper.Config,now time.Time) strategy.Result {
 refuse:=func(reason string)strategy.Result{s.Direction="WAIT";s.Reason=reason;return s}
 if !s.Ready||(s.Direction!="UP"&&s.Direction!="DOWN"){return refuse("V2: no aligned BTC signal")}
 if !m.Active||!now.Before(m.End.Add(-65*time.Second)){return refuse("V2: inactive/expiring market")}
 if c.StartingCash!=100||c.OrderUSD!=2||c.MaxExposureUSD!=2||c.MaxDailyLossUSD!=2 {return refuse("V2: risk profile mismatch")}
 if !valid(c.FeeBps)||c.FeeBps<0||!valid(c.SlippagePrice)||c.SlippagePrice<0||!valid(c.StopLossPct)||c.StopLossPct<=0{return refuse("V2: invalid cost or stop assumptions")}
 q:=m.Up
 if s.Direction=="DOWN"{q=m.Down}
 if q.UpdatedAt.IsZero()||q.UpdatedAt.After(now.Add(time.Second))||now.Sub(q.UpdatedAt)>5*time.Second{return refuse("V2: stale CLOB quote")}
 if !valid(q.Ask)||!valid(q.Bid)||!valid(q.AskDepth)||!valid(q.BidDepth)||q.Bid<=0||q.Ask<=0||q.Ask>=1||q.Bid>q.Ask{return refuse("V2: invalid/crossed CLOB")}
 buy:=q.Ask+c.SlippagePrice
 sell:=q.Bid-c.SlippagePrice
 if buy>=1||sell<=0||buy>c.MaxBuyPrice{return refuse("V2: non-executable prices")}
 unit:=fees.BufferedCost(1,buy,c.FeeBps,now)
 if !valid(unit)||unit<=0{return refuse("V2: invalid fee estimate")}
 qty:=c.OrderUSD/unit
 if q.AskDepth<math.Max(qty,c.MinDepthShares)||q.BidDepth<qty{return refuse("V2: insufficient book depth")}
 exit:=fees.BufferedProceeds(qty,sell,c.FeeBps,now)
 friction:=100*(c.OrderUSD-exit)/c.OrderUSD
 if !valid(friction)||friction<0||friction>=c.StopLossPct{return refuse("V2: spread/fees/slippage exceed stop")}
 return s
}
