package research100
import("math";"time")
type ReliabilityBin struct{
 Index int `json:"index"`
 Count int `json:"count"`
 Wins int `json:"wins"`
 MeanQuotedProbability *float64 `json:"mean_quoted_probability,omitempty"`
 ObservedWinFraction *float64 `json:"observed_win_fraction,omitempty"`
}
func AssessReliabilityBins(d Dataset,asOf time.Time)[]ReliabilityBin{
 out:=make([]ReliabilityBin,5);totals:=[5]float64{}
 for i:=range out{out[i].Index=i}
 for _,o:=range d.Items{
  if o.Winner==""||o.ResolvedAt.After(asOf)||o.CapturedAt.After(asOf){continue}
  p:=(o.Bid+o.Ask)/2
  if !finite(p)||p<=0||p>=1{continue}
  b:=int(math.Floor(p*5));if b>4{b=4}
  out[b].Count++;totals[b]+=p
  if o.Winner==o.Side{out[b].Wins++}
 }
 for i:=range out{
  if out[i].Count==0{continue}
  p:=totals[i]/float64(out[i].Count);f:=float64(out[i].Wins)/float64(out[i].Count)
  out[i].MeanQuotedProbability=&p;out[i].ObservedWinFraction=&f
 }
 return out
}
