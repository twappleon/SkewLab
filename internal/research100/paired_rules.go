package research100
import("time";"github.com/twappleon/SkewLab/internal/experiment")
type MatchedRuleScore struct{
 Name string `json:"name"`
 PublishedSameSnapshots int `json:"published_same_snapshots"`
 Decisions int `json:"decisions"`
 Abstentions int `json:"abstentions"`
 Wins int `json:"wins"`
 HypotheticalNetUSD float64 `json:"hypothetical_net_usd"`
}
type MatchedThreeRules struct{
 Rules []MatchedRuleScore `json:"rules"`
 Status string `json:"status"`
 TradingAuthorized bool `json:"trading_authorized"`
}
func ComparePairedRules(samples []experiment.Sample,asOf time.Time)MatchedThreeRules{
 out:=MatchedThreeRules{Status:"PAIRED_PREDECLARED_OBSERVATIONS_NOT_TRADE_APPROVAL",TradingAuthorized:false,Rules:[]MatchedRuleScore{{Name:"market_price_midpoint"},{Name:"momentum_only"},{Name:"momentum_semivariance"}}}
 for _,s:=range samples{
  if !ValidatePairedSnapshot(s,asOf).Valid||s.Winner==""||s.ResolvedAt.After(asOf){continue}
  sides:=[]string{MarketBaselineSide(s,asOf).Side,MomentumSide(s,asOf).Side,MomentumSkewSide(s,asOf).Side}
  for i,side:=range sides{
   r:=&out.Rules[i];r.PublishedSameSnapshots++
   if side=="WAIT"{r.Abstentions++;continue}
   p:=ScorePairedPayoff(s,side,asOf)
   if !p.Scored{r.Abstentions++;continue}
   r.Decisions++;r.HypotheticalNetUSD+=p.NetOneShareUSD
   if p.Correct{r.Wins++}
  }
 }
 return out
}
