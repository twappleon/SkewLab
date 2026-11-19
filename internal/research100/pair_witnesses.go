package research100
import("time")
type ReplayPair struct{
 Entry Observation `json:"entry"`
 Exit ExitWitness `json:"exit"`
}
type ReplayPairing struct{
 ObservedEntries int `json:"observed_entries"`
 ValidPairs int `json:"valid_pairs"`
 MissingWitnesses int `json:"missing_witnesses"`
 RejectedWitnesses int `json:"rejected_witnesses"`
 DuplicatedWitnesses int `json:"duplicated_witnesses"`
 Pairs []ReplayPair `json:"-"`
 Status string `json:"status"`
}
func PairExitEvidence(d Dataset,witnesses []ExitWitness,asOf time.Time)ReplayPairing{
 out:=ReplayPairing{Status:"OBSERVED_FIXED_EXIT_WITNESS_ONLY",Pairs:[]ReplayPair{}}
 indexed:=map[string][]ExitWitness{}
 for _,w:=range witnesses{indexed[w.ID]=append(indexed[w.ID],w)}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  out.ObservedEntries++
  candidates:=indexed[o.ID]
  if len(candidates)==0{out.MissingWitnesses++;continue}
  if len(candidates)>1{out.DuplicatedWitnesses++;continue}
  if !ValidateExitWitness(o,candidates[0],asOf).Valid{out.RejectedWitnesses++;continue}
  out.ValidPairs++;out.Pairs=append(out.Pairs,ReplayPair{Entry:o,Exit:candidates[0]})
 }
 return out
}
