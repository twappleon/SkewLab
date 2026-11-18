// Research-only, cost-aware, time-causal shadow strategy comparison.
// No positions, orders, wallet keys or live trading.
package experiment

import (
 "math"
 "sort"
 "time"
 "github.com/twappleon/SkewLab/internal/fees"
)
const (
 ShadowPriorPerCohort=100
 ShadowForwardPerWindow=30
 ShadowMinimumNetROI=0.02
 ShadowMaximumSpread=0.04
 ShadowSlippage=0.002
 ShadowExtraFeeBps=25.0
)
type ShadowRule struct {
 Name string `json:"name"`
 ForwardObservations int `json:"forward_observations"`
 EligibleCandidates int `json:"eligible_candidates"`
 ResolvedCandidates int `json:"resolved_candidates"`
 Winners int `json:"winners"`
 ShadowNetPerShareUSD float64 `json:"shadow_net_per_share_usd"`
 BaselineNetPerShareUSD float64 `json:"baseline_net_per_share_usd"`
 ExcessNetPerShareUSD float64 `json:"excess_net_per_share_usd"`
 MeanExcessUSD *float64 `json:"mean_excess_usd"`
 BlockedInsufficientHistory int `json:"blocked_insufficient_history"`
 BlockedNoConservativeEdge int `json:"blocked_no_conservative_edge"`
 BlockedSpreadOrDepth int `json:"blocked_spread_or_depth"`
 Abstained int `json:"abstained"`
}
type ShadowWindow struct {
 Window string `json:"window"`
 Rules []ShadowRule `json:"rules"`
 ResolvedMarketSamples int `json:"resolved_market_samples"`
 MaxPriorCohort int `json:"max_prior_cohort"`
 Status string `json:"status"`
}
type ShadowReport struct {
 Status string `json:"status"`
 AsOf time.Time `json:"as_of"`
 RequiredPriorPerCohort int `json:"required_prior_per_cohort"`
 RequiredForwardPerWindow int `json:"required_forward_per_window"`
 MinimumNetROIPct float64 `json:"minimum_net_roi_pct"`
 TradingAuthorized bool `json:"trading_authorized"`
 Windows []ShadowWindow `json:"windows"`
 Limitations string `json:"limitations"`
}
type cohort struct {window int;side string;bucket int}
type winCount struct{total,wins int}
func shadowBucket(p float64)int{v:=int(math.Floor(p*5));if v<0{return 0};if v>4{return 4};return v}
func shadowWilson(wins,n int)float64{
 if n<=0{return 0}
 p,nn,z:=float64(wins)/float64(n),float64(n),1.96
 return math.Max(0,(p+z*z/(2*nn)-z*math.Sqrt(p*(1-p)/nn+z*z/(4*nn*nn)))/(1+z*z/nn))
}
func shadowQuote(s Sample,side string)(bid,ask,depth float64){
 if side=="UP"{return s.UpBid,s.UpAsk,s.UpAskDepth}
 return s.DownBid,s.DownAsk,s.DownAskDepth
}
func shadowCohort(s Sample,side string)cohort{
 _,ask,_:=shadowQuote(s,side)
 return cohort{s.WindowSeconds,side,shadowBucket(ask)}
}
func shadowCost(s Sample,side string)float64{
 _,ask,_:=shadowQuote(s,side)
 return fees.BufferedCost(1,ask+ShadowSlippage,ShadowExtraFeeBps,s.CapturedAt)
}
func shadowProfit(s Sample,side string)float64{
 payoff:=0.0;if s.Winner==side{payoff=1}
 return payoff-shadowCost(s,side)
}
func shadowEligible(s Sample,side string)bool{
 if side!="UP"&&side!="DOWN"{return false}
 bid,ask,depth:=shadowQuote(s,side)
 return ask>0&&ask<1&&bid>0&&bid<=ask&&ask-bid<=ShadowMaximumSpread&&depth>=1&&ask+ShadowSlippage<1
}
// EvaluateShadow uses strictly prior published labels to decide; a later
// published outcome is used only to score the decision, never to create it.
func EvaluateShadow(samples []Sample,now time.Time)ShadowReport{
 out:=ShadowReport{
  Status:"NO_VALID_RESEARCH_SAMPLES",AsOf:now,RequiredPriorPerCohort:ShadowPriorPerCohort,
  RequiredForwardPerWindow:ShadowForwardPerWindow,MinimumNetROIPct:100*ShadowMinimumNetROI,TradingAuthorized:false,
  Limitations:"Research-only. Same fixed-time snapshots and settle-to-expiry hypothetical costs, not real fills or the earlier exit strategy. Correlation, market regime shifts, selective missing data and non-independent Gamma labels remain unresolved. Never automatically authorize paper or live orders.",
  Windows:[]ShadowWindow{{Window:"5m",Status:"NOT_ENOUGH_RESOLVED_HISTORY"},{Window:"15m",Status:"NOT_ENOUGH_RESOLVED_HISTORY"}},
 }
 for i:=range out.Windows{for _,name:=range []string{"momentum_only","momentum_semivariance","market_price_baseline"}{
  out.Windows[i].Rules=append(out.Windows[i].Rules,ShadowRule{Name:name})
 }}
 if now.IsZero(){out.Status="INVALID_AS_OF_TIME";return out}
 ordered:=make([]Sample,0,len(samples));seen:=map[string]bool{}
 for _,s:=range samples{
  if s.Policy!=SamplingPolicy||!validSnapshot(s)||s.CapturedAt.After(now)||s.Slug==""||seen[s.Slug]{continue}
  seen[s.Slug]=true;ordered=append(ordered,s)
 }
 if len(ordered)==0{return out}
 out.Status="COLLECTING_PRIOR_EVIDENCE"
 sort.Slice(ordered,func(i,j int)bool{return ordered[i].CapturedAt.Before(ordered[j].CapturedAt)})
 published:=make([]Sample,0,len(ordered))
 for _,s:=range ordered{
  if (s.Winner=="UP"||s.Winner=="DOWN")&&!s.ResolvedAt.IsZero()&&!s.ResolvedAt.After(now){published=append(published,s)}
 }
 sort.Slice(published,func(i,j int)bool{return published[i].ResolvedAt.Before(published[j].ResolvedAt)})
 prior:=map[cohort]winCount{};pi:=0
 for _,s:=range ordered{
  w:=&out.Windows[0];if s.WindowSeconds==900{w=&out.Windows[1]}
  for pi<len(published)&&published[pi].ResolvedAt.Before(s.CapturedAt){
   known:=published[pi];pi++
   for _,side:=range []string{"UP","DOWN"}{
    key:=shadowCohort(known,side);v:=prior[key];v.total++
    if known.Winner==side{v.wins++};prior[key]=v
   }
  }
  for ri:=0;ri<3;ri++{
   rule:=&w.Rules[ri];side:=direction(s,ri)
   if side==""{rule.Abstained++;continue}
   hist:=prior[shadowCohort(s,side)]
   if hist.total>w.MaxPriorCohort{w.MaxPriorCohort=hist.total}
   if hist.total<ShadowPriorPerCohort{rule.BlockedInsufficientHistory++;continue}
   rule.ForwardObservations++
   if !shadowEligible(s,side){rule.BlockedSpreadOrDepth++;continue}
   cost:=shadowCost(s,side)
   if !finite(cost)||cost<=0||shadowWilson(hist.wins,hist.total)/cost-1<ShadowMinimumNetROI{
    rule.BlockedNoConservativeEdge++;continue
   }
   rule.EligibleCandidates++
   if s.Winner==""||s.ResolvedAt.After(now){continue}
   baseline:=direction(s,2)
   if baseline==""||!shadowEligible(s,baseline){continue} // no matched control; not scored
   rule.ResolvedCandidates++
   if side==s.Winner{rule.Winners++}
   rule.ShadowNetPerShareUSD+=shadowProfit(s,side)
   rule.BaselineNetPerShareUSD+=shadowProfit(s,baseline)
  }
  if s.Winner!=""&&!s.ResolvedAt.After(now){w.ResolvedMarketSamples++}
 }
 ready:=0
 for i:=range out.Windows{
  w:=&out.Windows[i];maxResolved:=0
  for j:=range w.Rules{
   x:=&w.Rules[j]
   x.ShadowNetPerShareUSD=math.Round(x.ShadowNetPerShareUSD*10000)/10000
   x.BaselineNetPerShareUSD=math.Round(x.BaselineNetPerShareUSD*10000)/10000
   x.ExcessNetPerShareUSD=math.Round((x.ShadowNetPerShareUSD-x.BaselineNetPerShareUSD)*10000)/10000
   if x.ResolvedCandidates>0{v:=math.Round(x.ExcessNetPerShareUSD/float64(x.ResolvedCandidates)*10000)/10000;x.MeanExcessUSD=&v}
   if x.ResolvedCandidates>maxResolved{maxResolved=x.ResolvedCandidates}
  }
  if maxResolved>=ShadowForwardPerWindow{w.Status="RESEARCH_ONLY_FORWARD_SAMPLE_AVAILABLE";ready++
  }else if w.MaxPriorCohort>=ShadowPriorPerCohort{w.Status="COLLECTING_PROSPECTIVE_SHADOW_RESULTS"}
 }
 if ready==2{out.Status="RESEARCH_ONLY_FORWARD_REVIEW_AVAILABLE"}
 return out
}
