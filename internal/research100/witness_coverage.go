package research100
import("time")
type WitnessCoverage struct{
 Entries int `json:"entries"`
 Matched int `json:"matched"`
 Missing int `json:"missing"`
 Rejected int `json:"rejected"`
 Duplicate int `json:"duplicate"`
 Fraction *float64 `json:"fraction,omitempty"`
 Status string `json:"status"`
}
func AssessWitnessCoverage(d Dataset,witnesses []ExitWitness,asOf time.Time)WitnessCoverage{
 pairs:=PairExitEvidence(d,witnesses,asOf)
 out:=WitnessCoverage{Entries:pairs.ObservedEntries,Matched:pairs.ValidPairs,Missing:pairs.MissingWitnesses,Rejected:pairs.RejectedWitnesses,Duplicate:pairs.DuplicatedWitnesses,Status:"UNKNOWN_NO_ENTRIES"}
 if out.Entries>0{
  v:=float64(out.Matched)/float64(out.Entries);out.Fraction=&v
  out.Status="PARTIAL_OR_MISSING_EXIT_EVIDENCE"
  if out.Matched==out.Entries{out.Status="SINGLE_QUOTE_WITNESS_COVERAGE_NOT_FILL_PROOF"}
 }
 return out
}
