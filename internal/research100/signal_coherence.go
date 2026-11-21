package research100
import("time")
type SignalCoherence struct{
 Published int `json:"published"`
 MomentumAligned int `json:"momentum_aligned"`
 BothAligned int `json:"both_aligned"`
 BothAlignedWins int `json:"both_aligned_wins"`
 Status string `json:"status"`
}
func AssessSignalCoherence(d Dataset,asOf time.Time)SignalCoherence{
 out:=SignalCoherence{Status:"SELECTED_SIDE_ALIGNMENT_NOT_CAUSAL_PROOF"}
 for _,o:=range d.Items{
  if o.Winner==""||o.CapturedAt.After(asOf)||o.ResolvedAt.After(asOf){continue}
  out.Published++
  aligned:=(o.Side=="UP"&&o.Momentum>=.02)||(o.Side=="DOWN"&&o.Momentum<=-.02)
  skewOK:=(o.Side=="UP"&&o.Skew>=.2)||(o.Side=="DOWN"&&o.Skew<=-.2)
  if aligned{out.MomentumAligned++}
  if aligned&&skewOK{out.BothAligned++;if o.Winner==o.Side{out.BothAlignedWins++}}
 }
 return out
}
