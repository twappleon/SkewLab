package research100
import("time")
type AbstentionAudit struct{
 PublishedSameSnapshots int `json:"published_same_snapshots"`
 MomentumDecisions int `json:"momentum_decisions"`
 MomentumSkewDecisions int `json:"momentum_skew_decisions"`
 MomentumAbstentions int `json:"momentum_abstentions"`
 MomentumSkewAbstentions int `json:"momentum_skew_abstentions"`
 Status string `json:"status"`
}
func AssessAbstentions(d Dataset,asOf time.Time)AbstentionAudit{
 out:=AbstentionAudit{Status:"ABSTENTIONS_VISIBLE_NO_EQUAL_TRADE_COUNT_ASSUMPTION"}
 for _,o:=range d.Items{
  if o.Winner==""||o.CapturedAt.After(asOf)||o.ResolvedAt.After(asOf){continue}
  out.PublishedSameSnapshots++
  momentum:=(o.Side=="UP"&&o.Momentum>=.02)||(o.Side=="DOWN"&&o.Momentum<=-.02)
  skew:=(o.Side=="UP"&&o.Skew>=.2)||(o.Side=="DOWN"&&o.Skew<=-.2)
  if momentum{out.MomentumDecisions++}else{out.MomentumAbstentions++}
  if momentum&&skew{out.MomentumSkewDecisions++}else{out.MomentumSkewAbstentions++}
 }
 return out
}
